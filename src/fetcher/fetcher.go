package main

import gproto "code.google.com/p/goprotobuf/proto"
import "gamelog"

import "encoding/json"
import "net/http"
import "fmt"
import "time"
import "io/ioutil"
import "log"
import "strings"
import "os"
import "bufio"
import "strconv"

const API_KEY = "abebd3e9-00f2-4ba6-997d-0008c2072373"
const NUM_RETRIEVERS = 10

type JSONResponse struct {
	Games 				[]JSONGameResponse `json:"games"`
	SummonerId 			uint64
}

type JSONGameResponse struct {
	FellowPlayers		[]JSONPlayerResponse
	Stats				JSONGameStatsResponse
	
	GameId 				uint64
	CreateDate			uint64
	TeamId				uint
	
	GameMode			string
	GameType			string
	GameSubType			string
}

type JSONGameStatsResponse struct {
	Win					bool
}

type JSONPlayerResponse struct {
	SummonerId			uint64
	TeamId				uint
	ChampionId			uint
}

func read_summoner_ids(filename string) []uint64 {
	// Read the specified file.
	file, err := os.Open(filename)
	if err != nil {
		log.Panic("Cannot find file.")
	}
	defer file.Close()
	
	lines := make([]uint64, 0, 10000)
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		value, _ := strconv.ParseUint(scanner.Text(), 10, 64)
		lines = append(lines, value)
	}
	
	// Return a GameLog
	return lines
}

func main() {
	fmt.Println("Initializing...")
	user_queue := make(chan *gamelog.Player, 1000000)
	retrieval_inputs := make(chan *gamelog.Player, 100)
	retrieval_outputs := make(chan *JSONResponse, 100)

	// Kick off some retrievers that will pull from the retrieval queue.
	for i := 0; i < NUM_RETRIEVERS; i++ {
		go retriever(retrieval_inputs, retrieval_outputs)
	}

	// Launch a record writer that will periodically write out
	go record_writer(retrieval_outputs)

	// Load in a file full of summoner ID's.
	summoner_ids := read_summoner_ids("champions")
	for _, sid := range summoner_ids {
		player := gamelog.Player{}
		player.SummonerId = gproto.Uint64(sid)

		user_queue <- &player
	}
	
	fmt.Println(fmt.Sprintf("Loaded %d summoners...let's do this!", len(summoner_ids)))
	fmt.Println("Writing output every 20 min...")
	// Forever: pull an summoner ID from user_queue, toss it in retrieval_inputs
	// and add it back to user_queue.
	for {
		// Wait for 1.25 seconds
		time.Sleep( 1250 * time.Millisecond )
		current := <- user_queue

		// Push the player to the retrieval queue.
		retrieval_inputs <- current
//		user_queue <- player
	}
}

func retriever(input chan *gamelog.Player, output chan *JSONResponse) {
	url := "https://prod.api.pvp.net/api/lol/na/v1.3/game/by-summoner/%d/recent?api_key=%s"

	for {
		player := <- input

		resp, err := http.Get(fmt.Sprintf(url, *player.SummonerId, API_KEY))
		if err != nil {
			fmt.Println("Error retrieving data")
			fmt.Println(err)
		} else {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
		
			json_response := JSONResponse{}
			json.Unmarshal(body, &json_response)
			
			output <- &json_response
		}
	}
}

// Adaptor to convert the JSON format to GameLog format.
func record_writer(input chan *JSONResponse) {
	glog := gamelog.GameLog{}
	last_write := time.Now()
	
	for {
		response := <- input

		for _, game := range response.Games {
			// Only keep games that are matched 5v5 (no bot games, etc).
			if 	(game.GameMode != "CLASSIC") || (game.GameType != "MATCHED_GAME") || (strings.Contains(game.GameSubType, "5x5")) {
				continue
			}
			
			record := gamelog.GameRecord{}
			
			record.Timestamp = gproto.Uint64(game.CreateDate)
			record.GameId = gproto.Uint64(game.GameId)
			
			team1 := gamelog.Team{}
			team2 := gamelog.Team{}

			// Add the target player and set the outcome.
			plyr := gamelog.Player{}
			plyr.SummonerId = gproto.Uint64(response.SummonerId)

			if game.TeamId == 100 {
				pstats := gamelog.PlayerStats{}
				pstats.Player = &plyr
				team1.Players = append(team1.Players, &pstats)
					
				team1.Victory = gproto.Bool(game.Stats.Win)
				team2.Victory = gproto.Bool(!game.Stats.Win)
			} else if game.TeamId == 200 {
				pstats := gamelog.PlayerStats{}
				pstats.Player = &plyr
				team2.Players = append(team2.Players, &pstats)
								
				team1.Victory = gproto.Bool(!game.Stats.Win)
				team2.Victory = gproto.Bool(game.Stats.Win)
			} else {
				log.Println("Unknown team ID found on game", game.GameId)
			}
			
			// Add all fellow players.
			for _, player := range game.FellowPlayers {
				plyr := gamelog.Player{}
				pstats := gamelog.PlayerStats{}

				plyr.SummonerId = gproto.Uint64(player.SummonerId)
				
				if player.TeamId == 100 {
					pstats.Player = &plyr
					team1.Players = append(team1.Players, &pstats)
				} else if player.TeamId == 200 {
					pstats.Player = &plyr
					team2.Players = append(team2.Players, &pstats)
				} else {
					log.Println("Unknown team ID found on game", game.GameId)
				}
			}
			
			record.Teams = append(record.Teams, &team1, &team2)
			
			glog.Games = append(glog.Games, &record)
		}
		
		// Every hour, write out the gamelog and clear memory.
		if time.Now().Sub(last_write).Minutes() >= 20.0 {
			write_out(glog)
			glog = gamelog.GameLog{}
			
			last_write = time.Now()
		}
	}
}

func write_out(glog gamelog.GameLog) {
	data, _ := gproto.Marshal(&glog)
	
	t := time.Now()
	filename := fmt.Sprintf("logs/%s.gamelog", t.Local().Format("2006-01-02:15.04"))
	
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Panic("Cannot write file:", err)
	}
	log.Println(fmt.Sprintf("%d games written to %s", len(glog.Games), filename))
}
