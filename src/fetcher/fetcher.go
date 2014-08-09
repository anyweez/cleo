package main

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"libcleo"
	"log"
	"net/http"
	"os"
	gamelog "proto"
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

type JSONResponse struct {
	Games      []JSONGameResponse `json:"games"`
	SummonerId uint32
}

type JSONGameResponse struct {
	FellowPlayers []JSONPlayerResponse
	Stats         JSONGameStatsResponse

	GameId     uint64
	CreateDate uint64

	TeamId     uint32
	ChampionId uint32

	GameMode    string
	GameType    string
	GameSubType string
}

type JSONGameStatsResponse struct {
	Win bool
}

type JSONPlayerResponse struct {
	SummonerId uint32
	TeamId     uint32
	ChampionId uint32
}

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
		time.Sleep(1200 * time.Millisecond)

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
	url := "https://na.api.pvp.net/api/lol/na/v1.3/game/by-summoner/%d/recent?api_key=%s"

	resp, err := http.Get(fmt.Sprintf(url, summoner, *API_KEY))
	if err != nil {
		log.Println("Error retrieving data:", err)
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		json_response := JSONResponse{}
		json.Unmarshal(body, &json_response)

		// Write all games into permanent storage.
		for _, game := range convert(&json_response) {
			// Add all of the players to the candidate manager. It
			// takes care of removing duplicates automatically.
			for _, team := range game.Teams {
				for _, player := range team.Players {
					cm.Add(*player.Player.SummonerId)
				}
			}

			if STORE_RESPONSES {
				// Check to see if the game already exists. If so, don't do anything.
				record_count, _ := collection.Find(bson.M{"gameid": *game.GameId}).Count()

				if record_count == 0 {
					// Encode and store in the database.
					encoded_gamedata, _ := gproto.Marshal(&game)
					record := libcleo.RecordContainer{encoded_gamedata, *game.GameId, *game.Timestamp}

					collection.Insert(record)
				}
			}
		} // end for
	}
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

		record.Timestamp = gproto.Uint64(game.CreateDate)
		record.GameId = gproto.Uint64(game.GameId)

		team1 := gamelog.Team{}
		team2 := gamelog.Team{}

		// Add the target player and set the outcome.
		plyr := gamelog.Player{}
		plyr.SummonerId = gproto.Uint32(response.SummonerId)

		pstats := gamelog.PlayerStats{}
		pstats.Champion = libcleo.Rid2Cleo(game.ChampionId).Enum()
		pstats.Player = &plyr

		if game.TeamId == 100 {
			team1.Players = append(team1.Players, &pstats)

			team1.Victory = gproto.Bool(game.Stats.Win)
			team2.Victory = gproto.Bool(!game.Stats.Win)
		} else if game.TeamId == 200 {
			team2.Players = append(team2.Players, &pstats)

			team1.Victory = gproto.Bool(!game.Stats.Win)
			team2.Victory = gproto.Bool(game.Stats.Win)
		} else {
			log.Println("Unknown team ID found on game", game.GameId)
		}

		// Add all fellow players.
		for _, player := range game.FellowPlayers {
			plyr := gamelog.Player{}
			pstats := gamelog.PlayerStats{}

			plyr.SummonerId = gproto.Uint32(player.SummonerId)
			pstats.Champion = libcleo.Rid2Cleo(player.ChampionId).Enum()

			if player.TeamId == 100 {
				pstats.Player = &plyr
				team1.Players = append(team1.Players, &pstats)
			} else if player.TeamId == 200 {
				pstats.Player = &plyr
				team2.Players = append(team2.Players, &pstats)
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
