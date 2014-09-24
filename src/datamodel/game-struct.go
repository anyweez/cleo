package datamodel

const (
	LEAGUETYPE_BRONZE   = iota
	LEAGUETYPE_SILVER   = iota
	LEAGUETYPE_GOLD     = iota
	LEAGUETYPE_PLATINUM = iota
	LEAGUETYPE_DIAMOND  = iota
)

type LeagueType struct {
	Division uint32 `bson:"d"`
	Tier     uint32 `bson:"t"`
	Queue    string `bson:"q"`
}

type GameRecord struct {
	// The number of players that have been merged into this record (1 => 10)
	MergeCount uint32 `bson:"m"`
	Timestamp  uint64 `bson:"t"`
	Duration   uint32 `bson:"d"`
	QuickDate  uint32 `bson:"q"`
	GameId     uint64 `json:"id" bson:"_id"`

	Teams []*Team `bson:"e"`
}

type Team struct {
	Victory bool           `bson:"v"`
	Players []*PlayerStats `bson:"p"`
}

type PlayerStats struct {
	IsSet      bool        `bson:"s"`
	Player     *PlayerType `bson:"p"`
	Kills      uint32      `bson:"k"`
	Deaths     uint32      `bson:"d"`
	Assists    uint32      `bson:"a"`
	GoldEarned uint32      `bson:"g"`
	Minions    uint32      `bson:"m"`

	// champion from champions.go
	Champion uint32 `bson:"c"`
}

type PlayerRank struct {
	Level  uint32 `bson:"l"`
	League int    `bson:"e"`
}

type PlayerType struct {
	Name       string     `bson:"n"`
	SummonerId uint32     `bson:"s"`
	Ranking    PlayerRank `bson:"r"`
}
