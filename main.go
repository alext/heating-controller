package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var (
	port     = flag.Int("port", 8080, "The port to listen on")
	logDest  = flag.String("log", "STDERR", "Where to log to - STDOUT, STDERR or a filename")
	logLevel = flag.String("loglevel", "INFO", "Logging verbosity - DEBUG, INFO or WARN")
	zones    = flag.String("zones", "", "The list of zones to use with their corresponding outputs - (id:(pin|'v'),)*")
	schedule = flag.String("schedule", "", "The schedule to use - (hh:mm,(On|Off);)*")
)

type ZoneAdder interface {
	AddZone(*zone.Zone)
}

func main() {
	flag.Parse()

	setupLogging()

	srv := webserver.New(*port, "webserver/templates")
	err := setupZones(*zones, srv)
	if err != nil {
		logger.Fatal("Error setting up outputs: ", err)
	}
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

var output_New = output.New // variable indirection to facilitate testing

var zonePart = regexp.MustCompile(`^([a-z]+):(\d+|v)$`)

func setupZones(zonesParam string, server ZoneAdder) error {
	for _, part := range strings.Split(zonesParam, ",") {
		if part == "" {
			continue
		}

		matches := zonePart.FindStringSubmatch(part)
		if matches == nil {
			return fmt.Errorf("Invalid output entry '%s'", part)
		}

		id := matches[1]
		var out output.Output
		if matches[2] == "v" {
			out = output.Virtual(id)
		} else {
			pin, _ := strconv.Atoi(matches[2])
			var err error
			out, err = output_New(id, pin)
			if err != nil {
				return err
			}
		}
		z := zone.New(id, out)
		if z.ID == "ch" {
			err := processCmdlineSchedule(*schedule, z.Timer)
			if err != nil {
				return err
			}
			z.Timer.Start()
		}
		server.AddZone(z)
	}
	return nil
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
