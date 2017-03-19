package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
)

const (
	defaultConfigFile  = "./config.json"
	defaultDataDir     = "./data"
	defaultPort        = 8080
	defaultTemplateDir = "webserver/templates"
)

type ZoneAdder interface {
	AddZone(*controller.Zone)
}

func main() {
	var (
		logDest       = flag.String("log", "STDERR", "Where to log to - STDOUT, STDERR or a filename")
		dataDir       = flag.String("datadir", filepath.FromSlash(defaultDataDir), "The directory to save state information in")
		templateDir   = flag.String("templatedir", filepath.FromSlash(defaultTemplateDir), "The directory containing the templates")
		configFile    = flag.String("config-file", filepath.FromSlash(defaultConfigFile), "Path to the config file")
		returnVersion = flag.Bool("version", false, "Return version and exit")
	)

	flag.Parse()

	if *returnVersion {
		fmt.Printf("heating-controller %s\n", versionInfo())
		os.Exit(0)
	}

	err := setupLogging(*logDest)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[main] heating-controller starting")

	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalln("[main] Error reading config file:", err)
	}

	setupDataDir(*dataDir)

	srv := webserver.New(config.Port, filepath.FromSlash(*templateDir))
	err = setupZones(config.Zones, srv)
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
	controller.DataDir = dir
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

func setupZones(zones map[string]zoneConfig, server ZoneAdder) error {
	for id, config := range zones {
		var out output.Output
		if config.Virtual {
			out = output.Virtual(id)
		} else {
			var err error
			out, err = output_New(id, config.GPIOPin)
			if err != nil {
				return err
			}
		}
		z := controller.NewZone(id, out)
		if config.Thermostat != nil {
			z.SetupThermostat(config.Thermostat.SensorURL, config.Thermostat.DefaultTarget)
		}
		z.Restore()
		z.Scheduler.Start()
		server.AddZone(z)
	}
	return nil
}
