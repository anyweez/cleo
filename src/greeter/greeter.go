package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"snapshot"
	"time"
)

/**
 * This program looks up summoner names and appends them to summoner records
 * so that the data can be used in other processes (nameserver, for example).
 *
 * It depends on the Riot API for looking up names.
 */

var API_KEY = flag.String("apikey", "", "Riot API key")

func update(who snapshot.SummonerRecord, retriever snapshot.Retriever) {
	log.Println("Looking up name for summoner #", who.SummonerId)
	url := "https://na.api.pvp.net/api/lol/na/v1.4/summoner/%d/name?api_key=%s"

	resp, err := http.Get(fmt.Sprintf(url, who.SummonerId, *API_KEY))
        response := make(map[string]string)

        if err != nil {
                log.Println("Error retrieving data:", err)
        } else {
                defer resp.Body.Close()
                body, _ := ioutil.ReadAll(resp.Body)

                json.Unmarshal(body, &response)

		for k, v := range response {
			who.SummonerName = v
		}
	}

	retriever.UpdateSnapshot(&who)
}


func main() {
	flag.Parse()

	retriever := snapshot.Retriever{}
	retriever.Init()

	for {
		// Fetch all summoners.
		iter := retriever.GetSnapshotsIter()
		result := snapshot.SummonerRecord{}
		for iter.Next(&result) {
			// If the summoner name is not set, let's look it up.
			if len(result.SummonerName) == 0 {
				go update(result, retriever)
				time.Sleep(1100 * time.Millisecond)
			}
		}

		// After completing a loop, wait for a bit. This is primarily
		// to keep this loop from sending too many queries to the backend
		// when the result set is small or empty.
		time.Sleep(10 * time.Second)
	}
}
