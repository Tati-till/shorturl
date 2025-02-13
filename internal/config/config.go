package config

import (
	"flag"
	"os"
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

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}
	if envResAddr := os.Getenv("BASE_URL"); envResAddr != "" {
		c.ResAddr = envResAddr
	}
}

func GetConfig() *Config {
	return &c
}
