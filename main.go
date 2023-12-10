//    rlinks, a simple link shortener written in Go.
//    Copyright (C) 2023 redd
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
	e := getEnv(envFile)

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &configuration{
		db:                     database.DbConnect(e.dbType, e.dbURL),
		instanceName:           e.instanceName,
		instanceURL:            e.instanceURL,
		defaultShortLength:     e.defaultLength,
		defaultMaxShortLength:  e.defaultMaxLength,
		defaultMaxCustomLength: e.defaultMaxCustomLength,
		defaultExpiryTime:      e.defaultExpiryTime,
	}

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conf.db)

	// Create the links table, it will check if the table exists before creating it
	database.CreateLinksTable(conf.db, e.defaultMaxLength)

	fs := http.FileServer(http.Dir("static/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	http.HandleFunc("/status", handlerReadiness)               // Check the status of the server
	http.HandleFunc("/error", handlerErr)                      // Check if errors work as intended
	http.HandleFunc("/add", conf.frontHandlerAdd)              // Add a link
	http.HandleFunc("/access", conf.frontHandlerRedirectToUrl) // Access password protected link
	http.HandleFunc("/", conf.apiHandlerRoot)                  // UI for link creation

	// Periodically clean the database
	go conf.collectGarbage(e.timeBetweenCleanups)

	// Start to listen
	log.Printf("Listening on port : '%s'.", e.portStr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", e.portStr), nil))
}
