package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

const (
	defaultDataDir = "./data"
	templateDir    = "webserver/templates"
)

var (
	dataDir  = flag.String("datadir", filepath.FromSlash(defaultDataDir), "The directory to save state information in")
	port     = flag.Int("port", 8080, "The port to listen on")
	logDest  = flag.String("log", "STDERR", "Where to log to - STDOUT, STDERR or a filename")
	logLevel = flag.String("loglevel", "INFO", "Logging verbosity - DEBUG, INFO or WARN")
	zones    = flag.String("zones", "", "The list of zones to use with their corresponding outputs - (id:(pin|'v'),)*")
)

type ZoneAdder interface {
	AddZone(*zone.Zone)
}

func main() {
	flag.Parse()

	setupLogging(*logDest, *logLevel)

	setupDataDir(*dataDir)

	srv := webserver.New(*port, filepath.FromSlash(templateDir))
	err := setupZones(*zones, srv)
	if err != nil {
		logger.Fatal("Error setting up outputs: ", err)
	}
	err = srv.Run()
	if err != nil {
		logger.Fatal("Server.Run: ", err)
	}
}

func setupLogging(dest, level string) {
	err := logger.SetDestination(dest)
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
		log.Fatalln("Unrecognised log level:", level)
	}
}

func setupDataDir(dir string) {
	zone.DataDir = dir
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0777)
			if err == nil {
				return
			}
		}
		logger.Fatalf("Error using data dir '%s': %s", dir, err.Error())
	}
	if !fi.IsDir() {
		logger.Fatalf("Error, data dir '%s' is not a directory", dir)
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
		z.Restore()
		z.Scheduler.Start()
		server.AddZone(z)
	}
	return nil
}
