package main

import (
	data "datamodel"
	"flag"
	"fmt"
	loggly "github.com/go-loggly-search"
	"log"
	"logger"
	stats "github.com/GaryBoone/GoStats/stats"
	"time"
)

var (
	USERNAME = flag.String("username", "", "Loggly.com username")
	PASSWORD = flag.String("pass", "", "Loggly.com password")
	ACCOUNT  = flag.String("account", "", "Loggly.com account name")
)

type MetaLogEvent struct {
	Timestamp time.Time
	Event     *logger.LoLLogEvent
}

type FillStats struct {
	Histogram  []float32
	AvgFill    float64
	StddevFill float64
}

func (self *FillStats) pretty() string {
	var out string
	out += "[FILL]"
	out += "\nGames where MergeCount = x\n--------------------------\n"

	for i, val := range self.Histogram {
		out += fmt.Sprintf("%d: %.3f\n", i, val)
	}

	out += fmt.Sprintf("Avg fill: %.2f (stddev: %.2f)", self.AvgFill, self.StddevFill)

	return out
}

func (self *FreshnessStats) pretty() string {
	var out string
	out += "[FRESHNESS]\n"

	return out
}

type FreshnessStats struct {
	MeanLookupsPerId   float64
	StddevLookupsPerId float64
	MeanGap            float64
	StddevGap          float64
}

func getEvents() []*MetaLogEvent {
	client := loggly.NewClient(*ACCOUNT, *USERNAME, *PASSWORD)
	//      res, err := client.Query(`tag:"fetcher" AND (json.Operation:"0" OR json.Operation:"1") AND json.Outcome:0`).From("-2d").Fetch()
	res, err := client.Query(`json.Outcome:0`).From("-2d").Fetch()

	if err != nil {
		log.Fatal(err)
	}

	logs := make([]*MetaLogEvent, 1)

	// Convert the log event into an internally usable structure and add it to a slice.
	//	for _, event := range results.Events {
	//		meta := MetaLogEvent{ Timestamp: event.Timestamp, Event: event.Json.(logger.LoLLogEvent) }

	//		logs = append(logs, &meta)
	//	}

	log.Println(fmt.Sprintf("Read %d events.", res.Total))
	return logs
}

func getGameIter() data.GameIter {
	retriever := data.LoLRetriever{}
	return retriever.GetGameIter()
}

/**
 * This function generates statistics related to how dense the games collection
 * is; specifically, how many of the players in the known games are scanned.
 * Higher values o fthis should correspond with overall denser coverage of the
 * summoner graph; ideally we'd be able to reach a fill rate of 100%;
 */
func getFillStats() FillStats {
	iter := getGameIter()
	data := FillStats{}

	data.Histogram = make([]float32, 11)
	count := 0
	fullset := make([]float64, 0, 100)

	for iter.HasNext() {
		game := iter.Next()

		// Check to make sure no duplicates show up.
		if game.GameId == 0 {
			continue
		}

		data.Histogram[game.MergeCount] += 1
		fullset = append(fullset, (float64)(game.MergeCount))
		count += 1
	}

	for i := 0; i < len(data.Histogram); i++ {
		data.Histogram[i] = (float32)(data.Histogram[i]) / (float32)(count)
	}

	data.AvgFill = stats.StatsMean(fullset)
	data.StddevFill = stats.StatsSampleStandardDeviation(fullset)

	return data
}

/**
 * Freshness stats represent how regularly the snapshots of an individual summoner
 * are updated. Higher frequency of a single summoner leads to more complete data
 * on a per-summoner level but may mean that we have to reduce fill.
 */
func getFreshnessStats() FreshnessStats {
	logs := getEvents()
	stats := FreshnessStats{}

	// Get the number of checks per summoner
	summoner_count := make(map[uint64][]time.Time)
	for _, record := range logs {
		// For all successful events, count the number of times we've seen each summoner.
		if record.Event.Operation == logger.FETCH_MATCH_HISTORY && record.Event.Outcome == logger.SUCCESS {
			if _, exists := summoner_count[record.Event.Target]; !exists {
				summoner_count[record.Event.Target] = make([]time.Time, 0, 10)
			}
			summoner_count[record.Event.Target] = append(summoner_count[record.Event.Target], record.Timestamp)
		}
	}

	frequency := make([]int, 0, len(summoner_count))
	for _, times := range summoner_count {
		frequency = append(frequency, len(times))
	}

	// Compute statistics related to the # of lookups occuring per summoner.
	//	stats.MeanLookupsPerId = average(frequency)
	//	stats.StddevLookupsPerId = stddev(frequency)

	// Get the adverage duration between lookup events for a single summoner.
	gaps := make([]int64, 0, 100)
	for _, times := range summoner_count {
		//		sort.Sort(times)

		for i := 0; i < len(times)-1; i++ {
			gaps = append(gaps, (int64)(times[i-1].Sub(times[i])))
		}
	}

	// Compute the global gap between one lookup and the next.
	//	stats.MeanGap = average(gaps)
	//	stats.StddevGap = stddev(gaps)

	return stats
}

func main() {
	flag.Parse()
	log.Println("Fetching data and calculating...")

	// Use logs to find out how often we get to examine each summoner (depth)
	fresh := getFreshnessStats()
	//	fmt.Println("Frequency stats:", freq)
	fmt.Println(fresh.pretty())

	// Use the games database to determine how many summoners we typically see in a game (breadth).
	fill := getFillStats()
	fmt.Println(fill.pretty())
}
