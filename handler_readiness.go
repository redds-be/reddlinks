package main

import (
	"log"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)

	// Check method
	if r.Method != http.MethodGet {
		respondWithError(w, r, 405, "405 Method Not Allowed")
		return
	}

	// Define a JSON structure for the status
	type statusResponse struct {
		Status string `json:"status"`
	}

	// Respond to the client with the 'Alive.' message at '/status'
	respondWithJSON(w, 200, statusResponse{Status: "Alive."})
}
