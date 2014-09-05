package snapshot

/**
 * Retriever is a class that makes it easier to read and write data from
 * Mongo; it acts as an abstraction layer that separates schema from
 * business logic.
 */
 import (
	"gamelog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

/**
 * The high-level schemas are as follows:
 *
 * - lolstat
 * 	- games
 *  	    Keyed by: game id
 * 		Source: Riot API
 * 		Purpose: raw storage for game data
 * 	- players
 * 		Keyed by: summoner id
 * 		Source: built from games table by join-summoners process
 * 		Purpose: summarized player stats
 */
var session, _ = mgo.Dial("127.0.0.1:27017")

type Retriever struct {
	games_collection	*mgo.Collection
	summoner_collection	*mgo.Collection
}

func (r *Retriever) Init() {
    r.games_collection = session.DB("lolstat").C("games")
	r.summoner_collection = session.DB("lolstat").C("summoners")
}

/***************
 **** GAMES ****
 ***************/

/**
 * Get all of the games relevant for generating a snapshot for the
 * provided date.
 *
 * TODO: retrieve all games played in the two weeks before this date.
 */
func (r *Retriever) GetGamesIter(date_str string) mgo.Iter {
	games := make([]gamelog.GameRecord, 0, 100)
	start, end := ConvertTimestamp(date_str)

	// TODO: this needs to use the same timestamp format as what's being
	// stored, which is a millisecond-based UNIX timestamp.	
	query := r.games_collection.Find(bson.M{"timestamp": bson.M{ "$gt" : start, "$lt": end } })
	return query.Iter()
//	result_iter := query.Iter()

/*
	result := gamelog.GameRecord{}
	for result_iter.Next(&result) {
		games = append(games, result)
	}

	return games
*/
}

/**
 * Add a game to the game log.
 */
func (r *Retriever) SaveGame(record *gamelog.GameRecord) {

}

/*******************
 **** SNAPSHOTS ****
 *******************/

func (r *Retriever) NewSummoner(sid uint32) {
	record := SummonerRecord{}

        record.SummonerId = sid
        record.Daily = make(map[string]*PlayerSnapshot)

	r.summoner_collection.Insert(record)
}

func (r *Retriever) GetSnapshots(sid uint32) {

}

func (r *Retriever) SaveSnapshot(sid uint32, subset_name string, key string, ss *PlayerSnapshot) {
	query := r.summoner_collection.Find(bson.M{"_id": sid})

	// TODO: do this at some point once it's relevant.
	if subset_name != "daily" {
		log.Fatal("Haven't implemented support for saving snapshots of type '" + subset_name + "'")
	}

	// If we didn't find a record, the player exists and we can get rolling.
	count, _ := query.Count()
	if count == 0 {
		r.NewSummoner(sid)
	}

	record := SummonerRecord{}
	r.summoner_collection.Find(bson.M{"_id": sid}).One(&record)

	record.Daily[key] = ss
	record.LastUpdated = (uint64)(time.Now().Unix())
	r.summoner_collection.Update( bson.M{"_id": record.SummonerId}, record )
}
