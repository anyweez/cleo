package main

import "fmt"
import "gamelog"
import "query"
import "math"

// TODO: known data quality error. Some games only include a single team,
// and it looks like single teams often have < 5 players (1 and 3 seem
// to occur in 110-sample set). Check process.py for issues with parse_team().

type RecordTuple struct {
	id int
	
	Event *gamelog.GameRecord
	Query *query.Query
	
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
	filter_in := make(chan *RecordTuple, 250)
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
	game_log := gamelog.Read("/media/vortex/corpora/lolking/test.gamelog")
	
	fmt.Printf("Read in %d games in gamelog.\n", len(game_log.Games))
	
	parsed_query := query.ReadFile("queries/thresh.lkg")
	
	///// Running the query /////
	
	// Kick off a bunch of filter_event goroutines, then push events
	// into the channel that they read from.
	for i, event := range game_log.Games {
		filter_in <- &RecordTuple{i, event, parsed_query, FILTER_UNTESTED, LABELER_UNTESTED}
	}
	
//	filtered := make([]*RecordTuple, len(game_log.Games))
//	labeled := make([]*RecordTuple, len(game_log.Games))
	filtered_id := 0
	labeled_id := 0

	for i := 0; i < len(game_log.Games); i++ {
		record := <- label_out

		if record.filter_condition == FILTER_ACCEPTED {
//			filtered[filtered_id] = record
			filtered_id += 1
			
			if record.labeler_condition == LABELER_ACCEPTED {
//				labeled[labeled_id] = record
				labeled_id += 1
			}
		}
	}
	
	fmt.Println( "Filtered down to", filtered_id, "/", len(game_log.Games) )
	fmt.Println( "Matched down to", labeled_id, "/", filtered_id )
	success_rate := float64(labeled_id) / float64(filtered_id)
	
	fmt.Println( "Success rate:", math.Floor(success_rate * 10000) / 100, "%")
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
		
		for _, check := range record.Query.Filters {
			if !check(record.Event, record.Query) && !done {
				record.filter_condition = FILTER_REJECTED
				
				output <- record
				done = true
			}
		}
		
		if !done {
			record.filter_condition = FILTER_ACCEPTED
			output <- record
		}
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
		}
		
		for _, check := range record.Query.Labelers {
			if !check(record.Event, record.Query) && !done {
				record.labeler_condition = LABELER_REJECTED
				
				output <- record
				done = true
			}
		}
		
		if !done {
			record.labeler_condition = LABELER_ACCEPTED		
			output <- record
		}
	}
}
