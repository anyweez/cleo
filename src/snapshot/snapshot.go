package snapshot

import (
	"proto"
)

// List containing a bunch of pointers to functions. Each function will
// be called on snapshots to generate a 
// TODO: this needs to be a list of functions
var Computations = make([]int, 0, 10)

/**
 * Computes the mean KDA for a given snapshot.
 */
func kda(snapshot *proto.PlayerSnapshot) (string, float64, uint32) {
	return "kda", 10, 50
}

/**
 * Computes the mean # of minion kills for a given snapshot.
 */
func minionKills(snapshot *proto.PlayerSnapshot) (string, float64, uint32) {
	return "minionKills", 200, 15
}

/*
interface SnapshotComputation {
	
}
* */
