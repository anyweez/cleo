package main

/**
 * This process creates or updates a record for each summoner ID
 * in the list provided as an input. Each record includes a "daily"
 * key that contains a bunch of records with summary stats for a
 * given day.
 *
 * ./join-summoners --date=2014-08-07 --summoners=data/summoners/0001.list
 */

import (
	"flag"
	"fmt"
	"gamelog"
	"log"
	"snapshot"
	"strings"
	"time"
)

var SUMMONER_FILE = flag.String("summoners", "", "The filename containing a list of summoner ID's.")
var TARGET_DATE = flag.String("date", "0000-00-00", "The date to join in YYYY-MM-DD form.")
// TODO: Convert this to a channel
var SUMMONER_GR_RUNNING = 0

/**
 * Goroutine that generates a report for a single summoner ID. It reads
 * through all game records and retains those that were played by the
 * target summoner ID. It then condenses them into a single PlayerSnapshot
 * and saves it to MongoDB.
 */
func handle_summoner(sid uint32, input chan *gamelog.GameRecord, done chan bool) {
	games := make([]*gamelog.GameRecord, 0, 10)
	game_ids := make([]uint64, 0, 10)
	SUMMONER_GR_RUNNING += 1
	
	// Keep reading from the channel until nil comes through, then we're
	// done receiving info. If the summoner this goroutine is responsible
	// for played in the game, keep it. Otherwise forget about it.
	gr := <- input
	for gr != nil {
		keeper := false
		for _, team := range gr.Teams {
			for _, player := range team.Players {
				if player.Player.SummonerId == sid {
					keeper = true
				}
			}
		}

		if keeper {
			games = append(games, gr)
			game_ids = append(game_ids, gr.GameId)
		}

		gr = <-input
	}

	// Now all games have been processed. We need to save the set of
	// games to a PlayerSnapshot for today.
	snap := snapshot.PlayerSnapshot{}

	ts_start, ts_end := snapshot.ConvertTimestamp(*TARGET_DATE)
	snap.StartTimestamp = ts_start
	snap.EndTimestamp = ts_end
	snap.CreationTimestamp = (uint64)(time.Now().Unix())
	snap.SummonerId = (uint32)(sid)
	snap.GamesList = game_ids

	// TODO: populate snap.CreationTimestamp
	snap.Stats = make([]snapshot.PlayerStat, 0, 10)

	// Update each snapshot with new computations.
	for _, comp := range snapshot.Computations {
		sv := snapshot.PlayerStat{}
		sv.Name, sv.Absolute, sv.Normalized = comp(&snap, games)
		
		snap.Stats = append(snap.Stats, sv)
	}

	// Commit to datastore
	retriever := snapshot.Retriever{}
	retriever.Init()

	retriever.SaveSnapshot(sid, "daily", *TARGET_DATE, &snap)
	
	SUMMONER_GR_RUNNING -= 1
	done <- true
}

/**
 * The main function reads in all of the summoner ID's that this process
 * will be responsible for and forks off a separate goroutine for each
 * of them. It then reads all relevant games (see get_games) and passes
 * pointers to each of the goroutines for filtering.
 *
 * Finally, it waits for all goroutines to terminate before terminating
 * itself.
 */
func main() {
	flag.Parse()

	// Create a retriever for I/O
	log.Println("Initializing retriever...")
	retriever := snapshot.Retriever{}
	retriever.Init()

	// Check to make sure that a file was provided.
	if len(*SUMMONER_FILE) == 0 {
		log.Fatal("You must specify a list of summoner ID's with the --summoners flag")
	}

	// Basic format test for the target string. Making this a regex would be better.
	if len(strings.Split(*TARGET_DATE, "-")) != 3 {
		log.Fatal("Provided date must be in YYYY-MM-DD format")
	}

	/* Read in the list of summoner ID's from a file provided by the user. */
	sids := snapshot.ReadSummonerIds(*SUMMONER_FILE)[:3]
	sid_chan := make([]chan *gamelog.GameRecord, len(sids))
	log.Println(fmt.Sprintf("Read %d summoners from champion list.", len(sids)))

	running := make(chan bool)

	/* Create a bunch of goroutines, one per summoner, that can be used
	 * to filter records. */
	for i, sid := range sids {
		sid_chan[i] = make(chan *gamelog.GameRecord)
		go handle_summoner(sid, sid_chan[i], running)
	}

	/* Retrieve all events. */
//	log.Println(fmt.Sprintf("Retrieving games from %s...", *TARGET_DATE))
//	games := retriever.GetGames(*TARGET_DATE)
//	log.Println(fmt.Sprintf("Retrieved %d games.", len(games)))

	/* Pass each event to a goroutine that handles each summoner ID. */
/*
	for i, _ := range sids {
		log.Println("Creating summoner GR #", i)
		for _, game := range games {
			sid_chan[i] <- &game
		}
		
		// Sending a nil means that the goroutine doesn't need to wait
		// for more data.
		sid_chan[i] <- nil		
		
		for SUMMONER_GR_RUNNING > 10 {
			
		}
	}
*/
	for i := 0; i < len(sids); i++ {
		<-running
	}
}
