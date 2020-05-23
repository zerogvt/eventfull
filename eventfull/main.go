package main

import (
	"log"
	"os"

	client "github.com/zerogvt/eventfull/client"
	server "github.com/zerogvt/eventfull/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Usage: eventfull [client | server]")
	}
	whatamI := os.Args[1]
	if whatamI == "client" {
		client.Daemon("conf.json", "event.json")
	} else if whatamI == "server" {
		server.Exec()
	} else {
		log.Fatalf("Unknown command: %s", whatamI)
	}
}
