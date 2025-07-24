package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Serving file from . to port %s ", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
