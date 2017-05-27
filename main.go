package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/webserver"
)

const (
	defaultConfigFile  = "./config.json"
	defaultDataDir     = "./data"
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

	config, err := loadConfigFile(*configFile)
	if err != nil {
		log.Fatalln("[main] Error reading config file:", err)
	}

	setupDataDir(*dataDir)
	ctrl := controller.New()

	err = ctrl.Setup(config)
	if err != nil {
		log.Fatalln("[main] Error setting up controller:", err)
	}

	srv := webserver.New(ctrl, config.Port, filepath.FromSlash(*templateDir))
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

func loadConfigFile(filename string) (*config.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[main] Config file '%s' not found, ignoring", filename)
			return config.New(), nil
		}
		return nil, err
	}
	defer file.Close()

	return config.LoadConfig(file)
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
