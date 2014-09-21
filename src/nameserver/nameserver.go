package main

import (
	"fmt"
	gproto "code.google.com/p/goprotobuf/proto"
	"log"
	"proto"
	"query"
	"snapshot"
	"strings"
)

/**
 * Load the list of known summoners from MongoDB. If the summoner
 * isn't labeled with a name in the backend then it won't be
 * user-retrievable.
 */
func load_summoners() map[string]uint32 {
	retriever := snapshot.Retriever{}
	retriever.Init()

	iter := retriever.GetSnapshotsIter()
	summoners := make(map[string]uint32)

	snap := snapshot.SummonerRecord{}
	for iter.Next(&snap) {
		if len(snap.SummonerName) > 9 {
			summoners[strings.ToLower(snap.SummonerName)] = snap.SummonerId
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
	// Didn't find the name
	} else {
		response.Id = gproto.Uint32(0)
	}
	qm.Reply(request, &response)
}

func main() {
	// Load all summoner data
	summoner_map := load_summoners()
	log.Println( fmt.Sprintf("Loaded %d summoners from backend.", len(summoner_map)) )

	log.Println("Opening port...")
	manager := query.QueryManager{}
	manager.Connect(14004)

	log.Println("Nameserver ready.")
	for {
		request := manager.Listen(&proto.NameRequest{})
		go handle_request(&request, summoner_map, &manager)
	}
}
