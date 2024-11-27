package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	version string
)

//buildTime string

type config struct {
	port       int
	env        string
	storageDir string
}

type application struct {
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port.")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production).")
	flag.StringVar(&cfg.storageDir, "storage", "/home/ubu24-t480/Code/hloud/storage", "Directory where you store the files.")

	flag.Parse()

	app := application{
		config: cfg,
	}

	fmt.Printf("Listening on port :%d", app.config.port)
	err := app.serve()
	if err != nil {
		log.Fatalf("Error %v\n", err)
	}
}
