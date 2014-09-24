package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
)

/**
 * Logger class that can be instantiated to start doing all of the
 * logging that's necessary.
 */
type LoLLogger struct {
	logger      *syslog.Writer
	initialized bool
}

type LoLLogEvent struct {
	Priority  syslog.Priority `json:"-"`
	Operation LoLOperation
	Outcome   LoLOutcome
	Target    uint64
	Details   string
}

type LoLOperation int
type LoLOutcome int

/**
 * A list of operations that are logging events.
 */
const (
	FETCH_MATCH_HISTORY LoLOperation = iota
	FETCH_GAME_STATS    LoLOperation = iota
	FETCH_NAME          LoLOperation = iota
)

/**
 * A list of possible logged outcomes for the above operations.
 */
const (
	SUCCESS                 LoLOutcome = iota
	API_REQUEST_FAILURE     LoLOutcome = iota
	API_RATE_LIMIT_EXCEEDED LoLOutcome = iota
)

func (self *LoLLogger) Init() {
	if self.initialized {
		return
	}

	// The tag is the executable's name.
	exe_components := strings.Split(os.Args[0], "/")
	tag := exe_components[len(exe_components)-1]

	self.logger, _ = syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL0, tag)
	self.initialized = true

	log.Println("Logging service initialized.")
}

func (self *LoLLogger) Log(event LoLLogEvent) {
	self.Init()

	if event.Priority&syslog.LOG_INFO > 0 {
		event_str, _ := json.Marshal(event)
		self.logger.Info(fmt.Sprintf("[LOL] %s", event_str))
	} else {
		log.Println("Unknown priority for log event.")
	}
}
