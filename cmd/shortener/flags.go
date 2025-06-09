package main

import (
	"flag"
	"github.com/MaxRadzey/shortener/internal/config"
)

func ParseFlag(config *config.Config) {
	flag.StringVar(&config.Address, "a", config.Address, "address and port to run server")
	flag.StringVar(&config.ReturningAddress, "b", config.ReturningAddress, "address to return URL")

	flag.Parse()
}
