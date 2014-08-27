package snapshot

import (
)

type SnapshotFunction func(snapshot *PlayerSnapshot) (string, float64, uint32)

// List containing a bunch of pointers to functions. Each function will
// be called on snapshots to generate a
// TODO: this needs to be a list of functions
var Computations = make([]SnapshotFunction, 0, 10)

func init() {
	Computations = append(Computations, kda)
	Computations = append(Computations, minionKills)
}

/**
 * Computes the mean KDA for a given snapshot.
 */
func kda(snapshot *PlayerSnapshot) (string, float64, uint32) {
	return "kda", 10, 50
}

/**
 * Computes the mean # of minion kills for a given snapshot.
 */
func minionKills(snapshot *PlayerSnapshot) (string, float64, uint32) {
	return "minionKills", 200, 15
}

/*
interface SnapshotComputation {

}
* */
