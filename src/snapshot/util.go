package snapshot

/**
 * This file contains a set of utility functions used throughout the analysis.
 */

import (
	"bufio"
	"os"
	"strconv"
	"time"
)

/**
 * Reads in a list of summoner ID's. Summoner ID's are encoded as strings
 * in the file but need to be translated into libcleo.GameId's.
 */
func ReadSummonerIds(filename string) []uint32 {
	sids := make([]uint32, 0, 100)

	fp, _ := os.Open(filename)
	scanner := bufio.NewScanner(fp)

	// Read and convert the strings into summoner ID's
	for scanner.Scan() {
		sid, _ := strconv.Atoi(scanner.Text())
		sids = append(sids, (uint32)(sid))
	}

	return sids
}

/**
 * Convert string-based timestamp into an appropriate integer format.
 */
func ConvertTimestamp(date string) (uint64, uint64) {
	start, _ := time.Parse("2006-01-02", date)
	end_duration, _ := time.ParseDuration("23h59m59s")
	end := start.Add(end_duration)

	// The timestamps in the database are in milliseconds, not seconds.
	return (uint64)(start.Unix() * 1000), (uint64)(end.Unix() * 1000)
}
