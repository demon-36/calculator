package main

import (
	"log"
	"net/http"
	"time"

	"calculator/api"
)

func main() {
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Calculator API listening at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
