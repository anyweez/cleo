package main

import (
	data "datamodel"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"lolutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Constants
var API_KEY = flag.String("apikey", "", "Riot API key")
var CHAMPION_LIST = flag.String("summoners", "champions", "List of summoner ID's")

const STORE_RESPONSES = true

var logger, _ = syslog.New(syslog.LOG_INFO, "fetcher")

func main() {
	// Flag setup
	flag.Parse()

	if *API_KEY == "" {
		log.Fatal("You must provide an API key using the -apikey flag.")
	}

	fmt.Println("Initializing...")
	retriever := data.LoLRetriever{}

	cm := lolutil.LoadCandidates(retriever, *CHAMPION_LIST)

	fmt.Println(fmt.Sprintf("Loaded %d summoners...let's do this!", cm.Count()))

	counter := 0
	// Forever: pull an summoner ID from user_queue, toss it in retrieval_inputs
	// and add it back to user_queue.
	for {
		// Wait for 1.20 seconds to account for the rate limiting that
		// Riot requires.
		time.Sleep(1100 * time.Millisecond)

		// Push the player to the retrieval queue.
		go retrieve(cm.Next(), &retriever)
		counter += 1
	}
}

// Retrievers hang out until presented with a player to look up. Once they
// receive a record, they issue a request to the API, convert the respones
// into a series of GameRecord's, and insert that data into permanent
// storage.
//
// Note that all rate limiting is handled directly by the channel, meaning
// that everything in this goroutine can execute as quickly as possible.
func retrieve(summoner uint32, retriever *data.LoLRetriever) {
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
			// Store everything per game
			if STORE_RESPONSES {
				// Check to see if the game already exists. If so, don't do anything.
				record, exists := retriever.GetGame(game.GameId)

				// Insert a new record.
				if !exists {
					game.MergeCount = 1
					// Encode and store in the database.
					retriever.StoreGame(&game)
				} else {
					// Otherwise merge with a pre-existing record.
					// TODO: add a lock to make sure the response we retrieve doesn't go stale before we
					//   update it.
					
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
											retriever.StoreGame(&game)
										}
									}
								}
							}
						}
					}

					if found_target == 0 {
						log.Println("Found matching game ID's with non-matching summoner ID's.")
					}
				} // end else
			} // end STORE_RESPONSES block
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
func convert(response *JSONResponse) []data.GameRecord {
	games := make([]data.GameRecord, 0, 10)

	for _, game := range response.Games {
		// Only keep games that are matched 5v5 (no bot games, etc).
		if (game.GameMode != "CLASSIC") || (game.GameType != "MATCHED_GAME") || (strings.Contains(game.GameSubType, "5x5")) {
			continue
		}

		record := data.GameRecord{}

		record.Timestamp = game.CreateDate
		record.QuickDate = timestampToQuickdate(record.Timestamp)
		record.GameId = game.GameId

		team1 := data.Team{}
		team2 := data.Team{}

		// Add the target player and set the outcome.
		plyr := data.PlayerType{}
		plyr.SummonerId = response.SummonerId

		pstats := data.PlayerStats{}
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
			plyr := data.PlayerType{}
			fellow_stats := data.PlayerStats{}

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
