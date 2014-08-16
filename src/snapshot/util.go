package snapshot

/**
 * This file contains a set of utility functions used throughout the analysis.
 */

import (
	"bufio"
	"os"
	"strconv"
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
func ConvertTimestamp(date string) uint64 {
	val, _ := strconv.Atoi(strings.Replace(date, "-", "", -1))

	return (uint64)(val)
}
