package snapshot

/**
 * This file defines the data structure for storing snapshots.
 */

type SummonerRecord struct {
	SummonerId		uint32 					`json:"id" bson:"_id"`
	SummonerName	string						`bson:"s"`
	LastUpdated		uint64						`bson:"l"`
	Daily			map[string]*PlayerSnapshot	`bson:"d"`
}

type PlayerSnapshot struct {
	SummonerId		uint32						`bson:"s"`
	GamesList		[]uint64					`bson:"g"`
	// TODO: add rank to this snapshot

	// The relevant gameplay statistics for the period covered by this snapshot
	Stats 			[]PlayerStat				`bson:"t"`

	// When this record was generated.
	CreationTimestamp	uint64					`bson:"c"`
}

type PlayerStat struct {
	Name		string							`bson:"n"`
	Absolute	float64						`bson:"a"`
	Normalized	uint32							`bson:"o"`
}
