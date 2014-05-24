package packer

// The packer takes all known game records and condenses them into a PackedChampionGameList.
// It outputs the PCGL, which is then used for searching in online queries.

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"libcleo"
	"proto"
)

func main() {
	pcgl := libcleo.LivePCGL{}

	// Read all records from Mongo.
	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("games")
	defer session.Close()

	// For each record:
	//	- Get all champions. For each champion:
	//		- If team won, add game id to pcgl.Champions[champion].Winning
	//		- If loss, add to .Losing
	//		- In all cases add to pcgl.All
	results := []libcleo.RecordContainer{}
	games_collection.Find( bson.M{} ).All(&results)
	
	for _, record := range results {
		game := proto.GameRecord{}
		gproto.Unmarshal(record.GameData, &game)

		
		for _, team := range game.Teams {
			for _, player := range team.Players {
				_, exists := pcgl.Champions[*player.Champion]
				
				if !exists {
					pcgl.Champions[*player.Champion] = libcleo.LivePCGLRecord{}
				}
				r := pcgl.Champions[*player.Champion]

				// If the team won, add this game to this champion's win
				// pool.
				if *team.Victory {			
					r.Winning = append(pcgl.Champions[*player.Champion].Winning, *game.GameId)
				// If they lost, add it to the loss pool.
				} else {
					r.Losing = append(pcgl.Champions[*player.Champion].Losing, *game.GameId)
				}
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
	
//	data, _ := gproto.Marshal(&packed_pcgl)
	// Write to file.
}
