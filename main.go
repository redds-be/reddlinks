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
	"log"
	"net/http"
	"time"

	"github.com/redds-be/rlinks/database"
)

func main() {
	// Load the env file
	envFile := ".env"
	env := getEnv(envFile)

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &configuration{
		db:                     database.DBConnect(env.dbType, env.dbURL),
		instanceName:           env.instanceName,
		instanceURL:            env.instanceURL,
		defaultShortLength:     env.defaultLength,
		defaultMaxShortLength:  env.defaultMaxLength,
		defaultMaxCustomLength: env.defaultMaxCustomLength,
		defaultExpiryTime:      env.defaultExpiryTime,
		Version:                "noVersion",
	}

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conf.db)

	// Create the links table, it will check if the table exists before creating it
	database.CreateLinksTable(conf.db, env.defaultMaxLength)

	fs := http.FileServer(http.Dir("static/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	http.HandleFunc("/status", handlerReadiness)               // Check the status of the server
	http.HandleFunc("/error", handlerErr)                      // Check if errors work as intended
	http.HandleFunc("/add", conf.frontHandlerAdd)              // Add a link
	http.HandleFunc("/access", conf.frontHandlerRedirectToURL) // Access password protected link
	http.HandleFunc("/privacy", conf.frontHandlerPrivacyPage)  // Privacy policy information
	http.HandleFunc("/", conf.apiHandlerRoot)                  // UI for link creation

	// Periodically clean the database
	go conf.collectGarbage(env.timeBetweenCleanups)

	// Set default timeout time in seconds
	const readTimeout = 1 * time.Second
	const WriteTimeout = 1 * time.Second
	const IdleTimeout = 30 * time.Second
	const ReadHeaderTimeout = 2 * time.Second

	// Set the settings for the http servers
	srv := &http.Server{
		Addr:              ":" + env.portStr,
		ReadTimeout:       readTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	// Start to listen
	log.Printf("Listening on port : '%s'.", env.portStr)
	log.Panic(srv.ListenAndServe())
}
