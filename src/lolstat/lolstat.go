package main

import "fmt"
import "gamelog"
import "query"
import "proto"
import gproto "code.google.com/p/goprotobuf/proto"

type RecordTuple struct {
	id int
	
	Event *proto.GameRecord
	Query *proto.GameQuery
	
	filter_condition int
	labeler_condition int
}

const (
	FILTER_UNTESTED = iota
	FILTER_ACCEPTED = iota
	FILTER_REJECTED = iota
)

const (
	LABELER_UNTESTED = iota
	LABELER_ACCEPTED = iota
	LABELER_REJECTED = iota
)

// [1] Component that reads over the event log and reads them into memory.

// [2] Reader that parses queries and converts them into something that can
// be used to read over the datastore.

// [3] Topic-specific EventSelector that maps query atoms to conditions that
// can be used to test queries.

// [4] StatNode that takes a series of events and a query and determines 
// various probabilities given the events.

func main() {
	// Kick off some worker goroutines.	
	// NOTE: if there are 1M events then we'll hit deadlock. This limit
	//   will need to be adjusted.
	filter_in := make(chan *RecordTuple, 1000000)
	bridge := make(chan *RecordTuple)
	label_out := make(chan *RecordTuple)
	
	// Kick off some filter goroutines that run through a set of records
	// and set the filter_condition to ACCEPTED or REJECTED based on
	// parameters specified in a query.
	for i := 0; i < 1; i++ {
		go filter_event(filter_in, bridge)
	}
	
	// Kick off some labeler goroutines that run through a set of records
	// and set the labeler_condition to ACCEPTED or REJECTED based on
	// parameters specified in a query.
	for i := 0; i < 1; i++ {
		go label_filtered(bridge, label_out)
	}
	
	fmt.Printf("Loading gamelog.\n")
	game_log := gamelog.Read("/media/vortex/corpora/lolking/gamelogs/2014-05-04:01.36.gamelog")
	
	fmt.Printf("Read in %d games in gamelog.\n", len(game_log.Games))
	
	qm := query.QueryManager{}
	qm.Connect()
	
	// Infinitely loop through queries as they come in. Currently this
	// will only handle one at a time but should be trivial to parallelize
	// once the time is right.
	for {
		q := qm.Await()
	
		///// Running the query /////
	
		// Kick off a bunch of filter_event goroutines, then push events
		// into the channel that they read from.
		for i, event := range game_log.Games {
			filter_in <- &RecordTuple{i, event, &q, FILTER_UNTESTED, LABELER_UNTESTED}
		}
	
		var filtered_id uint32 = 0
		var labeled_id uint32 = 0

		for i := 0; i < len(game_log.Games); i++ {
			record := <- label_out

			if record.filter_condition == FILTER_ACCEPTED {
				filtered_id += 1
			
				if record.labeler_condition == LABELER_ACCEPTED {
					labeled_id += 1
				}
			}
		}
	
		response := &proto.GameQueryResponse{}
		response.Available = gproto.Uint32(filtered_id)
		response.Matching = gproto.Uint32(labeled_id)
		response.Total = gproto.Uint32( uint32(len(game_log.Games)) )
		
		qm.Respond(response)

//		fmt.Println( "Filtered down to", filtered_id, "/", len(game_log.Games) )
//		fmt.Println( "Matched down to", labeled_id, "/", filtered_id )
//		success_rate := float64(labeled_id) / float64(filtered_id)
//		fmt.Println( "Success rate:", math.Floor(success_rate * 10000) / 100, "%")
	}
}

// A goroutine that accepts an event and a query and will return the
// event or nil, depending on whether the event matches all of the
// conditions of the query. If yes, push the event onto the specified
// channel. If no, push nil.
func filter_event(input chan *RecordTuple, output chan *RecordTuple) {
	fmt.Println("Running filter goroutine.")

	for {
		record := <- input
		done := false

		// Check if both teams exist, regardless of who won.
		winner_team := -1
		loser_team := -1
		
		for i, team := range record.Event.Teams {
			if !done {
				has_winners := contains_all(team, record.Query.Winners)
				has_losers := contains_all(team, record.Query.Losers)
			
				// If this team doesn't contain winners or losers then it's
				// impossible for both teams to exist. Reject this record.
				if !(has_winners || has_losers) {
					record.filter_condition = FILTER_REJECTED
				
					done = true 
				} else {
					// If team matches the Winners requirements and no
					// matching team has been found yet, set this team
					// as the "winning" team.
					if has_winners && winner_team == -1 {
						winner_team = i
					// If another "winning" team has been discovered then
					// this record won't teach us anything, so we should
					// throw it out.
					} else if has_winners {
						record.filter_condition = FILTER_REJECTED
						
						done = true
					}
					// Same logic for Losers as above.
					if has_losers && loser_team == -1 {
						loser_team = i
					} else if has_losers {
						record.filter_condition = FILTER_REJECTED
					
						done = true
					}
				}
			}
		} // end for
		
		if winner_team >= 0 && loser_team >= 0 {
			record.filter_condition = FILTER_ACCEPTED
		} else {
			record.filter_condition = FILTER_REJECTED
		}
		
		output <- record		
	}	

}

func label_filtered(input chan *RecordTuple, output chan *RecordTuple) {
	fmt.Println("Running labeler goroutine.")
	
	for {
		record := <- input
		done := false
		
		// If the record wasn't matched in the filter, it's not eligible to
		// be labeled.
		if record.filter_condition != FILTER_ACCEPTED {
			record.labeler_condition = LABELER_UNTESTED
			
			output <- record
			done = true
		} else {
			for _, team := range record.Event.Teams {
				if *team.Victory {
					if !contains_all(team, record.Query.Winners) {
						record.labeler_condition = LABELER_REJECTED
						done = true
					}
				} else {
					if !contains_all(team, record.Query.Losers) {
						record.labeler_condition = LABELER_REJECTED
						done = true
					}
				}
			}
			if !done {
				record.labeler_condition = LABELER_ACCEPTED		
			}
			output <- record
		}
	} // end infinite loop
}

// Determines whether all champions in WINNERS are on the team TEAM.
// Returns boolean value.
//
// Assumes that there can only be one of each champion on a team in 
// order to improve runtime.
func contains_all(team *proto.Team, winners []proto.ChampionType) bool {
	champ_count := 0
	for _, player := range team.Players {
		for _, champ := range winners {
			if *player.Champion == champ {
				champ_count += 1
			}
		}
	}
	
	return champ_count == len(winners)
}
