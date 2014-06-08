package main

// Lolstat is the core binary that evaluates and responds to queries. It can
// currently handle queries of the form:
//   "How many games has [champion combination X] won against [champion combination Y]?
//
// It depends on the fetcher and packer binaries to prepare indices that it
// can use for fast searching, and only stores the game ID for each game
// (no additional metadata) in order to minimize the required memory
// footprint and maximize the amount of information that can be kept
// accessible at once.

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"io/ioutil"
	"libcleo"
	"log"
	"proto"
	"query"
	"sort"
	"time"
)

// Build a wrapper data structure that can be used to enable fast sorting
// on the game ID's.
type idList []uint64

func (x idList) Len() int {
	return len(x)
}

func (x idList) Less(i, j int) bool {
	return x[i] < x[j]
}

func (x idList) Swap(i, j int) {
	tmp := x[i]

	x[i] = x[j]
	x[j] = tmp
}

func (x idList) toUint() []uint64 {
	l := make([]uint64, 0, len(x))

	for _, val := range x {
		l = append(l, val)
	}

	return l
}

func get_sorted(x []uint64) []uint64 {
	ids := make(idList, 0, len(x))

	for _, val := range x {
		ids = append(ids, val)
	}
	sort.Sort(ids)

	return ids.toUint()
}

// Reads in a ChampionGameList file that can be used for searching.
// TODO: Retrieve the file.
func read_pcgl(filename string) libcleo.LivePCGL {
	packed_pcgl := proto.PackedChampionGameList{}
	// Unmarshal data.
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Couldn't read PCGL;", filename, "does not exist.")
	}
	gproto.Unmarshal(bytes, &packed_pcgl)

	// Convert to format that's faster to search through.
	pcgl := libcleo.LivePCGL{}
	pcgl.Champions = make(map[proto.ChampionType]libcleo.LivePCGLRecord)
	pcgl.All = make([]uint64, 0, 100)

	for _, champ := range packed_pcgl.Champions {
		_, exists := pcgl.Champions[*champ.Champion]

		if !exists {
			pcgl.Champions[*champ.Champion] = libcleo.LivePCGLRecord{}
		}
		r := pcgl.Champions[*champ.Champion]
		r.Winning = get_sorted(champ.Winning)
		r.Losing = get_sorted(champ.Losing)

		pcgl.Champions[*champ.Champion] = r
	}

	pcgl.All = get_sorted(packed_pcgl.All)

	return pcgl
}

func main() {
	// Query connection manager
	qm := query.QueryManager{}

	// Inputs
	query_requests := make(chan query.GameQueryRequest, 100)
	// Outputs
	query_completions := make(chan query.GameQueryResponse, 100)

	fmt.Printf("Loading gamelog.\n")
	pcgl := read_pcgl("latest.pcgl")
	log.Println("Read", len(pcgl.All), "events into PCGL.")

	qm.Connect()

	// Kick off some goroutines that can handle queries.
	for i := 0; i < 1; i++ {
		go query_handler(query_requests, &pcgl, query_completions, &qm)
	}

	// Infinitely loop through queries as they come in. Currently this
	// will only handle one at a time but should be trivial to parallelize
	// once the time is right.
	for {
		query_requests <- qm.Await()
		time.Sleep(1 * time.Second)
	}
}

