package main

import "fmt"
import "query"
import "time"
import "proto"
//import "sort"
import gproto "code.google.com/p/goprotobuf/proto"

// PCGL format: array[champion_id] = [game_ids]
type PackedChampionGameList struct {
	Winning		[][]uint32
	Losing		[][]uint32
	All			[]uint32
}
	
// Reads in a ChampionGameList file that can be used for searching.
// TODO: Retrieve the file and restructure.
func read_cgl(filename string) PackedChampionGameList {
	cgl := PackedChampionGameList{}
	
	// Read through and initialize to length of num_champions
	// Each subarray has a specific length depending on the number of games.
	return cgl
}

func main() {
	// Inputs
	query_requests := make(chan query.GameQueryRequest, 100)
	// Outputs
	query_completions := make(chan query.GameQueryResponse, 100)

	fmt.Printf("Loading gamelog.\n")
	cgl := read_cgl("latest.cgl")

	qm := query.QueryManager{}
	qm.Connect()

	// Kick off some goroutines that can handle queries.
	for i := 0; i < 1; i++ {
		go query_handler(query_requests, &cgl, query_completions)
	}

	// Kick off one goroutine that can handle responding to queries.
	go query_responder(query_completions)

	// Infinitely loop through queries as they come in. Currently this
	// will only handle one at a time but should be trivial to parallelize
	// once the time is right.
	for {
		gqr := query.GameQueryRequest{}
		gqr.Query = &proto.GameQuery{}
		gqr.Id = 1
		fmt.Println(gqr.Query)
		gqr.Query.Winners = append(gqr.Query.Winners, proto.ChampionType_THRESH)

		query_requests <- gqr
//		qm.Await()
//		query_requests <- qm.Await()
		time.Sleep(10 * time.Second)
	}
}

func query_handler(input chan query.GameQueryRequest, cgl *PackedChampionGameList, output chan query.GameQueryResponse) {
	for {
		request := <-input
		fmt.Println("Handling query #", request.Id)
		
		// Eligible gamelist contains all games that match, irrespective of team.
		eligible_gamelist := cgl.All
		
		// Matching gamelist contains all games that match, respective of team.
		matching_gamelist := cgl.All

//		fmt.Println("breakpoint")
		//// Step #1: Matching set (incl. team) ////
		// Merge all game ID's, first matching the winning parameters. 
		for _, champion_id := range request.Query.Winners {
			// Update the matching gamelist to include just the overlap between these two lists.
			overlap(&matching_gamelist, cgl.Winning[champion_id])
			overlap(&eligible_gamelist, cgl.Losing[champion_id])
		}
		
		// Then match all losers.
		for _, champion_id := range request.Query.Losers {
			overlap(&matching_gamelist, cgl.Losing[champion_id])
			overlap(&eligible_gamelist, cgl.Winning[champion_id])
		}
		
		//// Step #2: Eligible set (not incl. team) ////
		eligible_gamelist = append(eligible_gamelist, matching_gamelist...)
//		sort.Sort(eligible_gamelist)
		
		// Prepare the response.
		response := query.GameQueryResponse{}
		response.Id = request.Id
		response.Conn = request.Conn
		
		response.Response.Available = gproto.Uint32( uint32(len(eligible_gamelist)) )
		response.Response.Matching = gproto.Uint32( uint32(len(matching_gamelist)) )
		response.Response.Total = gproto.Uint32(0)
		
		output <- response
	}
}

func query_responder(input chan query.GameQueryResponse) {
	for {
		response := <-input

		fmt.Println(fmt.Sprintf("Responding to query #%d", response.Id))
		// TODO: Send the response.
//		response.Conn.Send(&gproto.Marshal(response.Response))
	}
}

// Overlaps accepsts two lists of uints and reduces FIRST to the overlap
// between both lists. 
// Assumes that both lists are ordered.
func overlap(first *[]uint32, second []uint32) {
	// parallel_counter indexes into SECOND and may move at a different
	// rate than i.
	parallel_counter := 0
	
	for i := 0; i < len(*first); i++ {
		fmt.Println("i=", i, ", parallel_counter=", parallel_counter)
		// Loop through until the second array's value is greater than or
		// equal to the primary array. We should not reset this counter
		// variable.
		for second[parallel_counter] < (*first)[i] {
			if parallel_counter + 1 < len(second) {
				parallel_counter += 1
			} else {
			// If parallel_counter is as big as it can get then none of
			// the other numbers in FIRST can overlap.
				(*first) = (*first)[:i]
				return
			}			
		}
		// Once the secondary index catches up, if it's beyond the primary
		// then the primary doesn't exist. If they're equivalent then we
		// keep the primary value.
		if second[parallel_counter] > (*first)[i] {			
			(*first)[i] = 0
			(*first) = append( (*first)[:i], (*first)[i+1:]... )
			
			fmt.Println(*first)
			i -= 1
		}
	}
}
