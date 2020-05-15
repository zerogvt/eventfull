package main

import "github.com/zerogvt/eventfull"

func main() {
	eventfull.Daemon("conf.json", "event.json")
}
