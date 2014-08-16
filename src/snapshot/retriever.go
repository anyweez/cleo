package snapshot

/**
 * Retriever is a class that makes it easier to read and write data from
 * Mongo; it acts as an abstraction layer that separates schema from
 * business logic.
 */

import (
	// 	"labix.org/v2/mgo"
	//	"labix.org/v2/mgo/bson"
	"proto"
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

type Retriever struct {
}

func (r *Retriever) Init() {

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
func (r *Retriever) GetGames(date_str string) []proto.GameRecord {
	games := make([]proto.GameRecord, 0, 100)
	date := ConvertTimestamp(date_str)

	query := r.games_collection.Find(bson.M{"timestamp": date})
	result_iter := query.Iter()

	result := libcleo.RecordContainer{}
	for result_iter.Next(&result) {
		game := proto.GameRecord{}
		gproto.Unmarshal(result.GameData, &game)

		games = append(games, game)
	}

	return games
}

/**
 * Add a game to the game log.
 */
func (r *Retriever) SaveGame(record *proto.GameRecord) {

}

/*******************
 **** SNAPSHOTS ****
 *******************/

func (r *Retriever) GetSnapshots(sid uint32) {

}

func (r *Retriever) SaveSnapshot(snapshot *proto.PlayerSnapshot) {

}

/**
 * Overwrites the existing snapshot with a newly provided one.
 */
func (r *Retriever) UpdateSnapshot(snapshot *proto.PlayerSnapshot) {

}
