package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gamelog"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

// Constants
var API_KEY = flag.String("apikey", "", "Riot API key")

const STORE_RESPONSES = true

// Flags
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

//////////////////////////////////////////////////////////////////
//// CandidateManager keeps track of a list of Players that can be fetched
//// and ensures that they're all unique in the queue.
//////////////////////////////////////////////////////////////////
type CandidateManager struct {
	Queue        chan uint32
	CandidateMap map[uint32]bool

	count uint32
}

func (cm *CandidateManager) Add(player uint32) {
	_, exists := cm.CandidateMap[player]
	if !exists {
		cm.CandidateMap[player] = true
		cm.count += 1

		cm.Queue <- player
	}
}

func (cm *CandidateManager) Next() uint32 {
	player := <-cm.Queue
	// Cycle the candidate back into the queue.
	cm.Queue <- player

	return player
}

func (cm *CandidateManager) Count() uint32 {
	return cm.count
}

// This function reads in a list of champions from a local file to start
// as the seeding set. The fetcher will automatically include new champions
// it discovers on its journey as well.
func read_summoner_ids(filename string) []uint32 {
	// Read the specified file.
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Cannot find champions file.")
	}
	defer file.Close()

	lines := make([]uint32, 0, 10000)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		value, _ := strconv.ParseUint(scanner.Text(), 10, 32)
		lines = append(lines, uint32(value))
	}

	// Return a set of summoner ID's.
	return lines
}

func load_starting_ids(cm *CandidateManager) {
	// Load in a file full of summoner ID's.
	summoner_ids := read_summoner_ids("champions")
	for _, sid := range summoner_ids {
		cm.Add(sid)
	}
}

func main() {
	// Flag setup
	flag.Parse()
	if *cpuprofile != "" {
		fmt.Println("Starting CPU profiling...")
		cf, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cf)
		defer pprof.StopCPUProfile()
	}

	if *API_KEY == "" {
		log.Fatal("You must provide an API key using the -apikey flag.")
	}

	fmt.Println("Initializing...")

	cm := CandidateManager{}
	// TODO: can the queue length be dynamic? static is a problem
	// because this will eventually fill up.
	cm.Queue = make(chan uint32, 50000000)
	cm.CandidateMap = make(map[uint32]bool)

	// Connect to MongoDB instance.
	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("games")
	defer session.Close()

	// Kick off some retrievers that will pull from the retrieval queue.
	load_starting_ids(&cm)
	fmt.Println(fmt.Sprintf("Loaded %d summoners...let's do this!", cm.Count()))

	counter := 0
	// Forever: pull an summoner ID from user_queue, toss it in retrieval_inputs
	// and add it back to user_queue.
	for {
		// Wait for 1.20 seconds to account for the rate limiting that
		// Riot requires.
		time.Sleep(1100 * time.Millisecond)

		// Push the player to the retrieval queue.
		go retrieve(cm.Next(), games_collection, &cm)
		counter += 1

		fmt.Print(fmt.Sprintf("Summoner queue size: %d [%.1f%% to next export]\r", cm.Count(), float32((counter%1000)/1000)))

		// Run for approx 2 hrs then dump data.
		if counter == 1000 && (*cpuprofile != "" || *memprofile != "") {
			pprof.StopCPUProfile()

			if *memprofile != "" {
				mf, err := os.Create(*memprofile)
				if err != nil {
					log.Fatal(err)
				}
				pprof.WriteHeapProfile(mf)
				mf.Close()
			}

			fmt.Println("Profiling complete")
		}

		// Every thousand requests save the new summoner list.
		if (counter%1000 == 0) && STORE_RESPONSES {
			write_candidates(&cm)
		}
	}
}

