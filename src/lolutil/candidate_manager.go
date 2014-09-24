package lolutil

import (
	"bufio"
	data "datamodel"
	"log"
	"os"
	"strconv"
)

/**
 *  CandidateManager keeps track of a list of Players that can be fetched
 *  and ensures that they're all unique in the queue.
 */
type CandidateManager struct {
	Queue        chan uint32
	CandidateMap map[uint32]bool

	count uint32
}

func (cm *CandidateManager) Add(player uint32) {
	_, exists := cm.CandidateMap[player]
	if !exists {
		cm.CandidateMap[player] = true
		cm.count += 1

		cm.Queue <- player
	}
}

func (cm *CandidateManager) Next() uint32 {
	player := <-cm.Queue
	// Cycle the candidate back into the queue.
	cm.Queue <- player

	return player
}

func (cm *CandidateManager) Pop() uint32 {
	if cm.Count() == 0 {
		return 0
	}

	cm.count -= 1
	return <-cm.Queue
}

func (cm *CandidateManager) Count() uint32 {
	return cm.count
}

func LoadCandidates(retriever data.LoLRetriever, seedfile string) CandidateManager {
	cm := CandidateManager{}

	// Load in a file full of summoner ID's as a seed set and add it to the list
	// of already-known summoners.
	var summoner_ids []uint32

	if len(seedfile) > 0 {
		summoner_ids = read_summoner_ids(seedfile)
	}
	num_summoners := retriever.CountKnownSummoners() + len(summoner_ids)

	// TODO: move this inititalization stuff into the object.
	cm.Queue = make(chan uint32, num_summoners)
	cm.CandidateMap = make(map[uint32]bool)

	// Add summoners from the seed file into the candidate manager.
	for _, sid := range summoner_ids {
		cm.Add(sid)
	}

	// Add all of the known summoners that are in the database.
	summ := data.SummonerRecord{}
	summoner_iter := retriever.GetKnownSummonersIter()

	for summoner_iter.HasNext() {
		summ = summoner_iter.Next()

		if summ.SummonerId != 0 {
			cm.Add(summ.SummonerId)
		}
	}

	return cm
}

/**
 * This function reads in a list of champions from a local file to start
 * as the seeding set. The fetcher will automatically include new champions
 * it discovers on its journey as well.
 */
func read_summoner_ids(filename string) []uint32 {
	// Read the specified file.
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Cannot find champions file.")
	}
	defer file.Close()

	lines := make([]uint32, 0, 10000)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		value, _ := strconv.ParseUint(scanner.Text(), 10, 32)
		lines = append(lines, uint32(value))
	}

	// Return a set of summoner ID's.
	return lines
}
