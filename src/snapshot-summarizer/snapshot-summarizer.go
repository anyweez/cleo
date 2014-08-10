package main

/**
 * This program reads in all snapshots for a given summoner and computes
 * summaries for them. It computes summaries for all snapshots for a 
 * provided summoner ID.
 */ 

import (
	"flags"
	"proto"
	"snapshot"
)

var SUMMONER_FILE = flag.String("summoners", "", "The filename containing a list of summoner ID's.")

/**
 * Goroutine that handles all of the computation for a summoner's
 * snapshots.
 */
func handle_summoner(sid uint32, running chan bool) {
	retriever := snapshot.Retriever{}
	retriever.Init()

	snapshots := retriever.GetSnapshots(sid)
	
	// Update each snapshot with new computations.
	for _, snapshot := range snapshots {
		for _, comp := snapshot.Computations {
			sv := proto.StatsValue{}
			sv.Name, sv.Absolute, sv.Normalized = comp(snapshot)
			
			snapshot.Stats = append(snapshot.Stats, sv)
		}
		retriever.UpdateSnapshot(snapshot)
	} 
	
	running <- true
}

func main() {
	flags.Parse()
	
	// Read in the summoner ID's that this process is responsible for.
	sids := snapshot.ReadSummonerIds(*SUMMONER_FILE)
	running := make(chan bool)
	
	// For each ID, kick off a goroutine that will retrieve all related
	// snapshots.
	for _, sid := range sids {
		go handle_summoner(sid, running)
	}
	
	// Wait until we get the same number of completion signals as we
	// have goroutines running.
	for _, sid := range sids {
		<- running
	}
}
