package main

type JSONResponse struct {
        Games           []JSONGameResponse `json:"games"`
        SummonerId      uint32
}

type JSONGameResponse struct {
        FellowPlayers []JSONPlayerResponse
        Stats         JSONGameStatsResponse

        GameId     uint64
        CreateDate uint64

        TeamId     uint32
        ChampionId uint32

        GameMode    string
        GameType    string
        GameSubType string
}

type JSONGameStatsResponse struct {
	Assists				uint32
        ChampionsKilled                 uint32
        DamageDealtPlayer               uint32
        Gold                            uint32
        GoldEarned                      uint32
        GoldSpent                       uint32
        Level                           uint32
        MagicDamageDealtPlayer          uint32
        MinionsKilled                   uint32
        MinionsDenied                   uint32
        NeutralMinionsKilled            uint32
        NeutralMinionsKilledEnemyJungle uint32
        NeutralMinionsKilledYourJungle  uint32
        NumDeaths                       uint32
        NumItemsBought                  uint32
        PhysicalDamageDealtPlayer       uint32
        SightWardsBought                uint32
        SuperMonstersKilled             uint32
        TimePlayed                      uint32
        TotalDamageDealt                uint32
        TotalDamageDealtToChampions     uint32
        TotalDamageTaken                uint32
        TurretsKilled                   uint32
        VisionWardsBought               uint32
        WardsKilled                     uint32
        WardPlaced                      uint32
        Win                             bool
}

type JSONPlayerResponse struct {
        SummonerId uint32
        TeamId     uint32
        ChampionId uint32
}

type JSONLeagueResponse struct {
        Queue           string
        ParticipantId   string
        Entries         []JSONLeagueEntryResponse
        Tier            string
}

type JSONLeagueEntryResponse struct {
        LeaguePoints    uint32
        Division        string
        PlayerOrTeamId  string
}
