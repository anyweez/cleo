package main

/**
 * This process creates or updates a record for each summoner ID
 * in the list provided as an input. Each record includes a "daily"
 * key that contains a bunch of records with summary stats for a
 * given day.
 *
 * ./join-summoners --date=2014-08-07
 */

import (
	data "datamodel"
	"flag"
	"fmt"
	"log"
	"lolutil"
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
func handle_summoner(sid uint32, input chan *data.GameRecord, done chan bool) {
	games := make([]*data.GameRecord, 0, 10)
	game_ids := make([]uint64, 0, 10)
	
	// Keep reading from the channel until nil comes through, then we're
	// done receiving info. If the summoner this goroutine is responsible
	// for played in the game, keep it. Otherwise forget about it.
	retriever := data.LoLRetriever{}
	games_iter := retriever.GetQuickdateGamesIter(*TARGET_DATE)

	for games_iter.HasNext() {
		result := games_iter.Next()

		// Skip this record if the gameid is zero.
		if result.GameId == 0 {
			continue
		}
		
		keeper := false
		for _, team := range result.Teams {
			for _, player := range team.Players {
				if player.Player.SummonerId == sid {
					keeper = true
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
	snap := data.PlayerSnapshot{}

	snap.CreationTimestamp = (uint64)(time.Now().Unix())
	snap.SummonerId = (uint32)(sid)
	snap.GamesList = game_ids

	snap.Stats = make(map[string]data.Metric)

	// Update each snapshot with new computations.
	for _, comp := range snapshot.Computations {
		name, metric := comp(snap, games)
		snap.Stats[name] = metric
	}

	// Fetch the summoner that this applies to.
	summoner, exists := retriever.GetSummoner(sid)

	// If the summoner doesn't exist, create it.
	if !exists {
		log.Println(fmt.Sprintf("Notice: Couldn't find summoner #%d; creating new instance.", sid)		)
		summoner = data.SummonerRecord{}
		summoner.SummonerId = sid
	}
	
	// Append the snapshot.
	if summoner.Daily == nil {
		summoner.Daily = make(map[string]*data.PlayerSnapshot)
	}

	summoner.Daily[*TARGET_DATE] = &snap
	// Store the revised summoner.
	retriever.StoreSummoner(&summoner)
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

	// Basic format test for the target string. Making this a regex would be better.
	if len(strings.Split(*TARGET_DATE, "-")) != 3 {
		log.Fatal("Provided date must be in YYYY-MM-DD format")
	}

	/* Read in the list of summoner ID's from a file provided by the user. */
	retriever := data.LoLRetriever{}
	cm := lolutil.LoadCandidates(retriever, *SUMMONER_FILE)
	sid_chan := make([]chan *data.GameRecord, cm.Count())
	log.Println(fmt.Sprintf("Read %d summoners from champion list.", cm.Count()))

	/* Create a bunch of goroutines, one per summoner, that can be used
	 * to filter records. */
	for i := 0; i < (int)(cm.Count()); i++ {
		<- running_queue
		sid_chan[i] = make(chan *data.GameRecord)

		GR_GROUP.Add(1)		
		go handle_summoner(cm.Next(), sid_chan[i], running_queue)
	}

	// Wait for all goroutines to finish up before exiting.
	GR_GROUP.Wait()
}
