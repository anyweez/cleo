package gamelog

const (
	LEAGUETYPE_BRONZE   = iota
	LEAGUETYPE_SILVER   = iota
	LEAGUETYPE_GOLD     = iota
	LEAGUETYPE_PLATINUM = iota
	LEAGUETYPE_DIAMOND  = iota
)

type LeagueType struct {
	Division uint32
	Tier     uint32
	Queue    string
}

type GameRecord struct {
	MergeCount uint32 // The number of players that have been merged into this record (1 => 10)
	Timestamp  uint64
	Duration   uint32
	QuickDate	uint32
	GameId     uint64 `json:"id" bson:"_id"`

	Teams []*Team
}

type Team struct {
	Victory bool
	Players []*PlayerStats
}

type PlayerStats struct {
	IsSet      bool
	Player     *PlayerType
	Kills      uint32
	Deaths     uint32
	Assists    uint32
	GoldEarned uint32
	Minions    uint32

	// champion from champions.go
	Champion uint32
}

type PlayerRank struct {
	Level  uint32
	League int
}

type PlayerType struct {
	Name       string
	SummonerId uint32
	Ranking    PlayerRank
}
