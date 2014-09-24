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
	gproto "code.google.com/p/goprotobuf/proto"
	data "datamodel"
	"flag"
	"fmt"
	beanstalk "github.com/iwanbk/gobeanstalk"
	"log"
	"proto"
	"snapshot"
	"sort"
	"sync"
	"time"
)

var GR_GROUP sync.WaitGroup

/**
 * Goroutine that generates a report for a single summoner ID. It reads
 * through all game records and retains those that were played by the
 * target summoner ID. It then condenses them into a single PlayerSnapshot
 * and saves it to MongoDB.
 */
func handle_summoner(request proto.JoinRequest, sid uint32) {
	games := make([]*data.GameRecord, 0, 10)
	game_ids := make([]uint64, 0, 10)

	// Keep reading from the channel until nil comes through, then we're
	// done receiving info. If the summoner this goroutine is responsible
	// for played in the game, keep it. Otherwise forget about it.
	retriever := data.LoLRetriever{}

	for _, qd := range request.Quickdates {
		games_iter := retriever.GetQuickdateGamesIter(qd)

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
		log.Println(fmt.Sprintf("Notice: Couldn't find summoner #%d; creating new instance.", sid))
		summoner = data.SummonerRecord{}
		summoner.SummonerId = sid
	}

	// Append the snapshot.
	if summoner.Daily == nil {
		summoner.Daily = make(map[string]*data.PlayerSnapshot)
	}

	sort.Strings(request.Quickdates)
	quickdate_label := request.Quickdates[0]
	// Store the snapshot in the right bucket, depending on the label name.
	if *request.Label == "daily" {
		summoner.Daily[quickdate_label] = &snap
	} else if *request.Label == "weekly" {
		summoner.Weekly[quickdate_label] = &snap
	} else if *request.Label == "monthly" {
		summoner.Monthly[quickdate_label] = &snap
	} else {
		log.Fatal("Unknown time label:", request.Label)
	}

	// Store the revised summoner.
	retriever.StoreSummoner(&summoner)

	log.Println(fmt.Sprintf("Saved %s snapshot for summoner #%d on %s",
		*request.Label,
		sid,
		request.Quickdates[0]))

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

	log.Println("Establishing connection to beanstalk...")
	bs, cerr := beanstalk.Dial("localhost:11300")

	if cerr != nil {
		log.Fatal(cerr)
	}

	for {
		// Wait until there's a message available.
		j, err := bs.Reserve()
		log.Println("Received request", j.ID)

		if err != nil {
			log.Fatal(err)
		}

		// Unmarshal the request and kick off a bunch of goroutines, one
		// per summoner included in the request.
		request := proto.JoinRequest{}
		gproto.Unmarshal(j.Body, &request)

		for _, summoner := range request.Summoners {
			go handle_summoner(request, summoner)
			GR_GROUP.Add(1)
		}

		// Wait until all summoners are done before moving on to the next request.
		GR_GROUP.Wait()

		// The task is done; we can delete it from the queue.
		bs.Delete(j.ID)
	}
}
