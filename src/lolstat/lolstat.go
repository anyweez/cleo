package main

import "fmt"
import "query"
import gproto "code.google.com/p/goprotobuf/proto"

type PackedChampionGameList struct {
	Winning		[][]uint32
	Losing		[][]uint32
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
	t := make([]uint32, 5, 10)
	u := make([]uint32, 3, 10)
	
	t[0], t[1], t[2], t[3], t[4] = 1, 2, 3, 4, 5
	u[0], u[1], u[2] = 2, 4, 5
	
	fmt.Println("Before:", t)
	overlap(&t, u)
	fmt.Println("After:", t)
	
	return

	query_requests := make(chan query.GameQueryRequest, 100)
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
		query_requests <- qm.Await()
	}
}

func query_handler(input chan query.GameQueryRequest, cgl *PackedChampionGameList, output chan query.GameQueryResponse) {
	for {
		request := <-input
		
		// Eligible gamelist contains all games that match, irrespective of team.
		eligible_gamelist := make([]uint32, 0, 100)
		eligible_added := false
		
		// Matching gamelist contains all games that match, respective of team.
		matching_gamelist := make([]uint32, 0, 100)
		matching_added := false
		
		// Merge all game ID's, first matching the winning parameters. 
		for _, champion_id := range request.Query.Winners {
			// Update the matching gamelist.
			if matching_added {
				overlap(&matching_gamelist, cgl.Winning[champion_id])
			} else {
				matching_gamelist = append(matching_gamelist, cgl.Winning[champion_id]...)
				matching_added = true
			}
			
			// Update the eligible gamelist.
			if eligible_added {
				overlap(&eligible_gamelist, cgl.Losing[champion_id])
			} else {
				eligible_gamelist = append(eligible_gamelist, cgl.Losing[champion_id]...)
				eligible_added = true
			}
		}
		
		eligible_added = false
		matching_added = false
		
		// Merge all game ID's, now matching the losing parameters.
		for _, champion_id := range request.Query.Losers {
			if matching_added {
				overlap(&matching_gamelist, cgl.Losing[champion_id])
			} else {
				matching_gamelist = append(matching_gamelist, cgl.Losing[champion_id]...)
				matching_added = true
			}
			
			if eligible_added {
				overlap(&eligible_gamelist, cgl.Winning[champion_id])
			} else {
				eligible_gamelist = append(eligible_gamelist, cgl.Winning[champion_id]...)
				eligible_added = true
			}
		}
		
		eligible_gamelist = append(eligible_gamelist, matching_gamelist...)
		
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

func overlap(first *[]uint32, second []uint32) {
	parallel_counter := 0
	
	for i := 0; i < len(*first); i++ {
		// Loop through until the second array's value i greater than or
		// equal to the primary array. We do not need to reset this counter
		// variable.
		for second[parallel_counter] < (*first)[i] {
			if parallel_counter + 1 <= len(second) {
				parallel_counter += 1
			}
			
			fmt.Println("Incrementing counter to", parallel_counter)
		}
		// Once the secondary index catches up, if it's beyond the primary
		// then the primary doesn't exist. If they're equivalent then we
		// keep the primary value.
		if second[parallel_counter] > (*first)[i] {
			(*first)[i] = 0
			
			fmt.Println("Value doesn't exist. Clearing.")
		}

		// Increment if we're not a tthe end.
		if parallel_counter + 1 <= len(second) {
			parallel_counter += 1
		}
	}
}
