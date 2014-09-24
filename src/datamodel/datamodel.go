package datamodel

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"strconv"
	"time"
)

// Create a single sessions with mongo that'll be shared.
var session, serr = mgo.Dial("request.loltracker.com:27017")


type Retriever interface {
	init()
}

/**
 * CRUD operations on individual summoners.
 */
type SummonerRetriever struct {
	collection	*mgo.Collection
}

/**
 * CRUD operatoins on individual games.
 */
type GameRetriever struct {
	collection	*mgo.Collection
}

type SummonerMetadataRetriever struct {
	collection *mgo.Collection
}

/**
 * Retrieval operations on collections of games and summoners.
 */
//type UniverseRetriever struct {
//	initialized	bool
//}


type LoLRetriever struct {
//	universe	*UniverseRetriever
	games		GameRetriever
	summoners	SummonerRetriever
	summoner_md	SummonerMetadataRetriever

	initialized	bool
}

/**
 * Iterator structs and methods.
 */
type SummonerIter struct {
	queue		chan SummonerRecord
	initialized	bool
	reached_end	bool
}

func (i *SummonerIter) Init() {
	if i.initialized {
		return
	}

	i.queue = make(chan SummonerRecord, 20)
	i.initialized = true
	i.reached_end = false
}

func (i *SummonerIter) HasNext() bool {
	return !i.reached_end
}

func (i *SummonerIter) Next() SummonerRecord {
	i.Init()
	
	// Once we reach the end there's nothing else to send.
	if i.reached_end {
		return SummonerRecord{ SummonerId: 0 }
	}
	
	s := <- i.queue
	
	if s.SummonerId == 0 {
		i.reached_end = true
	}

	return s
}

type GameIter struct {
	queue		chan GameRecord
	initialized	bool
	reached_end	bool
}

func (i *GameIter) Init() {
	if i.initialized {
		return
	}

	i.queue = make(chan GameRecord, 20)
	i.initialized = true
	i.reached_end = false	
}

func (i *GameIter) HasNext() bool {
	return !i.reached_end
}

func (i *GameIter) Next() GameRecord {
	i.Init()
	
	// Once we reach the end there's nothing else to send.
	if i.reached_end {
		return GameRecord{ GameId: 0 }
	}
	
	g := <- i.queue
	
	if g.GameId == 0 {
		i.reached_end = true
	}

	return g
}

/**
 *
 */
func (r *LoLRetriever) init() {
	if r.initialized {
		return
	}

	r.games.collection = session.DB("lolstat").C("games")
	r.summoners.collection = session.DB("lolstat").C("summoners")
	r.summoner_md.collection = session.DB("lolstat").C("summonermd")

	// Mark the retriever as initialized.
	r.initialized = true
}

/*********************
 *** Universe CRUD ***
 ********************/

func (r *LoLRetriever) CountKnownSummoners() int {
	r.init()
	
	query := r.summoners.collection.Find( bson.M{} )
	count, _ := query.Count()
	
	return count
}

func (r *LoLRetriever) GetAllSummonersIter() SummonerIter {
	r.init()

	iter := SummonerIter{}
	iter.Init()

	go func() {
		dedup := make(map[uint32]bool)
		query_iter := r.games.collection.Find( bson.M{} ).Iter()

		record := GameRecord{}
		// Loop through all game records and find unique summoner ID's.
		for query_iter.Next(&record) {
			for _, recorded_team := range record.Teams {
				for _, recorded_player := range recorded_team.Players {
					summoner_id := recorded_player.Player.SummonerId
					if _, exists := dedup[summoner_id]; !exists {
						dedup[summoner_id] = true
						summ, exists := r.GetSummoner(summoner_id)

						// If the summoner doesn't yet exist, create a shell.
						if !exists {
							summ = SummonerRecord{}
							summ.SummonerId = summoner_id
						}
						iter.queue <- summ
					}
				}
			}
		}
		
		// Send the "null" after we've made it through the full list.
		iter.queue <- SummonerRecord{ SummonerId: 0 }
	} ();

	return iter
}

func (r *LoLRetriever) GetKnownSummonersIter() SummonerIter {
	r.init()

	iter := SummonerIter{}
	iter.Init()

	go func() {
		query_iter := r.summoners.collection.Find( bson.M{} ).Iter()

		summoner := SummonerRecord{}
		for query_iter.Next(&summoner) {
			iter.queue <- summoner
		}
		
		iter.queue <- SummonerRecord{ SummonerId: 0 }
	} ();

	return iter
}

func (r *LoLRetriever) GetQuickdateGamesIter(quickdate string) GameIter {
	r.init()

	iter := GameIter{}
	iter.Init()

	go func() {
		start, _ := time.Parse("2006-01-02", quickdate)
		start_str, _ := strconv.Atoi( start.Format("20060102") )

		// TODO: this needs to use the same timestamp format as what's being
		// stored, which is a millisecond-based UNIX timestamp.	
		// 'q' is the database-side name for the "quickdate" field.
		query_iter := r.games.collection.Find( bson.M{"q": start_str} ).Iter()

		game := GameRecord{}
		for query_iter.Next(&game) {
			iter.queue <- game
		}
		
		iter.queue <- GameRecord{ GameId: 0 }
	} ();
	
	return iter
}

