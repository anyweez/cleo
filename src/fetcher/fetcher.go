package fetcher

import gproto "code.google.com/p/goprotobuf/proto"
import "gamelog"

//import "encoding/json"
import "net/http"
import "fmt"

const API_KEY = "abebd3e9-00f2-4ba6-997d-0008c2072373"
const NUM_RETRIEVERS = 5

func main() {
	user_queue := make(chan *gamelog.Player)
	retrieval_inputs := make(chan *gamelog.Player)
	retrieval_outputs := make(chan *gamelog.GameRecord)

	// Kick off some retrievers that will pull from the retrieval queue.
	for i := 0; i < NUM_RETRIEVERS; i++ {
		go retriever(retrieval_inputs, retrieval_outputs)
	}

	// Launch a record writer that will periodically write out
	go record_writer(retrieval_outputs)

	// Load in a file full of summoner ID's.
	player := gamelog.Player{}
	player.Name = gproto.String("Brigado")
	user_queue <- &player
	// Forever: pull an summoner ID from user_queue, toss it in retrieval_inputs
	// and add it back to user_queue.

	for {
		// Wait for 1.25 seconds
		player := <-user_queue

		retrieval_inputs <- player
		user_queue <- player
	}
}

func retriever(input chan *gamelog.Player, output chan *gamelog.GameRecord) {
	url := "https://prod.api.pvp.net/api/lol/na/v1.3/summoner/by-name/%s?api_key=%s"

	for {
		player := <-input

		fmt.Println("Retrieving data for %s...", player.Name)

		resp, err := http.Get(fmt.Sprintf(url, player.Name, API_KEY))
		if err != nil {
			fmt.Println(resp.Body)
		} else {
			fmt.Println("Error retrieving data")
			fmt.Println(err)
		}
	}
}

func record_writer(input chan *gamelog.GameRecord) {
	for {
		record := <-input

		fmt.Println(record)
	}
}
