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
	"strconv"
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
var session, _ = mgo.Dial("request.loltracker.com:27017")

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
 * Note that the current implementation only works for single days and
 * gains a ton of speed from this optimization (quickdate is indexed).
 */
func (r *Retriever) GetGamesIter(date_str string) *mgo.Iter {
	start, _ := time.Parse("2006-01-02", date_str)
	start_str, _ := strconv.Atoi( start.Format("20060102") )

	// TODO: this needs to use the same timestamp format as what's being
	// stored, which is a millisecond-based UNIX timestamp.	
	query := r.games_collection.Find(bson.M{"quickdate": start_str})
	return query.Iter()
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

func (r *Retriever) GetSnapshotsIter() *mgo.Iter {
	return r.summoner_collection.Find().Iter()
}

func (r *Retriever) GetSnapshots(sid uint32) SummonerRecord {
	record := SummonerRecord{}
	r.summoner_collection.Find( bson.M{"_id":sid} ).One(&record)

	return record
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

	_, ok := record.Daily[key]
	if !ok {
		record.Daily = make(map[string]*PlayerSnapshot)
	}
	record.Daily[key] = ss
	record.LastUpdated = (uint64)(time.Now().Unix())
	r.summoner_collection.Update( bson.M{"_id": record.SummonerId}, record )
}
