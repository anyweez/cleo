package main

import (
	data "datamodel"
	"flag"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
)

/**
 * This program looks up summoner names and appends them to summoner records
 * so that the data can be used in other processes (nameserver, for example).
 *
 * It depends on the Riot API for looking up names.
 */

var API_KEY = flag.String("apikey", "", "Riot API key")

func update(who *data.SummonerRecord, retriever *data.LoLRetriever) {
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

		for _, v := range response {
			who.SummonerName = v
		}
	}

	retriever.StoreSummoner(who)
}

func main() {
	flag.Parse()

	retriever := data.LoLRetriever{}

	for {
		summoners_iter := retriever.GetAllSummonersIter()
		
		for summoners_iter.HasNext() {
			summoner := summoners_iter.Next()
	
			for summoner.SummonerId != 0 {
				// If the summoner name is not set, let's look it up.
                if len(summoner.SummonerName) == 0 {
					go update(&summoner, &retriever)
                    time.Sleep(1100 * time.Millisecond)
				}

				summoner = summoners_iter.Next()
			}
		}

		// After completing a loop, wait for a bit. This is primarily
		// to keep this loop from sending too many queries to the backend
		// when the result set is small or empty.
		time.Sleep(10 * time.Second)
	}
}