// Retrievers hang out until presented with a player to look up. Once they
// receive a record, they issue a request to the API, convert the respones
// into a series of GameRecord's, and insert that data into permanent
// storage.
//
// Note that all rate limiting is handled directly by the channel, meaning
// that everything in this goroutine can execute as quickly as possible.
func retrieve(summoner uint32, collection *mgo.Collection, cm *CandidateManager) {
	// Retrieve game data.
	url := "https://na.api.pvp.net/api/lol/na/v1.3/game/by-summoner/%d/recent?api_key=%s"
	resp, err := http.Get(fmt.Sprintf(url, summoner, *API_KEY))
	json_response := JSONResponse{}

	if err != nil {
		log.Println("Error retrieving data:", err)
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		json.Unmarshal(body, &json_response)

		// Write all games into permanent storage.
		for _, game := range convert(&json_response) {
			// Add all of the players to the candidate manager. It
			// takes care of removing duplicates automatically.
			for _, team := range game.Teams {
				for _, player := range team.Players {
					cm.Add(player.Player.SummonerId)
				}
			}

			// Store everything per game
			if STORE_RESPONSES {
				// Check to see if the game already exists. If so, don't do anything.
				record_count, _ := collection.Find(bson.M{"_id": game.GameId}).Count()

				// Insert a new record.
				if record_count == 0 {
					game.MergeCount = 1
					// Encode and store in the database.
					collection.Insert(game)
				} else {
					// Otherwise merge with a pre-existing record.
					// TODO: add a lock to make sure the response we retrieve doesn't go stale before we
					//   update it.
					record := gamelog.GameRecord{}
					collection.Find(bson.M{"_id": game.GameId}).One(&record)

					// A ridiculously nested loop that compares players in the stored game record with
					// players from the new record and merges them if it finds overlap (should be exactly
					// one player). The found_player variable is used to confirm that only one player
					// is updated.
					found_target := 0
					for i, recorded_team := range record.Teams {
						for j, recorded_player := range recorded_team.Players {
							for k, incoming_team := range game.Teams {
								for m, incoming_player := range incoming_team.Players {
									if recorded_player.Player.SummonerId == incoming_player.Player.SummonerId {
										// Check to make sure the player actually has data to add.
										if incoming_player.IsSet && !recorded_player.IsSet {
											record.Teams[i].Players[j] = game.Teams[k].Players[m]

											found_target += 1

											// Increment the counter for the number of players that have
											// been merged into this game record.
											record.MergeCount += 1
											collection.Update(bson.M{"_id": game.GameId}, record)
										}
									}
								}
							}
						}
					}

					if found_target == 0 {
						log.Println("Found matching game ID's with non-matching summoner ID's.")
					}
				}
			}
		} // end for
	}
}

func timestampToQuickdate(ts uint64) uint32 {
	num, _ := strconv.Atoi( time.Unix( (int64)(ts / 1000), 0).Format("20060102") )
	
	return (uint32)(num)
}

// Adaptor to convert the JSON format to GameLog format.
//
// This function converts Riot's JSON format into GameLog entries which
// are used for everything internally. Note that certain fields are
// dropped here at the moment.
func convert(response *JSONResponse) []gamelog.GameRecord {
	games := make([]gamelog.GameRecord, 0, 10)

	for _, game := range response.Games {
		// Only keep games that are matched 5v5 (no bot games, etc).
		if (game.GameMode != "CLASSIC") || (game.GameType != "MATCHED_GAME") || (strings.Contains(game.GameSubType, "5x5")) {
			continue
		}

		record := gamelog.GameRecord{}

		record.Timestamp = game.CreateDate
		record.QuickDate = timestampToQuickdate(record.Timestamp)
		record.GameId = game.GameId

		team1 := gamelog.Team{}
		team2 := gamelog.Team{}

		// Add the target player and set the outcome.
		plyr := gamelog.PlayerType{}
		plyr.SummonerId = response.SummonerId

		pstats := gamelog.PlayerStats{}
		pstats.Champion = game.ChampionId
		pstats.Player = &plyr

		// Populate stats fields.
		pstats.Kills = game.Stats.ChampionsKilled
		pstats.Deaths = game.Stats.NumDeaths
		pstats.Assists = game.Stats.Assists
		pstats.GoldEarned = game.Stats.GoldEarned
		pstats.Minions = game.Stats.MinionsKilled
		pstats.IsSet = true

		if game.TeamId == 100 {
			team1.Players = append(team1.Players, &pstats)

			team1.Victory = game.Stats.Win
			team2.Victory = !game.Stats.Win
		} else if game.TeamId == 200 {
			team2.Players = append(team2.Players, &pstats)

			team1.Victory = !game.Stats.Win
			team2.Victory = game.Stats.Win
		} else {
			log.Println("Unknown team ID found on game", game.GameId)
		}

		// Add all fellow players. Note that we only get stats for the player that we're
		// querying for.
		for _, player := range game.FellowPlayers {
			plyr := gamelog.PlayerType{}
			fellow_stats := gamelog.PlayerStats{}

			plyr.SummonerId = player.SummonerId
			fellow_stats.Champion = player.ChampionId
			fellow_stats.IsSet = false

			if player.TeamId == 100 {
				fellow_stats.Player = &plyr
				team1.Players = append(team1.Players, &fellow_stats)
			} else if player.TeamId == 200 {
				fellow_stats.Player = &plyr
				team2.Players = append(team2.Players, &fellow_stats)
			} else {
				log.Println("Unknown team ID found on game", game.GameId)
			}
		}
		// Add teams to the game record.
		record.Teams = append(record.Teams, &team1, &team2)

		games = append(games, record)
	}

	return games
}

// Write out a list of all of the candidates we've found so far. This is
// used to seed the list next time we load the fetcher.
func write_candidates(cm *CandidateManager) {
	t := time.Now()
	filename := fmt.Sprintf("gamelogs/%s.summ", t.Local().Format("2006-01-02:15.04"))
	f, _ := os.Create(filename)
	defer f.Close()

	for k, _ := range cm.CandidateMap {
		io.WriteString(f, strconv.FormatUint(uint64(k), 10)+"\n")
	}
}
