package main

import (
	"log"
	"net/http"
	"os"

	handler "json_to_qr_api/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, http.HandlerFunc(handler.Handler)); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
