//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2025 redd
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

// Package main is the package that will drive the program.
package main

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
)

// version is a variable for the version set by ldflags.
var version string

// embeddedStatic is the variable that will hold the assets within the binary.
//
//go:embed static
var embeddedStatic embed.FS

// main function drives the application.
//
// It starts by loading the environnement variables using [env.GetEnv],
// then it connects to the dabaase using [database.DBConnect] and creates the links table using [database.CreateLinksTable],
// following that, the env vars and the database are gathered into a configuration struct [utils.Configuration].
// It starts a go routines that calls [utils.CollectGarbage] inside an infinite loop with a sleep period defines in the config.
// Following that, HTML templates stored in [embeddedStatic] (containing the 'static/' dir) are parsed using [template.Must].
// At then end, an adapter for the internal HTTP package is created using [http.NewAdapter],
// lastly, the HTTP server gets started using [http.Run].
func main() { //nolint:funlen
	// Load the env file
	envFile := ".env"
	envVars := env.GetEnv(envFile)

	// Connect to the database
	dbase, err := database.DBConnect(
		envVars.DBType,
		envVars.DBURL,
		envVars.DBUser,
		envVars.DBPass,
		envVars.DBHost,
		envVars.DBPort,
		envVars.DBName,
	)
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
	err = database.CreateLinksTable(dbase, envVars.DBType, envVars.DefaultMaxLength)
	if err != nil {
		log.Panic(err)
	}

	// Parse html templates and get the locales
	var locales map[string]utils.PageLocaleTl
	var supportedLocales []string
	if _, err := os.Stat("./custom_static"); !os.IsNotExist(err) {
		http.Templates = template.Must(template.ParseGlob("custom_static/templates/*.tmpl"))
		locales, supportedLocales, err = utils.GetLocales("custom_static/locales/", embeddedStatic)
		if err != nil {
			log.Panic(err)
		}
	} else {
		http.Templates = template.Must(template.ParseFS(embeddedStatic, "static/templates/*.tmpl"))
		// Get locales and the list of supported ones
		locales, supportedLocales, err = utils.GetLocales("", embeddedStatic)
		if err != nil {
			log.Panic(err)
		}
	}

	// Create a struct to connect to the database and send the instance name and url to the handlers
	conf := &utils.Configuration{
		DB:                     dbase,
		AddrAndPort:            envVars.AddrAndPort,
		InstanceName:           envVars.InstanceName,
		InstanceURL:            envVars.InstanceURL,
		DefaultShortLength:     envVars.DefaultLength,
		DefaultMaxShortLength:  envVars.DefaultMaxLength,
		DefaultMaxCustomLength: envVars.DefaultMaxCustomLength,
		DefaultExpiryTime:      envVars.DefaultExpiryTime,
		ContactEmail:           envVars.ContactEmail,
		Static:                 embeddedStatic,
		Version:                version,
		SupportedLocales:       supportedLocales,
		Locales:                locales,
	}

	// Periodically clean the database
	go func(duration time.Duration) {
		for {
			err := conf.CollectGarbage()
			if err != nil {
				log.Println("Could not collect garbage:", err)
			}
			time.Sleep(duration)
		}
	}(time.Duration(envVars.TimeBetweenCleanups) * time.Minute)

	// Create an adapter for the server
	httpAdapter := http.NewAdapter(*conf)

	// Start the server
	err = httpAdapter.Run()
	if err != nil {
		log.Panic(err)
	}
}
