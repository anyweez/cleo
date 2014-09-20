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
	"sync"
	"time"
)

var SUMMONER_FILE = flag.String("summoners", "", "The filename containing a list of summoner ID's.")
var TARGET_DATE = flag.String("date", "0000-00-00", "The date to join in YYYY-MM-DD form.")

var GR_GROUP sync.WaitGroup
var MAX_CONCURRENT_GR = flag.Int("max_concurrent", 100, "The number of simultaneous summoners that will be examined.")

/**
 * Goroutine that generates a report for a single summoner ID. It reads
 * through all game records and retains those that were played by the
 * target summoner ID. It then condenses them into a single PlayerSnapshot
 * and saves it to MongoDB.
 */
func handle_summoner(sid uint32, input chan *gamelog.GameRecord, done chan bool) {
	games := make([]*gamelog.GameRecord, 0, 10)
	game_ids := make([]uint64, 0, 10)
	
	// Keep reading from the channel until nil comes through, then we're
	// done receiving info. If the summoner this goroutine is responsible
	// for played in the game, keep it. Otherwise forget about it.
	
	retriever := snapshot.Retriever{}
	retriever.Init()
	
	iter := retriever.GetGamesIter(*TARGET_DATE)
	result := gamelog.GameRecord{}

	for iter.Next(&result) {
		log.Println(fmt.Sprintf("Found game: %d", result.GameId))
		keeper := false
		for _, team := range result.Teams {
			for _, player := range team.Players {
				if player.Player.SummonerId == sid {
					keeper = true
					log.Println(fmt.Sprintf("Game found: %d", result.GameId))
				}
			}
		}

		if keeper {
			games = append(games, &result)
			game_ids = append(game_ids, result.GameId)
		}
	}

	// Now all games have been processed. We need to save the set of
	// games to a PlayerSnapshot for today.
	snap := snapshot.PlayerSnapshot{}

	snap.CreationTimestamp = (uint64)(time.Now().Unix())
	snap.SummonerId = (uint32)(sid)
	snap.GamesList = game_ids

	snap.Stats = make([]snapshot.PlayerStat, 0, 10)

	// Update each snapshot with new computations.
	for _, comp := range snapshot.Computations {
		sv := snapshot.PlayerStat{}
		sv.Name, sv.Absolute, sv.Normalized = comp(&snap, games)
		
		snap.Stats = append(snap.Stats, sv)
	}

	retriever.SaveSnapshot(sid, "daily", *TARGET_DATE, &snap)
	log.Println(fmt.Sprintf("Saved daily snapshot for summoner #%d on %s", sid, *TARGET_DATE))
	
	done <- true
	GR_GROUP.Done()
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
	running_queue := make(chan bool, *MAX_CONCURRENT_GR)
	
	// Load up the initial queue.
	for i := 0; i < *MAX_CONCURRENT_GR; i++ {
		running_queue <- true
	}
	
	// Check to make sure that a file was provided.
	if len(*SUMMONER_FILE) == 0 {
		log.Fatal("You must specify a list of summoner ID's with the --summoners flag")
	}

	// Basic format test for the target string. Making this a regex would be better.
	if len(strings.Split(*TARGET_DATE, "-")) != 3 {
		log.Fatal("Provided date must be in YYYY-MM-DD format")
	}

	/* Read in the list of summoner ID's from a file provided by the user. */
	sids := snapshot.ReadSummonerIds(*SUMMONER_FILE)
	sid_chan := make([]chan *gamelog.GameRecord, len(sids))
	log.Println(fmt.Sprintf("Read %d summoners from champion list.", len(sids)))

	/* Create a bunch of goroutines, one per summoner, that can be used
	 * to filter records. */
	for i, sid := range sids {
		<- running_queue
		sid_chan[i] = make(chan *gamelog.GameRecord)

		GR_GROUP.Add(1)		
		go handle_summoner(sid, sid_chan[i], running_queue)
	}

	// Wait for all goroutines to finish up before exiting.
	GR_GROUP.Wait()
}
