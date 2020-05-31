package main

import (
	"log"
	"os"

	client "github.com/zerogvt/eventfull/client"
	server "github.com/zerogvt/eventfull/server"
)

const usage = "Usage: eventfull [client | server]"

func main() {
	if len(os.Args) < 2 {
		log.Fatalln(usage)
	}
	whatamI := os.Args[1]
	if whatamI == "client" {
		configfile := "conf.json"
		evttemplate := "event.json"
		if len(os.Args) > 2 {
			configfile = os.Args[2]
		}
		if len(os.Args) > 3 {
			evttemplate = os.Args[3]
		}
		client.Daemon(configfile, evttemplate)
	} else if whatamI == "server" {
		server.Exec()
	} else {
		log.Fatalf("Unknown command: %s\n%s", whatamI, usage)
	}
}