// A query handler filters down the list of game ID's to the set identified in
// the provided query. They are run as goroutines and can handle a single
// query at a time. They each make a copy of the lists in the CGl so that
// all queries are independent and unaffected by others.
//
// Two values need to be computed: the MATCHING games and the ELIGIBLE games.
//   - Matching games are those that have all of the requested players
//     on the requested teams (numerator).
//   - Eligible games are those that have all of the requested players
//     on one team or another (denominator).
//
// The general algorithm for computing each is as follows:
//
// MATCHING
// Start with a list of all games. For each winning champion, find the
// overlap between the full set with the winning game set for that champion.
// Then do the same thing for all losing champions.
//
// ELIGIBLE
// Start with a list of all games. For each winning champion, find the
// overlap between the full set and the losing game set for that champion.
// Then do the same thing for all losing champions (find the winning set
// for them). Then merge the output from the MATCHING set with the lists
// from the ELIGIBLE set to produce the final ELIGIBLE set.
func query_handler(input chan query.GameQueryRequest, pcgl *libcleo.LivePCGL, output chan query.GameQueryResponse, qm *query.QueryManager) {
	for {
		request := <-input
		log.Println(fmt.Sprintf("%s: handling query", request.Id))

		// Eligible gamelist contains all games that match, irrespective of team.
		eligible_wins_gamelist := make([]uint64, len(pcgl.All))
		eligible_losses_gamelist := make([]uint64, len(pcgl.All))

		// Matching gamelist contains all games that match, respective of team.
		matching_gamelist := make([]uint64, len(pcgl.All))

		copy(eligible_wins_gamelist, pcgl.All)
		copy(eligible_losses_gamelist, pcgl.All)
		copy(matching_gamelist, pcgl.All)

		// Merge all game ID's, first matching the winning parameters.
		for _, champion := range request.Query.Winners {
			// Update the matching gamelist to include just the overlap between these two lists.
			overlap(&matching_gamelist, pcgl.Champions[champion].Winning)
			overlap(&eligible_losses_gamelist, pcgl.Champions[champion].Losing)
		}

		// Then match all losers.
		if len(request.Query.Losers) > 0 {
			for _, champion := range request.Query.Losers {
				overlap(&matching_gamelist, pcgl.Champions[champion].Losing)
				overlap(&eligible_wins_gamelist, pcgl.Champions[champion].Winning)
			}
		} else {
			eligible_wins_gamelist = eligible_wins_gamelist[:0]
		}

		// Step #2: Eligible set includes all that matched and all those that contained
		//   the proposed champions in the teams provided (victory status ignored).
		eligible_gamelist := merge(eligible_wins_gamelist, eligible_losses_gamelist)
		eligible_gamelist = merge(matching_gamelist, eligible_gamelist)

		// Prepare the response.
		response := query.GameQueryResponse{}
		response.Request = &request

		response.Response = &proto.QueryResponse{
			Available:  gproto.Uint32(uint32(len(eligible_gamelist))),
			Matching:   gproto.Uint32(uint32(len(matching_gamelist))),
			Total:      gproto.Uint32(uint32(len(pcgl.All))),
			Successful: gproto.Bool(true),
		}

		log.Println(fmt.Sprintf("%s: response generated", request.Id))
		// Send it to the query responder queue to take care of the
		// actual transmission and associated events.
		qm.Respond(&response)
	}
}

// Overlap accepts two lists of uints and reduces FIRST to the overlap
// between both lists.
// Assumes that both lists are ordered.
func overlap(first *[]uint64, second []uint64) {
	// parallel_counter indexes into SECOND and may move at a different
	// rate than i.
	parallel_counter := 0

	if len(*first) == 0 || len(second) == 0 {
		(*first) = (*first)[:0]
		return
	}

	for i := 0; i < len(*first); i++ {
		// Loop through until the second array's value is greater than or
		// equal to the primary array. We should not reset this counter
		// variable.
		for second[parallel_counter] < (*first)[i] {
			if parallel_counter+1 < len(second) {
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
			(*first) = append((*first)[:i], (*first)[i+1:]...)

			i -= 1
		}
	}
}

// Merge combined two lists into a single ordered list. It assumes that
// both input lists are ordered as well. It will only copy duplicated
// values one time, i.e. it removes duplicates.
func merge(first []uint64, second []uint64) []uint64 {
	full := make([]uint64, 0, len(first)+len(second))

	first_i := 0
	second_i := 0

	// Move through the list until we get to the end of one of them.
	for first_i < len(first) && second_i < len(second) {
		// If next value in FIRST is less than next value in SECOND,
		// copy value from FIRST and move on.
		if first[first_i] < second[second_i] {
			full = append(full, first[first_i])

			first_i += 1
			// If the two values are the same, copy one of them over. This
			// will remove duplicates.
		} else if first[first_i] == second[second_i] {
			full = append(full, first[first_i])

			first_i += 1
			second_i += 1
			// Otherwise if FIRST > SECOND, copy over second.
		} else {
			full = append(full, second[second_i])

			second_i += 1
		}
	}

	// Copy over all remaining values from FIRST and SECOND.
	if first_i < len(first) {
		full = append(full, first[first_i:]...)
	}

	if second_i < len(second) {
		full = append(full, second[second_i:]...)
	}

	return full
}