func (r *LoLRetriever) GetGameIter() GameIter {
	r.init()

	iter := GameIter{}
	iter.Init()

	go func() {
		query_iter := r.games.collection.Find( bson.M{} ).Iter()

		game := GameRecord{}
		for query_iter.Next(&game) {
			iter.queue <- game
		}

		iter.queue <- GameRecord{ GameId: 0 }
	} ();

	return iter
}

/*****************
 *** Game CRUD ***
 *****************/



/**
 * Looks up a game object and returns it. Also returns whether
 * the game was found or not with a boolean.
 */
func (r *LoLRetriever) GetGame(gameId uint64) (GameRecord, bool) {
	r.init()
	
	query := r.games.collection.Find( bson.M{ "_id": gameId } )
	count, _ := query.Count()
	
	if count == 0 {
		return GameRecord{}, false
	} else if count > 1 {
		log.Println("WARNING: more than one game found for GameId", gameId)
		return GameRecord{}, false
	} else {
		record := GameRecord{}
		query.One(&record)
		
		return record, true
	}
}

/**
 * StoreGame will either update an existing record or create a new one, depending
 * on whether the game ID already exists in the database.
 */
func (r *LoLRetriever) StoreGame(gr *GameRecord) {
	r.init()
	
	_, exists := r.GetGame(gr.GameId)
	
	if exists {
		r.games.collection.Update(bson.M{"_id": gr.GameId}, gr)
	} else {
		r.games.collection.Insert(gr)
	}	
}


func (r *LoLRetriever) RemoveGame(gr *GameRecord) {
	r.init()
	
	r.games.collection.Remove(gr)
}

/*********************
 *** Summoner CRUD ***
 ********************/

/**
 *
 * If the summoner can't be found then an initialized summoner object is returned
 * with a SummonerId of 0.
 */
func (r *LoLRetriever) GetSummoner(sid uint32) (SummonerRecord, bool) {
	r.init()
	
	query := r.summoners.collection.Find( bson.M{ "_id": sid } )	
	num_summoners, _ := query.Count()
	
	if num_summoners == 0 {
		return SummonerRecord{}, false
	} else if num_summoners > 1 {
		log.Println("WARNING: more than one summoner found for #", sid)
		return SummonerRecord{}, false
	} else {
		summoner := SummonerRecord{}
		query.One(&summoner)
		
		// Check to see if there's a metadata record for this summoner
		// and fetch + join it to the summoner record if so. If the two
		// record have already been normalized then we can skip this step
		// and save ourselves a query.
		
		// TODO: is there a better way to check for existence?	
		empty := SummonerMetadata{}
		if summoner.Metadata == empty {
			smd, exists := r.getSummonerMetadata(summoner.SummonerId)
			if exists {
				summoner.Metadata = smd
			}
		}
		
		return summoner, true
	}
}

func (r *LoLRetriever) StoreSummoner(summoner *SummonerRecord) {
	// Store primary data struct in summoners collection
	// Store name in summonerdata collection
	r.init()
	summoner.LastUpdated = (uint64)(time.Now().Unix())
	
	_, exists := r.GetSummoner(summoner.SummonerId)
	
	if exists {
		r.summoners.collection.Update(bson.M{"_id": summoner.SummonerId}, summoner)
	} else {
		r.summoners.collection.Insert(summoner)
	}
	
	// Also update the SummonerMetadata record if it exists. Note that
	// this write will currently occur whenever the record exists; it
	// does not check to see if there have been any changes.
	empty := SummonerMetadata{}
	if summoner.Metadata != empty {
		r.storeSummonerMetadata(summoner)
	}
}

/**
 * Fetch the metadata for the provided summoner, which may include the 
 * summoner's name.
 * 
 * The method can only be called from within the module. Metadata records
 * will automatically be joined with summoner records that come from
 * GetSummoner(). Note that this does NOT currently happen for bulk
 * requests at the moment.
 */
func (r *LoLRetriever) getSummonerMetadata(sid uint32) (SummonerMetadata, bool) {
	r.init()
	
	query := r.summoner_md.collection.Find( bson.M{ "_id": sid} )
	count, _ := query.Count()
	
	if count == 0 {
		return SummonerMetadata{}, false
	} else if count > 1 {
		log.Println("WARNING: more than one metadata record found for #", sid)
		return SummonerMetadata{}, false
	} else {
		smd := SummonerMetadata{}
		query.One(&smd)
		
		return smd, true
	}
}

func (r *LoLRetriever) storeSummonerMetadata(summ *SummonerRecord) {
	r.init()
	
	summ.Metadata.SummonerId = summ.SummonerId
	_, exists := r.getSummonerMetadata(summ.SummonerId)
	
	if exists {
		r.summoner_md.collection.Update(bson.M{"_id": summ.SummonerId}, summ.Metadata)
	} else {
		r.summoner_md.collection.Insert(summ.Metadata)
	}
}
