package main

import (
	"bufio"
	"gamelog"
	gproto "code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"net"
	"net/http"
	"log"
	"proto"
	"snapshot"
	"switchboard"
)

type FrontendStatsRequest struct {
	KnownSummoner	bool
	Player		gamelog.PlayerType
	Records		snapshot.SummonerRecord
}

// Switchboard to be used by all of the goroutines.
var lookup switchboard.SwitchboardClient
var cerr = error(nil)

func init() {
        conn, _ := net.ResolveTCPAddr("tcp", "lookup.loltracker.com:14004")
	lookup, cerr = switchboard.NewClient("tcp", conn)

        if cerr != nil {
                log.Fatal("Couldn't find any available backends.")
        } else {
		log.Println("Connected to nameserver.")
	} 
}

/**
 * Make a call to the lookup server to convert a summoner name into a
 * summoner ID. This function will return the summoner ID as well as
 * a boolean flag indicating whether the summoner was found by the
 * lookup server. If !valid, the value of the summoner ID can be
 * assumed to be unusable.
 */
func lookup_summoner(name string) (uint32, bool) {
	// Get a stream to the lookup server.
        conn, _ := lookup.GetStream()

	request := proto.NameRequest{}
        request.Name = gproto.String(name)

        // Send a request
        rw := bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn))
        data, _ := gproto.Marshal(&request)
        rw.WriteString(string(data) + "|")
        rw.Flush()
        
        // Receive a response
        response := proto.NameResponse{}
        reply, _ := rw.ReadString('|')
        gproto.Unmarshal([]byte(reply[:len(reply)-1]), &response)

	return *response.Id, (*response.Id != 0)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("index requested")

	name := r.FormValue("name")
	// Lookup the summoner ID from a lookup server.
	summoner_id, valid := lookup_summoner(name)

        stat_request := FrontendStatsRequest{}
        stat_request.KnownSummoner = valid
        stat_request.Player.Name = name
        stat_request.Player.SummonerId = summoner_id

	if valid {
		retriever := snapshot.Retriever{}
		retriever.Init()
		// Make a request to the backend to get the snapshot data for
		// this summoner.
		stat_request.Records = retriever.GetSnapshots(summoner_id)
	}

	response_string, _ := json.Marshal(stat_request)
	w.Write( []byte(response_string) )
}

func main() {
        http.HandleFunc("/", index_handler)
	// No-op handler for favicon.ico, since it'll otherwise generate an extra call to index_handler.
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {}) 
//        http.HandleFunc("/summoner/", simple_summoner)

        // Initialize the connection to the name lookup service
//	conn, _ := net.ResolveTCPAddr("tcp", "lookup.loltracker.com:14003")
//      lookup, cerr := switchboard.NewClient("tcp", conn)

      //  if cerr != nil {
      //          log.Fatal("Couldn't find any available backends.")
      //  }

        // Serve any files in static/ directly from the filesystem.
        http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
                log.Println("GET", r.URL.Path[1:])
                http.ServeFile(w, r, "html/"+r.URL.Path[1:])
        })

        log.Println("Awaiting requests...")
        log.Fatal("Couldn't listen on port 8088:", http.ListenAndServe(":8088", nil))
}

