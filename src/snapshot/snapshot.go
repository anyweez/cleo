package snapshot

import (
	"gamelog"
)

// TODO: Handle return values of NaN correctly.

type SnapshotFunction func(snapshot *PlayerSnapshot, games []*gamelog.GameRecord) (string, float64, uint32)

// List containing a bunch of pointers to functions. Each function will
// be called on snapshots to generate a stat.
var Computations = make([]SnapshotFunction, 0, 10)

func init() {
	Computations = append(Computations, kda)
	Computations = append(Computations, minionKills)
}

/**
 * Computes the mean KDA for a given snapshot.
 */
func kda(snapshot *PlayerSnapshot, games []*gamelog.GameRecord) (string, float64, uint32) {
	var num_kills uint32 = 0
	var num_deaths uint32 = 0
	var num_assists uint32 = 0
	
	for _, game := range games {
		for _, team := range game.Teams {
			for _, player := range team.Players {
				if snapshot.SummonerId == player.Player.SummonerId {
					num_kills += player.Kills
					num_deaths += player.Deaths
					num_assists += player.Assists
				}
			}
		}
	}
	
	return "kda", (float64)(num_kills + num_assists) / (float64)(num_deaths), 0
}

/**
 * Computes the mean # of minion kills for a given snapshot.
 */
func minionKills(snapshot *PlayerSnapshot, games []*gamelog.GameRecord) (string, float64, uint32) {
	var num_minions uint32 = 0

	for _, game := range games {
		for _, team := range game.Teams {
			for _, player := range team.Players {
				if snapshot.SummonerId == player.Player.SummonerId {
					num_minions += player.Minions
				}
			}
		}
	}

	return "minionKills", (float64)(num_minions) / (float64)(len(games)), 0
}
