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

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

const (
	defaultDataDir = "./data"
	templateDir    = "webserver/templates"
)

var (
	dataDir = flag.String("datadir", filepath.FromSlash(defaultDataDir), "The directory to save state information in")
	port    = flag.Int("port", 8080, "The port to listen on")
	logDest = flag.String("log", "STDERR", "Where to log to - STDOUT, STDERR or a filename")
	zones   = flag.String("zones", "", "The list of zones to use with their corresponding outputs - (id:(pin|'v'),)*")
)

type ZoneAdder interface {
	AddZone(*zone.Zone)
}

func main() {
	returnVersion := flag.Bool("version", false, "Return version and exit")

	flag.Parse()

	if *returnVersion {
		fmt.Printf("heating-controller %s\n", versionInfo())
		os.Exit(0)
	}

	err := setupLogging(*logDest)
	if err != nil {
		log.Fatal(err)
	}

	setupDataDir(*dataDir)

	srv := webserver.New(*port, filepath.FromSlash(templateDir))
	err = setupZones(*zones, srv)
	if err != nil {
		log.Fatalln("[main] Error setting up outputs:", err)
	}
	err = srv.Run()
	if err != nil {
		log.Fatalln("[main] Server.Run:", err)
	}
}

func setupLogging(destination string) error {
	switch destination {
	case "STDERR":
		log.SetOutput(os.Stderr)
	case "STDOUT":
		log.SetOutput(os.Stdout)
	default:
		file, err := os.OpenFile(destination, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("Error opening log %s: %s", destination, err.Error())
		}
		log.SetOutput(file)
	}
	return nil
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
		log.Fatalf("[main] Error using data dir '%s': %s", dir, err.Error())
	}
	if !fi.IsDir() {
		log.Fatalf("[main] Error, data dir '%s' is not a directory", dir)
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
