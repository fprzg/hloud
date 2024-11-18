package main

import (
	"flag"
	"fmt"
)

var (
	version   string
	buildTime string
)

type config struct {
	port int
	env  string
}

type application struct {
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.Parse()

	app := application{
		config: cfg,
	}

	fmt.Printf("Listening on port :%d", app.config.port)
	err := app.serve()
	if err != nil {
		return
	}
}
