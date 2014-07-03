package main

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"libcleo"
	"log"
	"net"
	"net/http"
	"os"
	"proto"
	"query"
	"strings"
	"switchboard"
)

type ChampionPageParam struct {
	Name   string
	ImgURL string
}

type PageParams struct {
	Title string

	Matching            uint32
	Matching_Percentage float32
	Available           uint32
	Total               uint32

	Allies  []ChampionPageParam
	Enemies []ChampionPageParam

	Valid bool
}

type SubqueryBundle struct {
	Explorer	int32
	Response	proto.QueryResponse
}

const ENABLE_EXPLORATORY_SUBQUERIES = true

// TODO: this probably shouldn't be a global.
var query_id = 0

// TODO: figure out how to pass this to function handler in a way that will be
// 	maintained between connections.
var switchb = switchboard.SwitchboardClient{} //, _ = switchboard.NewClient("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 14002})

// Fetch index.html (the main app). Simple, static file.
func index_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("index requested")

	data, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		log.Println("index.html not present!")
		http.NotFound(w, r)
	} else {
		w.Write(data)
	}
}

/**
 * This function is a handler for basic team queries that specify a
 * list of allies and enemies. It builds a query from the URL parameters
 * and sends it to a Cleo backend. Once the Cleo backend responds it
 * serializes the response to JSON and returns it to the client.
 */
func simple_team(w http.ResponseWriter, r *http.Request) {
	allies := strings.Split(r.FormValue("allies"), ",")
	enemies := strings.Split(r.FormValue("enemies"), ",")

	qry := form_request(allies, enemies)
	response := proto.QueryResponse{}

	is_valid := validate_request(qry)
	if is_valid {
		log.Println(fmt.Sprintf("%s: valid team query [allies=;enemies=]", query.GetQueryId(qry)))
		response = request(qry)
		
		if ENABLE_EXPLORATORY_SUBQUERIES {
			log.Println(fmt.Sprintf("%s: submitting subqueries", query.GetQueryId(qry)))
			
			subqueries := make(chan SubqueryBundle)
			num_subqueries := 0
			
			// Launch a bunch of different subqueries.
			for _, cid := range proto.ChampionType_value {
				go explore_subquery(qry, cid, subqueries)
				num_subqueries += 1
			}
			
			// Collect all of the subquery responses.
			for i := 0; i < num_subqueries; i++ {
				bundle := <- subqueries
				
				next_champ := proto.QueryResponse_ExploratoryChampionSubquery {
					Explorer: proto.ChampionType(bundle.Explorer),
					Results: bundle.Results,
					Valid: (bundle.Response != nil),
				}
					
				response.NextChamp = append(response.NextChamp, next_champ)
			}
		}

		data, err := json.Marshal(response)

		if err != nil {
			log.Println("SimpleTeam: invalid response.")
			// TODO: handle this error appropriately.
		}

		w.Write(data)
	} else {
		log.Println(w, "SimpleTeam: invalid query.")
	}
}

/**
 * This function issues another modified query to the backend that includes
 * an explorer ID (champion ID) that was not specified by the user. This
 * will compute all of the standard stats if the user were to add this
 * champion.
 */
func explore_subquery(qry proto.GameQuery, explorer_id int32, out chan SubqueryBundle) {
	qry.Winners = append(qry.Winners, explorer_id)
	
	// If the query is valid, submit it and pass the response back to the
	// output channel.
	if validate_request(qry) {
		response := request(qry)
		
		bundle_response := SubqueryBundle {
			Explorer: explorer_id,
			Response: response,
		}	
		
		out <- bundle_response
	// If it's not a valid query we should return a nil response so that
	// we can still aggregate everything appropriately.
	} else {
		bundle_response := SubqueryBundle {
			Explorer: explorer_id,
			Response: nil,
		}
		
		out <- bundle_response
	}
}


// Validate current just checks to make sure that all tokens are real.
func validate_request(qry proto.GameQuery) bool {
	for _, winner := range qry.Winners {
		if winner == proto.ChampionType_UNKNOWN {
			return false
		}
	}

	for _, loser := range qry.Losers {
		if loser == proto.ChampionType_UNKNOWN {
			return false
		}
	}

	return true
}

func form_request(allies []string, enemies []string) proto.GameQuery {
	qry := proto.GameQuery{}

	qry.QueryProcess = gproto.Uint64(uint64(os.Getpid()))
	qry.QueryId = gproto.Uint64(uint64(query_id))

	query_id += 1

	// Map the strings specified in the url to ChampionType's.
	for _, name := range allies {
		if len(name) > 0 {
			log.Println(fmt.Sprintf("%s: ally required = %s", query.GetQueryId(qry), libcleo.String2ChampionType(name)))
			qry.Winners = append(qry.Winners, libcleo.String2ChampionType(name))
		}
	}

	for _, name := range enemies {
		if len(name) > 0 {
			log.Println(fmt.Sprintf("%s: enemy required = %s", query.GetQueryId(qry), libcleo.String2ChampionType(name)))
			qry.Winners = append(qry.Losers, libcleo.String2ChampionType(name))
		}
	}

	return qry
}

func request(qry proto.GameQuery) proto.QueryResponse {
	//  Get a switchboard socket to talk to server
	conn, cerr := switchb.GetStream()
	
	if cerr != nil {
		log.Println(fmt.Sprintf("%s: couldn't connect to a Cleo server.", query.GetQueryId(qry)))
		return proto.QueryResponse{Successful: gproto.Bool(false)}
	}

	// Form a GameQuery.
	data, _ := gproto.Marshal(&qry)
	log.Println(fmt.Sprintf("%s: query sent", query.GetQueryId(qry)))

	rw := bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn))
	rw.WriteString(string(data) + "|")
	rw.Flush()

	// Unmarshal the response.
	response := proto.QueryResponse{}
	log.Println(fmt.Sprintf("%s: awaiting response...", query.GetQueryId(qry)))

	reply, _ := rw.ReadString('|')
	
	// If we get a zero-length reply this means the backend crashed. Don't
	//  freak out. We got this.
	if len(reply) == 0 {
		return response
	}
	
	gproto.Unmarshal([]byte(reply[:len(reply)-1]), &response)
	log.Println(fmt.Sprintf("%s: valid response received", query.GetQueryId(qry)))

	return response
}

func main() {
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/team/", simple_team)
	
	// Initialize the connection to 
	cerr := error(nil)
	switchb, cerr = switchboard.NewClient("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 14002})

	if cerr != nil {
		log.Fatal("Couldn't find any available backends.")
	}

	// Serve any files in static/ directly from the filesystem.
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET", r.URL.Path[1:])
		http.ServeFile(w, r, "html/"+r.URL.Path[1:])
	})

	log.Println("Awaiting requests...")
	log.Fatal("Couldn't listen on port 8088:", http.ListenAndServe(":8088", nil))
}
