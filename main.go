package main

import (
	"flag"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
	"log"
	"time"
)

var port = flag.Int("port", 8080, "The port to listen on")

func main() {
	flag.Parse()

	//toggleLoop()

	srv := webserver.New(*port)
	out, err := output.New("ch", 22)
	if err != nil {
		log.Fatal("Error creating output: ", err)
	}
	srv.AddOutput(out)
	err = srv.Run()
	if err != nil {
		log.Fatal("Server.Run: ", err)
	}
}

func toggleLoop() {
	out, err := output.New("ch", 22)
	if err != nil {
		log.Fatal("Error creating output:", err)
	}
	for ; true ; {
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
