package snapshot

import (
	data "datamodel"
)
// TODO: Handle return values of NaN correctly.

type SnapshotFunction func(snapshot data.PlayerSnapshot, games []*data.GameRecord) (string, data.Metric)

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
func kda(snapshot data.PlayerSnapshot, games []*data.GameRecord) (string, data.Metric) {
	var num_kills uint32 = 0
	var num_deaths uint32 = 0
	var num_assists uint32 = 0

	for _, game := range games {
		for _, team := range game.Teams {
			for _, player := range team.Players {
				if snapshot.SummonerId == player.Player.SummonerId && player.IsSet {
					num_kills += player.Kills
					num_deaths += player.Deaths
					num_assists += player.Assists
				}
			}
		}
	}

	if num_deaths > 0 {
		return "kda", data.SimpleNumberMetric{ (float64)(num_kills + num_assists) / (float64)(num_deaths) }
	} else {
		return "kda", data.SimpleNumberMetric{}
	}
}

/**
 * Computes the mean # of minion kills for a given snapshot.
 */
func minionKills(snapshot data.PlayerSnapshot, games []*data.GameRecord) (string, data.Metric) {
	var num_minions uint32 = 0
	var num_set_games = 0

	for _, game := range games {
		for _, team := range game.Teams {
			for _, player := range team.Players {
				if snapshot.SummonerId == player.Player.SummonerId && player.IsSet {
					num_minions += player.Minions
					num_set_games += 1
				}
			}
		}
	}
	if num_set_games > 0 && len(games) > 0 {
		return "minionKills", data.SimpleNumberMetric{(float64)(num_minions) / (float64)(len(games)) }
	} else {
		return "minionKills", data.SimpleNumberMetric{}
	}
}
