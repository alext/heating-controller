package main

import (
	"flag"
	"log"
)

var port = flag.Int("port", 8080, "The port to listen on")

func main() {
	flag.Parse()

	srv := NewWebServer(*port)
	err := srv.Run()
	if err != nil {
		log.Fatal("Server.Run: ", err)
	}
}
