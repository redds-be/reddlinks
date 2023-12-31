//    reddlinks, a simple link shortener written in Go.
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

	"github.com/redds-be/reddlinks/database"
)

// Set a global variable for a token.
var token string //nolint:gochecknoglobals

func main() { //nolint:funlen
	// Load the env file
	envFile := ".env"
	env := getEnv(envFile)

	dataBase, err := database.DBConnect(env.dbType, env.dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &configuration{
		db:                     dataBase,
		instanceName:           env.instanceName,
		instanceURL:            env.instanceURL,
		defaultShortLength:     env.defaultLength,
		defaultMaxShortLength:  env.defaultMaxLength,
		defaultMaxCustomLength: env.defaultMaxCustomLength,
		defaultExpiryTime:      env.defaultExpiryTime,
		version:                "noVersion",
	}

	// Set default timeout time in seconds
	const readTimeout = 1 * time.Second
	const WriteTimeout = 1 * time.Second
	const IdleTimeout = 30 * time.Second
	const ReadHeaderTimeout = 2 * time.Second

	// Generate a new token every x time
	go func(duration time.Duration) {
		for {
			token = randomToken()
			time.Sleep(duration)
		}
	}(3 * time.Hour) //nolint:gomnd

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conf.db)

	// Create the links table if it doesn't exist
	err = database.CreateLinksTable(conf.db, env.defaultMaxLength)
	if err != nil {
		log.Panic(err)
	}

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
	go func(duration time.Duration) {
		for {
			err := conf.collectGarbage()
			if err != nil {
				log.Println("Could not collect garbage:", err)
			}
			time.Sleep(duration)
		}
	}(time.Duration(env.timeBetweenCleanups) * time.Minute)

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
