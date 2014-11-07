package main

import (
	gproto "code.google.com/p/goprotobuf/proto"
	data "datamodel"
	"fmt"
	"log"
	"proto"
	"query"
	"strings"
)

/**
 * Load the list of known summoners from MongoDB. If the summoner
 * isn't labeled with a name in the backend then it won't be
 * user-retrievable.
 */
func load_summoners() map[string]uint32 {
	retriever := data.LoLRetriever{}

	summoners_iter := retriever.GetKnownSummonersIter()
	summoners := make(map[string]uint32)

	for summoners_iter.HasNext() {
		summoner := summoners_iter.Next()

		// If we've got a valid summoner and their name is set, store it
		// in the lookup table.
		if summoner.SummonerId > 0 && len(summoner.Metadata.SummonerName) > 0 {
			summoners[strings.ToLower(summoner.Metadata.SummonerName)] = summoner.SummonerId
		}
	}

	summoners["brigado"] = 36142441
	return summoners
}

func handle_request(request *query.QueryRequest, summoners map[string]uint32, qm *query.QueryManager) {
	// The data structure stores the Query generically so we need to cast it to the application-spceific
	// query type.
	name_request := request.Query.(*proto.NameRequest)
	lc_name := strings.ToLower(*name_request.Name)

	log.Println("Request received:", *name_request.Name)

	sid, ok := summoners[lc_name]
	// Found the name
	response := proto.NameResponse{}
	response.Name = name_request.Name

	if ok {
		response.Id = gproto.Uint32(sid)
		log.Println( fmt.Sprintf("Found mapping [%s = %d]", *name_request.Name, sid) )
		// Didn't find the name
	} else {
		response.Id = gproto.Uint32(0)
		log.Println( fmt.Sprintf("Mapping not found [%s = ?]", *name_request.Name) )
	}
	qm.Reply(request, &response)
}

func main() {
	// Load all summoner data
	summoner_map := load_summoners()
	log.Println(fmt.Sprintf("Loaded %d summoners from backend.", len(summoner_map)))

	log.Println("Opening port...")
	manager := query.QueryManager{}
	manager.Connect(14004)

	log.Println("Nameserver ready.")
	for {
		request := manager.Listen(&proto.NameRequest{})
		go handle_request(&request, summoner_map, &manager)
	}
}
