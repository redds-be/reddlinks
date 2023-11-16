package main

import (
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	// Load the env file
	var envFile = ".env"
	portStr, dbURL := getEnv(envFile)

	// Create the router
	router := getRouter(dbURL)

	// Create the http handler
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portStr,
	}

	go collectGarbage(portStr)

	// Start to listen
	log.Printf("Listening on port : '%s'.", portStr)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
