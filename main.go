package main

import (
	"database/sql"
	"fmt"
	"github.com/redds-be/rlinks/database"
	"log"
	"net/http"
)

func main() {
	// Load the env file
	var envFile = ".env"
	portStr, instanceName, instanceURL, dbURL, timeBetweenCleanups := getEnv(envFile)

	// Create a struct to connect to the database and send the instance name and url to the handlers
	handlersInfo := &sendToHandlers{
		db:           database.DbConnect(dbURL),
		instanceName: instanceName,
		instanceURL:  instanceURL,
	}

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(handlersInfo.db)

	// Create the links table, it will check if the table exists before creating it
	database.CreateLinksTable(handlersInfo.db)

	fs := http.FileServer(http.Dir("static/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	http.HandleFunc("/status", handlerReadiness)                       // Check the status of the server
	http.HandleFunc("/error", handlerErr)                              // Check if errors work as intended
	http.HandleFunc("/add", handlersInfo.frontHandlerAdd)              // Add a link
	http.HandleFunc("/access", handlersInfo.frontHandlerRedirectToUrl) // Access password protected link
	http.HandleFunc("/", handlersInfo.apiHandlerRoot)                  // UI for link creation

	// Periodically clean the database
	go handlersInfo.collectGarbage(timeBetweenCleanups)

	// Start to listen
	log.Printf("Listening on port : '%s'.", portStr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", portStr), nil))
}
