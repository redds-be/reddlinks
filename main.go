//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2024 redd
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
	"html/template"
	"log"

	"github.com/redds-be/reddlinks/internal/cron"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
)

// main drives the application.
func main() {
	// Load the env file
	envFile := ".env"
	envVars := env.GetEnv(envFile)

	dbase, err := database.DBConnect(envVars.DBType, envVars.DBURL)
	if err != nil {
		log.Fatal(err)
	}

	// Defer the closing of the database connection
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dbase)

	// Create the links table if it doesn't exist
	err = database.CreateLinksTable(dbase, envVars.DefaultMaxLength)
	if err != nil {
		log.Panic(err)
	}

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &utils.Configuration{
		DB:                     dbase,
		PortSTR:                envVars.PortStr,
		InstanceName:           envVars.InstanceName,
		InstanceURL:            envVars.InstanceURL,
		DefaultShortLength:     envVars.DefaultLength,
		DefaultMaxShortLength:  envVars.DefaultMaxLength,
		DefaultMaxCustomLength: envVars.DefaultMaxCustomLength,
		DefaultExpiryTime:      envVars.DefaultExpiryTime,
		Version:                "noVersion",
	}

	cron.StartJobs(*conf, envVars)

	// Parse html templates
	http.Templates = template.Must(template.ParseFiles("static/index.html", "static/add.html",
		"static/error.html", "static/pass.html", "static/privacy.html"))

	// Create an adapter for the server
	httpAdapter := http.NewAdapter(*conf)

	// Start the server
	err = httpAdapter.Run()
	if err != nil {
		log.Panic(err)
	}
}
