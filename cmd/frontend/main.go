package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	const filePath = "index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading index.htlm", http.StatusInternalServerError)
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

	port := 6969
	fmt.Printf("Listening on port :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
