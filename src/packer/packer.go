package main

// The packer takes all known game records and condenses them into a PackedChampionGameList.
// It outputs the PCGL, which is then used for searching in online queries. All of the
// game fields of the PCGL are in sorted order.

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"libcleo"
	"log"
	"net/http"
	"proto"
	"regexp"
	"strings"
	"time"
)

var API_KEY = flag.String("apikey", "", "Riot API key")

/**
 * StaticRequestInfo defines the data that should be extracted from the
 * JSON response that we get back from Riot's API.
 */
type StaticRequestInfo struct {
	Data map[string]StaticEntry
}

/**
 * StaticEntry defines what a single entry in the output JSON looks
 * like. It also acts as the receiving structure for each row in the
 * parent StaticRequestInfo, but some of the fields are mutated from
 * the value that comes in.
 */
type StaticEntry struct {
	Id        uint32 `json:"id"`
	Name      string `json:"name"`
	Shortname string `json:"shortname"`
	Title     string `json:"title"`
	Img       string `json:"img"`
	Games     uint32 `json:"games"`
}

type StaticOutputJSON struct {
	LastUpdated		int64			`json:"lastUpdated"`
	NumGames		int				`json:"numGames"`
	Champions		[]StaticEntry	`json:"champions"`
}

/**
 * This function generates static output that can be consumed by the
 * frontend based on data compiled during the packing process. Additional
 * metadata for each champion is also fetched from Riot and included in
 * the output.
 *
 * This function currently writes out championList.json, a file that's
 * consumed by frontends that includes a list of all champions and some
 * metadata about them, including how many games are included in the
 * PCGL for them.
 */
func write_statics(filename string, pcgl libcleo.LivePCGL) {
	entries := StaticRequestInfo{}

	url := "https://na.api.pvp.net/api/lol/static-data/na/v1.2/champion?&api_key=%s"
	log.Println("Requesting latest champion data from Riot...")

	// Retrieve a list of all champions according to Riot, along with
	// some core info about each (name, title, etc)
	resp, err := http.Get(fmt.Sprintf(url, *API_KEY))
	if err != nil {
		log.Println("Error retrieving data:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &entries)

	outjson := StaticOutputJSON{}
	// Export the number of games in this pcgl export.
	outjson.NumGames = len(pcgl.All)
	outjson.LastUpdated = time.Now().Unix()
	
	outjson.Champions = make([]StaticEntry, 0, 200)
	// Remove non-alphanumeric characters.
	reg, _ := regexp.Compile("[^A-Za-z0-9 ]+")

	for _, entry := range entries.Data {
		champ := libcleo.Rid2Cleo(entry.Id)
		clean_name := reg.ReplaceAllString(entry.Name, "")

		entry.Id = uint32(champ)
		// Shortname is the clean_name with spaces replaced with underscores (internally defined).
		entry.Shortname = strings.ToLower(strings.Replace(clean_name, " ", "_", -1))
		// Img path is the clean_name with spaces removed (defined by Riot).
		entry.Img = fmt.Sprintf("http://ddragon.leagueoflegends.com/cdn/4.9.1/img/champion/%s.png", strings.Replace(clean_name, " ", "", -1))
		entry.Games = uint32(len(pcgl.Champions[champ].Winning) + len(pcgl.Champions[champ].Losing))

		outjson.Champions = append(outjson.Champions, entry)
	}

	data, _ := json.Marshal(outjson)
	ioutil.WriteFile(filename, data, 0644)
	log.Println(fmt.Sprintf("Written static champion file to %s", filename))
}

func main() {
	// TODO: Make this part optional via a command line flag.
	flag.Parse()

	if *API_KEY == "" {
		log.Fatal("You must provide an API key using the -apikey flag.")
	}

	pcgl := libcleo.LivePCGL{}
	pcgl.Champions = make(map[proto.ChampionType]libcleo.LivePCGLRecord)
	pcgl.All = make([]libcleo.GameId, 0, 100)

	// Read all records from Mongo.
	session, _ := mgo.Dial("127.0.0.1:27017")
	games_collection := session.DB("lolstat").C("games")
	defer session.Close()
	log.Println("Connection to MongoDB instance established.")

	gid_map := make(map[uint64]libcleo.GameId)
	var next_gid libcleo.GameId = 0

	// For each record:
	//	- Get all champions. For each champion:
	//		- If team won, add game id to pcgl.Champions[champion].Winning
	//		- If loss, add to .Losing
	//		- In all cases add to pcgl.All
	result := libcleo.RecordContainer{}
	query := games_collection.Find(bson.M{})
	result_iter := query.Iter()
	total_count, _ := query.Count()
	current := 1

	for result_iter.Next(&result) {
		fmt.Print(fmt.Sprintf("Packing %d of %d...", current, total_count), "\r")

		game := proto.GameRecord{}
		gproto.Unmarshal(result.GameData, &game)

		// Map game ID's to something much closer to zero (and tightly
		// packed). This will make it possible to work in 32-bit land
		// at serving time until we get beyond 4B games. That's far away.
		gid, exists := gid_map[*game.GameId]
		if exists {
			game.GameId = gproto.Uint64( uint64(gid) )
		} else {
			gid_map[*game.GameId] = next_gid
			game.GameId = gproto.Uint64( uint64(next_gid) )

			next_gid += 1
		}

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
					r.Winning = append(pcgl.Champions[*player.Champion].Winning, libcleo.GameId(*game.GameId))
					// If they lost, add it to the loss pool.
				} else {
					r.Losing = append(pcgl.Champions[*player.Champion].Losing, libcleo.GameId(*game.GameId))
				}
				// Reassign to the master struct
				pcgl.Champions[*player.Champion] = r
			}
		}

		pcgl.All = append(pcgl.All, libcleo.GameId(*game.GameId))
		current += 1
	}

	// Then convert into the serializable form.
	packed_pcgl := proto.PackedChampionGameList{}

	for k, v := range pcgl.Champions {
		record := proto.PackedChampionGameList_ChampionGameList{}
		record.Champion = k.Enum()

		// Copy over casted values. They're the same type but v.Winning
		// uses an aliased type and Go doesn't recognize that they're the
		// same...
		// TODO: is there a better (faster, more memory efficient) way
		// to do this without removing the type alias?
		for _, val := range v.Winning {
			record.Winning = append(record.Winning, uint32(val))
		}
		for _, val := range v.Losing {
			record.Losing = append(record.Losing, uint32(val))
		}

		packed_pcgl.Champions = append(packed_pcgl.Champions, &record)
	}

	for _, val := range pcgl.All {
		packed_pcgl.All = append(packed_pcgl.All, uint32(val))
	}
	data, _ := gproto.Marshal(&packed_pcgl)

	// Write to file.
	err := ioutil.WriteFile("all.pcgl", data, 0644)
	if err != nil {
		log.Fatal("Could not write PCGL file.")
	} else {
		log.Println(fmt.Sprintf("Successfully wrote %d records to all.pcgl.", len(packed_pcgl.All)))
	}

	write_statics("html/static/data/metadata.json", pcgl)
}
