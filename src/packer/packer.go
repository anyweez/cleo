package main

// The packer takes all known game records and condenses them into a PackedChampionGameList.
// It outputs the PCGL, which is then used for searching in online queries. All of the
// game fields of the PCGL are in sorted order.

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"libcleo"
	"log"
	"proto"
)

func main() {
	pcgl := libcleo.LivePCGL{}
	pcgl.Champions = make(map[proto.ChampionType]libcleo.LivePCGLRecord)
	pcgl.All = make([]uint64, 0, 100)

	// Read all records from Mongo.
	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("games")
	defer session.Close()
	log.Println("Connection to MongoDB instance established.")

	// For each record:
	//	- Get all champions. For each champion:
	//		- If team won, add game id to pcgl.Champions[champion].Winning
	//		- If loss, add to .Losing
	//		- In all cases add to pcgl.All
	results := []libcleo.RecordContainer{}
	games_collection.Find( bson.M{} ).All(&results)
	
	log.Println("Retrieved", len(results), "records. Packing...")
	for _, record := range results[:100] {
		game := proto.GameRecord{}
		gproto.Unmarshal(record.GameData, &game)
		
		for _, team := range game.Teams {
			for _, player := range team.Players {
				_, exists := pcgl.Champions[*player.Champion]
				
				if !exists {
					pcgl.Champions[*player.Champion] = libcleo.LivePCGLRecord{}
				}
				// Copy this value out. We'll need to reassign a bit later once
				// the necessary modifications have been made.
				r := pcgl.Champions[*player.Champion]

				// If the team won, add this game to this champion's win
				// pool.
				if *team.Victory {			
					r.Winning = append(pcgl.Champions[*player.Champion].Winning, *game.GameId)
				// If they lost, add it to the loss pool.
				} else {
					r.Losing = append(pcgl.Champions[*player.Champion].Losing, *game.GameId)
				}
				// Reassign to the master struct
				pcgl.Champions[*player.Champion] = r
			}
		}
		
		pcgl.All = append(pcgl.All, *game.GameId)
	}
	
	// Then convert into the serializable form.
	packed_pcgl := proto.PackedChampionGameList{}
	
	for k, v := range pcgl.Champions {
		record := proto.PackedChampionGameList_ChampionGameList{}
		record.Champion = k.Enum()
		
		record.Winning = v.Winning
		record.Losing = v.Losing
		
		packed_pcgl.Champions = append(packed_pcgl.Champions, &record)
	}
	
	packed_pcgl.All = pcgl.All	
	data, _ := gproto.Marshal(&packed_pcgl)

	// Write to file.
	err := ioutil.WriteFile("all.pcgl", data, 0644)
	if err != nil {
		log.Fatal("Could not write PCGL file.")
	} else {
		log.Println(fmt.Sprintf("Successfully wrote %d records to all.pcgl", len(packed_pcgl.All)))
	}
}
