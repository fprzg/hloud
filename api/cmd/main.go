package main

import (
	"flag"
	"fmt"
	"log"

	"api.hloud.fprzg.net/internal/handlers"
	"api.hloud.fprzg.net/internal/info"
)

type application struct {
	handlers handlers.Handlers
	cfg      *info.Config
}

var (
	build info.Build
)

func main() {
	var cfg info.Config

	flag.IntVar(&cfg.Port, "port", 4000, "API server port.")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production).")
	flag.StringVar(&cfg.StorageDir, "storage", "/home/ubu24-t480/Code/hloud/storage", "Directory where you store the files.")

	flag.Parse()

	app := application{
		cfg:      &cfg,
		handlers: handlers.NewHandlers(&cfg, &build),
	}

	fmt.Printf("Listening on port %s", app.cfg.GetPort())
	err := app.serve()
	if err != nil {
		log.Fatalf("Error %v\n", err)
	}
}
