package main

import (
	beanstalk "github.com/iwanbk/gobeanstalk"
	data "datamodel"
	"flag"
	"fmt"
	gproto "code.google.com/p/goprotobuf/proto"
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
	START_DATE		= flag.String("start_date", "", "")
)
 
func getDates(label string, date_string string) []string {
	var dates []string
	num_strings := 1

	// Set the number of strings that we want for a given label
	if label == "daily" {
		num_strings = 1
	} else if label == "weekly" {
		num_strings = 7
	} else if label == "monthly" {
		num_strings = 30
	}

	// Generate all of the quickdates and add them to a single slice.
	for i := 0; i < num_strings; i++ {
		dates = append(dates, date_string)
		
		// Parse the current time and add a day. Then reformat it as a string.
		next_date, _ := time.Parse("2006-01-02", date_string)
		next_date = next_date.Add( 24 * time.Hour )
		date_string = next_date.Format("2006-01-02")

		dates = append(dates, date_string)
	}

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

	for i := 0; i < (int) (math.Ceil( (float64)(cm.Count()) / (float64)(*MAX_PER_NODE) )); i++ {
		var summoners []uint32
		
		// Build up the list of summoners that should be included in this
		// request.
		for j := 0; j < *MAX_PER_NODE; j++ {
			val := cm.Pop()
			
			if val != 0 {
				summoners = append( summoners, val )
			}
		}

		// Initialize a JoinRequest for this segment of summoners.
		jr := proto.JoinRequest{
			Label: LABEL,
			Quickdates: getDates(*LABEL, *START_DATE),
			Summoners: summoners,
		}

		message, _ := gproto.Marshal(&jr)
		
		// Send the message at priority 10, with no delay, and with a ten-minute
		// time-to-live before being returned to the queue in case of a worker
		// failure.
		id, _ := bs.Put([]byte(message), 10, 0, 10 * time.Minute)
		log.Println( fmt.Sprintf("Message %d sent.", id) ) 
	}
}
