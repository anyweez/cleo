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
)

type ChampionPageParam struct {
	Name	string
	ImgURL	string
}

type PageParams struct {
	Title				string
	
	Matching			uint32
	Matching_Percentage	float32
	Available			uint32
	Total				uint32	
	
	Allies				[]ChampionPageParam
	Enemies				[]ChampionPageParam
	
	Valid		bool
}

// TODO: this probably shouldn't be a global.
var query_id = 0;

// Fetch index.html (the main app). Simple, static file.
func index_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("index requested")

	data, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		log.Println("index.html not present!")
		http.NotFound(w, r)
	} else {
		w.Write( data )
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
		log.Println( fmt.Sprintf("%s: valid team query [allies=;enemies=]", query.GetQueryId(qry)) )
		response = request(qry)
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
	
	qry.QueryProcess = gproto.Uint64( uint64(os.Getpid()) )
	qry.QueryId = gproto.Uint64( uint64(query_id) )
	
	query_id += 1 
	
	// Map the strings specified in the url to ChampionType's.
	for _, name := range allies {
		if len(name) > 0 {
			log.Println( fmt.Sprintf("%s: ally required = %s", query.GetQueryId(qry), libcleo.String2ChampionType(name)) )
			qry.Winners = append(qry.Winners, libcleo.String2ChampionType(name))
		}
	}
	
	for _, name := range enemies {
		if len(name) > 0 {
			log.Println( fmt.Sprintf("%s: enemy required = %s", query.GetQueryId(qry), libcleo.String2ChampionType(name)) )
			qry.Winners = append(qry.Losers, libcleo.String2ChampionType(name))
		}
	}

	return qry
}

func request(qry proto.GameQuery) proto.QueryResponse {
	//  Socket to talk to server
	log.Println( fmt.Sprintf("%s: connecting to cleo server...", query.GetQueryId(qry)) )
	conn, cerr := net.DialTCP("tcp", nil, &net.TCPAddr{IP:net.ParseIP("127.0.0.1"), Port: 14002})
	
	if cerr != nil {
		log.Println( fmt.Sprintf("%s: couldn't connect to a Cleo server.", query.GetQueryId(qry)) )
		return proto.QueryResponse{Successful: gproto.Bool(false)}
	}
	
	// Form a GameQuery.
	data, _ := gproto.Marshal(&qry)	
	log.Println( fmt.Sprintf("%s: query sent", query.GetQueryId(qry)) )
	
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	rw.WriteString(string(data) + "|")
	rw.Flush()
	
	// Unmarshal the response.
	response := proto.QueryResponse{}
	log.Println( fmt.Sprintf("%s: awaiting response...", query.GetQueryId(qry)) )

	reply, _ := rw.ReadString('|')	
	gproto.Unmarshal( []byte(reply[:len(reply)-1]), &response )	
	log.Println( fmt.Sprintf("%s: valid response received", query.GetQueryId(qry)) )
	
	return response
}

func main() {
    http.HandleFunc("/", index_handler)
    http.HandleFunc("/team/", simple_team)

    // Serve any files in static/ directly from the filesystem.
    http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
      log.Println("GET", r.URL.Path[1:])
      http.ServeFile(w, r, "html/" + r.URL.Path[1:])
    })

    log.Println("Awaiting requests...")
    log.Fatal("Couldn't listen on port 8088:", http.ListenAndServe(":8088", nil))
}
