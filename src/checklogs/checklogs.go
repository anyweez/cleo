package main

import (
	"flag"
	"fmt"
	logparser "github.com/jeromer/syslogparser/rfc3164"
	"logger"
	"io/ioutil"
	"strings"
)

var (
	LOGS_DIRECTORY 	= flag.String("logs", "/media/vortex/logs/", "The directory where all log files are stored.")
	OPERATION	= flag.String("operation", "report", "The operation to be performed on the logs.")
)

func parseRecord(record string) logger.LoLLogEvent {
	p := logparser.NewParser( []byte(record) )
	p.Parse()

	for k, v := range p.Dump() {
		fmt.Println(k, ":", v)
	}

	return logger.LoLLogEvent{}
}

func readLogs(directory string, events chan logger.LoLLogEvent) {
	files, _ := ioutil.ReadDir(directory)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".log") {
			data, _ := ioutil.ReadFile(file.Name())

			// Parse each line of the file using https://github.com/jeromer/syslogparser
			lines := strings.Split( string(data), "\n")

			for _, line := range lines {
				parseRecord(line)
			}

			// Pass along single events to the EVENTS channel
		}
	}
}

func main() {
	flag.Parse()
	messages := make(chan logger.LoLLogEvent, 10000)

	// Kick off a gortoutine to start reading and parsing log messages.
	go readLogs(*LOGS_DIRECTORY, messages)

	// Depending on what operation th euser requested, do something idfferent
	// with the records as they come in.
	if *OPERATION == "report" {
		build_report(messages)
	}
}
