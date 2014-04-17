package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
	"github.com/alext/heating-controller/webserver"
)

var (
	port     = flag.Int("port", 8080, "The port to listen on")
	logDest  = flag.String("log", "STDERR", "Where to log to - STDOUT, STDERR or a filename")
	logLevel = flag.String("loglevel", "INFO", "Logging verbosity - DEBUG, INFO or WARN")
	schedule = flag.String("schedule", "", "The schedule to use - (hh:mm,(On|Off);)*")
)

func main() {
	flag.Parse()

	setupLogging()

	//toggleLoop()

	srv := webserver.New(*port)
	out, err := output.New("ch", 22)
	if err != nil {
		logger.Fatal("Error creating output: ", err)
	}
	t := timer.New(out)
	err = processCmdlineSchedule(*schedule, t)
	if err != nil {
		logger.Fatal(err)
	}

	t.Start()

	srv.AddOutput(out)
	err = srv.Run()
	if err != nil {
		logger.Fatal("Server.Run: ", err)
	}
}

func setupLogging() {
	err := logger.SetDestination(*logDest)
	if err != nil {
		log.Fatalln("Error opening log", err)
	}
	switch *logLevel {
	case "DEBUG":
		logger.Level = logger.DEBUG
	case "INFO":
		logger.Level = logger.INFO
	case "WARN":
		logger.Level = logger.WARN
	default:
		log.Fatalln("Unrecognised log level:", *logLevel)
	}
}

var schedulePart = regexp.MustCompile(`^(\d+):(\d+),(On|Off)$`)

func processCmdlineSchedule(schedule string, t timer.Timer) error {
	for _, part := range strings.Split(schedule, ";") {
		if part == "" {
			continue
		}
		matches := schedulePart.FindStringSubmatch(part)
		if matches == nil {
			return fmt.Errorf("Invalid schedule entry %s", part)
		}

		hour, _ := strconv.Atoi(matches[1])
		min, _ := strconv.Atoi(matches[2])
		if hour < 0 || hour > 23 || min < 0 || min > 59 {
			return fmt.Errorf("Invalid schedule entry %s", part)
		}
		if matches[3] == "On" {
			t.AddEntry(hour, min, timer.TurnOn)
		} else {
			t.AddEntry(hour, min, timer.TurnOff)
		}
	}
	return nil
}

func toggleLoop() {
	out, err := output.New("ch", 22)
	if err != nil {
		log.Fatal("Error creating output:", err)
	}
	for true {
		log.Print("Activating output")
		err = out.Activate()
		if err != nil {
			log.Fatal("Error activating output:", err)
		}
		state, err := out.Active()
		if err != nil {
			log.Fatal("Error reading state:", err)
		}
		log.Printf("  Current state: %v", state)
		state, err = out.Active()
		if err != nil {
			log.Fatal("Error reading state:", err)
		}
		log.Printf("  Current state: %v", state)

		log.Print("   sleeping...")
		time.Sleep(5 * time.Second)

		log.Print("Deactivating output")
		err = out.Deactivate()
		if err != nil {
			log.Fatal("Error deactivating output:", err)
		}
		state, err = out.Active()
		if err != nil {
			log.Fatal("Error reading state:", err)
		}
		log.Printf("  Current state: %v", state)

		log.Print("   sleeping...")
		time.Sleep(5 * time.Second)
	}
}
