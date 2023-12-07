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
	portStr, defaultLength, defaultMaxLength, defaultMaxCustomLength, defaultExpiryTime, instanceName, instanceURL, dbURL, timeBetweenCleanups := getEnv(envFile)

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &configuration{
		db:                     database.DbConnect(dbURL),
		instanceName:           instanceName,
		instanceURL:            instanceURL,
		defaultShortLength:     defaultLength,
		defaultMaxShortLength:  defaultMaxLength,
		defaultMaxCustomLength: defaultMaxCustomLength,
		defaultExpiryTime:      defaultExpiryTime,
	}

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conf.db)

	// Create the links table, it will check if the table exists before creating it
	database.CreateLinksTable(conf.db, defaultMaxLength)

	fs := http.FileServer(http.Dir("static/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	http.HandleFunc("/status", handlerReadiness)               // Check the status of the server
	http.HandleFunc("/error", handlerErr)                      // Check if errors work as intended
	http.HandleFunc("/add", conf.frontHandlerAdd)              // Add a link
	http.HandleFunc("/access", conf.frontHandlerRedirectToUrl) // Access password protected link
	http.HandleFunc("/", conf.apiHandlerRoot)                  // UI for link creation

	// Periodically clean the database
	go conf.collectGarbage(timeBetweenCleanups)

	// Start to listen
	log.Printf("Listening on port : '%s'.", portStr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", portStr), nil))
}
