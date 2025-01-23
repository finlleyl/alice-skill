package main

import (
	"flag"
	"os"
)

var (
	flagRunAddr  string
	flagLogLevel string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if envRunAddr, exists := os.LookupEnv("RUN_ADDR"); exists {
		flagRunAddr = envRunAddr
	}

	if envLogLevel, exists := os.LookupEnv("LOG_LEVEL"); exists {
		flagLogLevel = envLogLevel
	}
}
