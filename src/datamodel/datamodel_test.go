package datamodel

import (
	"log"
	"math/rand"
	"testing"
)

// TODO: Need to add a truth check to the iterator tests to make sure everything's being fetched.

/**
 * Test to ensure that GetKnownSummonersIter() returns a list of unique
 * summoner ID's. This test also prints the number of records returned,
 * which can be manually compared to the count() of the summoners collection. 
 */
func TestGetKnownSummonersIter(t *testing.T) {
	retriever := LoLRetriever{}
	iter := retriever.GetKnownSummonersIter()

	dedup := make(map[uint32]bool)
	count := 0

	for iter.HasNext() {
		summoner := iter.Next()
		
		// Check to make sure no duplicates show up.
		if _, exists := dedup[summoner.SummonerId]; exists {
			t.Error("Duplicate summoner ID detected.")
		} else {
			dedup[summoner.SummonerId] = true
		}
		
		count += 1
	}
	
	log.Println("count(KnownSummoners) =", count)
}

func TestGetAllSummonersIter(t *testing.T) {
	retriever := LoLRetriever{}
	iter := retriever.GetAllSummonersIter()
	
	dedup := make(map[uint32]bool)
	count := 0
	
	for iter.HasNext() {
		summoner := iter.Next()

		// Check to make sure no duplicates show up.
		if _, exists := dedup[summoner.SummonerId]; exists {
			t.Error("Duplicate summoner ID detected.")
		} else {
			dedup[summoner.SummonerId] = true
		}
		
		count += 1		
	}
	
	log.Println("count(AllSummoners) =", count)
}

func TestGetGame(t *testing.T) {
	retriever := LoLRetriever{}
	
	_, exists := retriever.GetGame(1)
	
	if exists {
		t.Error("Game ID 1 was said to exist; very unlikely.")
	}

	_, exists = retriever.GetGame(1544951968)
	
	if !exists {
		t.Error("Couldn't find Game ID 1544951968 which is expected to exist")
	}
}

func TestAddRemoveGame(t *testing.T) {
	retriever := LoLRetriever{}
	
	var gameid uint64 = 1
	_, exists := retriever.GetGame(gameid)
	
	for exists {
		gameid = (uint64)(rand.Uint32() % 100000)
		_, exists = retriever.GetGame(gameid)
	}

	// We now have a gameid that doesn't exist yet.
	gr := GameRecord{ GameId: gameid }
	// Add it to the collection.
	retriever.StoreGame(&gr)
	
	// Confirm that it's there.
	_, exists = retriever.GetGame(gr.GameId)
	
	if !exists {
		t.Error("Couldn't retrieve added game.")
	}
	
	// Remove it.
	retriever.RemoveGame(&gr)
}
