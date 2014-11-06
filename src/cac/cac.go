package main

/**
 * The "command and control" binary that loads all of the distributed tasks into
 * thei beanstalk queue. The tasks that are generated are based on the flags that
 * are sent. 
 */

import (
	gproto "code.google.com/p/goprotobuf/proto"
	data "datamodel"
	"flag"
	"fmt"
	beanstalk "github.com/iwanbk/gobeanstalk"
	"log"
	"lolutil"
	"math"
	"proto"
	"time"
)

/**
 * Command and control application that populates the beanstalk job queue
 * for join-summmoners workers.
 */

var (
	SUMMONER_FILE 	= flag.String("summoners", "", "The file containing the list of summoners to handle.")
	MAX_PER_NODE  	= flag.Int("max_node", 100, "The maximum number of summoners that should be directed to a single worker")
	LABEL			= flag.String("label", "daily", "")
	START_DATE		= flag.String("target_date", "", "The specific date to be analyzed or a date from within the range to be analyzed.")
)

func daysIn(m time.Month, year int) int { 
    // This is equivalent to time.daysIn(m, year). 
    return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day() 
} 
 
func getDates(label string, date_string string) []string {
	date, _ := time.Parse("2006-01-02", date_string)
	var dates []string

	// Set the number of strings that we want for a given label
	if label == "daily" {
		// Add a single time string
		dates = append(dates, date_string)
	} else if label == "weekly" {
		// Find the start date of the week first.
		week_start := date
		for week_start.Weekday() != time.Sunday {
			week_start = week_start.AddDate(0, 0, -1)
		}

		// Then count seven dates.
		for i := 0; i < 7; i++ {
			dates = append(dates, week_start.Format("2006-01-02")) 
			week_start = week_start.AddDate(0, 0, 1)
		}
	} else if label == "monthly" {
                // Find the start date of the month first.
                month_start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, nil)

                // Then count seven dates.
                for i := 0; i < daysIn(month_start.Month(), month_start.Year()); i++ {
                        dates = append(dates, month_start.Format("2006-01-02")) 
			month_start = month_start.AddDate(0, 0, 1)
                }
	} else {
		log.Fatal("Unknown label:", label)
	}

<<<<<<< HEAD
	// Generate all of the quickdates and add them to a single slice.
	for i := 0; i < num_strings; i++ {
		dates = append(dates, date_string)

		// Parse the current time and add a day. Then reformat it as a string.
		next_date, _ := time.Parse("2006-01-02", date_string)
		next_date = next_date.Add(24 * time.Hour)
		date_string = next_date.Format("2006-01-02")
	}

=======
>>>>>>> d9ad7df3f76fb88840bdfe08fca3b05cbec52b1d
	log.Println(dates)
	return dates
}

func main() {
	flag.Parse()

	retriever := data.LoLRetriever{}
	cm := lolutil.LoadCandidates(retriever, *SUMMONER_FILE)

	log.Println("Connecting to beanstalkd...")
	bs, cerr := beanstalk.Dial("localhost:11300")
	if cerr != nil {
		log.Fatal(cerr)
	}
	log.Println("Connected.")

	for i := 0; i < (int)(math.Ceil((float64)(cm.Count())/(float64)(*MAX_PER_NODE))); i++ {
		var summoners []uint32

		// Build up the list of summoners that should be included in this
		// request.
		for j := 0; j < *MAX_PER_NODE; j++ {
			val := cm.Pop()

			if val != 0 {
				summoners = append(summoners, val)
			}
		}

		// Initialize a JoinRequest for this segment of summoners.
		jr := proto.JoinRequest{
			Label:      LABEL,
			Quickdates: getDates(*LABEL, *START_DATE),
			Summoners:  summoners,
		}

		message, _ := gproto.Marshal(&jr)

		// Send the message at priority 10, with no delay, and with a ten-minute
		// time-to-live before being returned to the queue in case of a worker
		// failure.
		id, _ := bs.Put([]byte(message), 10, 0, 10*time.Minute)
		log.Println(fmt.Sprintf("Message %d sent.", id))
	}
}
