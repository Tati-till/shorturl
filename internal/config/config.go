package config

import (
	"flag"
)

type Config struct {
	RunAddr string
	ResAddr string
}

var c Config

func ParseFlags() {
	flag.StringVar(&c.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.ResAddr, "b", "http://localhost:8080", "address and port for the result short URL")
	flag.Parse()
}

func GetConfig() *Config {
	return &c
}
