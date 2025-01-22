package main

import (
	"flag"
	"os"
)

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
}
