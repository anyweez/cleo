package main

import (
	"bufio"
//	"bytes"
	gproto "code.google.com/p/goprotobuf/proto"
	data "datamodel"
	"encoding/json"
	"fmt"
//	"io/ioutil"
	"log"
	"net"
	"net/http"
	"proto"
	"switchboard"
//	"text/template"
)

type FrontendStatsRequest struct {
	KnownSummoner bool
	Player        data.PlayerType
	Records       data.SummonerRecord
}

//type IndexTemplate struct {
//	OverviewTab		string
//}

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

func summoner_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("summoner requested")

	name := r.FormValue("name")
	// Lookup the summoner ID from a lookup server.
	summoner_id, valid := lookup_summoner(name)

	stat_request := FrontendStatsRequest{}
	stat_request.KnownSummoner = valid
	stat_request.Player.Name = name
	stat_request.Player.SummonerId = summoner_id

	retriever := data.LoLRetriever{}

	if valid {
		// Make a request to the backend to get the snapshot data for
		// this summoner.
		stat_request.Records, _ = retriever.GetSummoner(summoner_id)
	}

	// TODO: Remove this once everything works well.
	if name == "brigado" {
		stat_request.Records.Daily = make(map[string]*data.PlayerSnapshot)
		stat_request.Records.Daily["2014-09-17"] = &data.PlayerSnapshot {
			SummonerId: stat_request.Player.SummonerId,
			GamesList: []uint64{5, 10, 11, 19},
			Stats: make(map[string]data.Metric),
			CreationTimestamp: 0,
		}
		
		stat_request.Records.Daily["2014-09-17"].Stats["kda"] = data.SimpleNumberMetric { Value: 11 }
		stat_request.Records.Daily["2014-09-17"].Stats["minionKills"] = data.SimpleNumberMetric { Value: 20 }

		// Add data for 9/18
		stat_request.Records.Daily["2014-09-18"] = &data.PlayerSnapshot {
			SummonerId: stat_request.Player.SummonerId,
			GamesList: []uint64{5, 10, 11, 19},
			Stats: make(map[string]data.Metric),
			CreationTimestamp: 0,
		}
		
		stat_request.Records.Daily["2014-09-18"].Stats["kda"] = data.SimpleNumberMetric { Value: 13 }
		stat_request.Records.Daily["2014-09-18"].Stats["minionKills"] = data.SimpleNumberMetric { Value: 24 }

	// Add data for 9/19
		stat_request.Records.Daily["2014-09-19"] = &data.PlayerSnapshot {
			SummonerId: stat_request.Player.SummonerId,
			GamesList: []uint64{5, 10, 11, 19},
			Stats: make(map[string]data.Metric),
			CreationTimestamp: 0,
		}
		
		stat_request.Records.Daily["2014-09-19"].Stats["kda"] = data.SimpleNumberMetric { Value: 14.1 }
		stat_request.Records.Daily["2014-09-19"].Stats["minionKills"] = data.SimpleNumberMetric { Value: 37 }
	}

	log.Println(stat_request)
	response_string, jerr := json.Marshal(stat_request)

	if jerr != nil {
		log.Fatal("fatal error:", jerr)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(response_string))
	
	log.Println("Response sent")
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("index requested")

/*
	it := IndexTemplate{}
	// Overview tab
	ovt, _ := ioutil.ReadFile("html.stats/static/overview-tab.html")
	it.OverviewTab = string(ovt)
	
	tmpl := template.New("index").Delims("<<", ">>")
	tmpl, _ = tmpl.ParseFiles("html.stats/index.html")
	output := bytes.NewBufferString("")
	
	tmpl.Execute(output, it)
	log.Println( len(output.String()) )
	fmt.Fprintf(w, output.String())
	*/
	http.ServeFile(w, r, "html.stats/index.html")
}

func main() {
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/summoner/", summoner_handler)
	// No-op handler for favicon.ico, since it'll otherwise generate an extra call to index_handler.
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	//        http.HandleFunc("/summoner/", simple_summoner)

	// Serve any files in static/ directly from the filesystem.
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET", r.URL.Path[1:])
		http.ServeFile(w, r, "html.stats/"+r.URL.Path[1:])
	})

	log.Println("Awaiting requests...")
	log.Fatal("Couldn't listen on port 8088:", http.ListenAndServe(":8088", nil))
}
