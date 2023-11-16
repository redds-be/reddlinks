package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	// Respond the client with a JSON error
	log.Printf("Responding with an error to %s (%s) at '%s' with method '%s':\nError: %s (%d)\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method, msg, code)

	// Define a JSON structure for the error
	type errResponse struct {
		Error string `json:"error"`
	}

	// Send the JSON the client along with the error code
	respondWithJSON(w, code, errResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Create a JSON response, internal error if it can't
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v\n", payload)
		w.WriteHeader(500)
		return
	}

	// Add the json header to the response sa that the client can interpret it as JSON, internal error if it can't
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(dat)
	if err != nil {
		log.Printf("Failed to write JSON response: %v\n", dat)
		w.WriteHeader(500)
		return
	}
}
