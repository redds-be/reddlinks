package main

import (
	"log"
	"net/http"
)

func handlerErr(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
	// Respond with a generic error at '/error'
	respondWithError(w, r, 400, "Something went wrong.")
}
