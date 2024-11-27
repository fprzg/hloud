package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	port int
	env  string
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4001, "Frontend port.")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production).")

	flag.Parse()

	const filePath = "./cmd/frontend/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading index.html", http.StatusInternalServerError)
			log.Printf("Error reading file: %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(htmlContent)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	fmt.Printf("Listening on port :%d\n", cfg.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil))
}
