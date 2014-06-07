package main

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"io/ioutil"
	"libcleo"
	"log"
	"net"
	"net/http"
	"proto"
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

func simple_team(w http.ResponseWriter, r *http.Request) {
	allies := strings.Split(r.FormValue("allies"), ",")
	enemies := strings.Split(r.FormValue("enemies"), ",")


	query := form_request(allies, enemies)
	response := proto.QueryResponse{}
	
	is_valid := validate_request(query)
	if is_valid {
		log.Println("SimpleTeam: valid query [allies=;enemies=]")
		response = request(query)
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
func validate_request(query proto.GameQuery) bool {
	for _, winner := range query.Winners {
		if winner == proto.ChampionType_UNKNOWN {
			return false
		}
	}

	for _, loser := range query.Losers {
		if loser == proto.ChampionType_UNKNOWN {
			return false
		}
	}

	return true
}

func form_request(allies []string, enemies []string) proto.GameQuery {
	query := proto.GameQuery{}
	
	// Map the 
	for _, name := range allies {
		if len(name) > 0 {
			log.Println("Ally:", libcleo.String2ChampionType(name))
			query.Winners = append(query.Winners, libcleo.String2ChampionType(name))
		}
	}
	
	for _, name := range enemies {
		if len(name) > 0 {
			log.Println("Enemy:", libcleo.String2ChampionType(name))
			query.Winners = append(query.Losers, libcleo.String2ChampionType(name))
		}
	}

	return query
}

func request(query proto.GameQuery) proto.QueryResponse {
	//  Socket to talk to server
	log.Println("Connecting to cleo server...")
	conn, cerr := net.DialTCP("tcp", nil, &net.TCPAddr{IP:net.ParseIP("127.0.0.1"), Port: 14002})
	
	if cerr != nil {
		log.Println("Couldn't connect to a Cleo server.")
	}
	
	// Form a GameQuery.
	data, _ := gproto.Marshal(&query)
	log.Println("CLEO_SEND: allies(thresh) enemies()")
	
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	rw.WriteString(string(data) + "|")
	rw.Flush()
	
	// Unmarshal the response.
	response := proto.QueryResponse{}
	log.Println("Awaiting response...")
	reply, _ := rw.ReadString('|')
	
	gproto.Unmarshal( []byte(reply[:len(reply)-1]), &response )	
	
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
    log.Fatal("Couldn't listen on port 80:", http.ListenAndServe(":80", nil))
}
