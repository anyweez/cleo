package main

/**
 * This process reads in a list of summoner ID's and produces a record for each
 * summoner keyed by summoner ID. Usage is as follows:
 *
 * ./jsummoner --date=2014-08-07 --summoners=data/summoners/0001.list
 */

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"flag"
	//	"libcleo"
	//	"libproc"
	"log"
	"proto"
	"snapshot"
	"strconv"
	"strings"
)

var SUMMONER_FILE = flag.String("summoners", "", "The filename containing a list of summoner ID's.")
var TARGET_DATE = flag.String("date", "", "The date to join in YYYY-MM-DD form.")

/**
 * Goroutine that generates a report for a single summoner ID. It reads
 * through all game records and retains those that were played by the
 * target summoner ID. It then condenses them into a single PlayerSnapshot
 * and saves it to MongoDB.
 */
func handle_summoner(sid uint32, input chan *proto.GameRecord) {
	games := make([]*proto.GameRecord, 0, 10)

	// Keep reading from the channel until nil comes through, then we're
	// done receiving info.
	keeper := false
	gr := <-input
	for gr != nil {
		for _, team := range gr.Teams {
			for _, player := range team.Players {
				if *player.Player.SummonerId == sid {
					keeper = true
				}
			}
		}

		if keeper {
			games = append(games, gr)
		}

		gr = <-input
	}

	// Now all games have been processed. We need to save the set of
	// games to a PlayerSnapshot for today.
	snap := proto.PlayerSnapshot{}
	snap.Timestamp = gproto.Uint64(convert_ts(*TARGET_DATE))
	snap.Games = games
	snap.SummonerId = gproto.Uint32((uint32)(sid))
	// TODO: add rank to this snapshot.

	// Commit to datastore
	retriever := snapshot.Retriever{}
	retriever.Init()

	retriever.SaveSnapshot(&snap)

	//	save_snapshot(snapshot)
}

/**
 * Record a single snapshot into MongoDB.
 */
/*
 * func save_snapshot(snapshot proto.PlayerSnapshot) {
	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("players-" + *TARGET_DATE)
	defer session.Close()

	snap := libproc.PlayerSnapshotContainer{}
	snap.Snapshot, _ = gproto.Marshal(&snapshot)
	snap.Timestamp = *snapshot.Timestamp
	snap.SummonerId = *snapshot.SummonerId

	games_collection.Insert(snap)
}
*/

/**
 * Retrieve all games that are relevant to the calculation. Currently
 * fetching all games, but should probably be reduced to X days.
 *
 * TODO: fetch games in range (two_weeks_ago, TARGET_DATE)
 */

/*func get_games() []proto.GameRecord {
	games := make([]proto.GameRecord, 0, 100)

	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("games")
	defer session.Close()

	// TODO: this should be a sliding window (2 weeks) instead of a single day
	query := games_collection.Find(bson.M{ "timestamp": convert_ts(*TARGET_DATE) })
	result_iter := query.Iter()

	result := libcleo.RecordContainer{}
	for result_iter.Next(&result) {
		game := proto.GameRecord{}
		gproto.Unmarshal(result.GameData, &game)

		games = append(games, game)
	}

	return games
}*/

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
	sids := snapshot.ReadSummonerIds(*SUMMONER_FILE)
	sid_chan := make([]chan *proto.GameRecord, len(sids))

	running := make(chan bool)

	/* Create a bunch of goroutines, one per summoner, that can be used
	 * to filter records. */
	for i, sid := range sids {
		go handle_summoner(sid, sid_chan[i])
	}

	/* Retrieve all events. */
	games := retriever.GetGames(*TARGET_DATE)

	/* Pass each event to a goroutine that handles each summoner ID. */
	for i, _ := range sids {
		for _, game := range games {
			sid_chan[i] <- &game
		}

		// Sending a nil means that the goroutine doesn't need to wait
		// for more data.
		sid_chan[i] <- nil
	}

	for i := 0; i < len(sids); i++ {
		<-running
	}
}
