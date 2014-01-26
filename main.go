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

	output, err := output.New("ch", 22)
	if err != nil {
		log.Fatal("Error creating output:", err)
	}
	for ; true ; {
		log.Print("Activating output")
		err = output.Activate()
		if err != nil {
			log.Fatal("Error activating output:", err)
		}

		log.Print("   sleeping...")
		time.Sleep(5 * time.Second)

		log.Print("Deactivating output")
		err = output.Deactivate()
		if err != nil {
			log.Fatal("Error deactivating output:", err)
		}

		log.Print("   sleeping...")
		time.Sleep(5 * time.Second)
	}
	srv := webserver.New(*port)
	err = srv.Run()
	if err != nil {
		log.Fatal("Server.Run: ", err)
	}
}
