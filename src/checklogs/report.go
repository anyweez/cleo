package main

import (
	"io/ioutil"
	"logger"
	"strings"
)

type ReportData struct {
	Name	string
	Data	map[int]float32
}

/**
 * Create a mapping from minutes to the number of queries that occurred during that minute.
 */
func get_qps(messages chan logger.LoLLogEvent, output chan ReportData) {
	report := ReportData{}
	report.Data = make(map[int]float32, 0, 100)

	for msg := range messages {
		// Check that this log entry corresponds to an API-based event type. Note that
		// we currently don't check to see if the opreation succeeded or not.
		if msg.Operation == logger.FETCH_GAME_STATS {
			timestamp := msg.Timestamp.Round(time.Minute).Unix()
			report.Data[timestamp] += 1
		}
	}

	output <- report
}

/**
 * This function generes an HTML report thatgibves a bunch of
 * stats about internal system performance.
 */
func build_report(messages chan logger.LoLLogEvent) {
	elements := make(chan ReportData)
	num_reports := 0

	// Open a channel for QPS counting and start a goroutine to read from the channel.
	// build_report will fan out events to all of the processes required to build the
	// report.
	qps_channel := make(chan logger.LoLLogEvent, 10000)
	go get_qps(qps_channel, elements)
	num_reports += 1

	for event := range message {
		qps_channel <- event
	}
	close(qps_channel)

	// Collect reports together again, then render them all once all have completed.
	reports := make(ReportData, num_reports)
	for i := 0; i < num_reports; i++ {
		reports[i] = <- elements
	}

	render(reports, "report.html")
}

func render(reports []ReportData, outfile string) {
	html := make([]string, 0, 10)

	html = append(html, "<html><head></head><body>")
	for i := 0;i < len(reports); i++ {
		if reports[i].Name == "qps" {
			html = append(html, render_qps(reports[i]))
		}
	}

	ioutil.WriteFile(outfile, strings.Join(html, ""), 0444)
}

func render_qps(report ReportData) string  {

}
