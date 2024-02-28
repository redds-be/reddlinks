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

package http

import (
	"log"
	"net/http"
	"time"

	"github.com/redds-be/reddlinks/internal/utils"
)

type Configuration utils.Configuration

// NewAdapter returns a configuration to be used by Run() and the handlers.
func NewAdapter(configuration utils.Configuration) Configuration {
	return Configuration{
		DB:                     configuration.DB,
		InstanceName:           configuration.InstanceName,
		InstanceURL:            configuration.InstanceURL,
		Version:                configuration.Version,
		PortSTR:                configuration.PortSTR,
		DefaultShortLength:     configuration.DefaultShortLength,
		DefaultMaxShortLength:  configuration.DefaultMaxShortLength,
		DefaultMaxCustomLength: configuration.DefaultMaxCustomLength,
		DefaultExpiryTime:      configuration.DefaultExpiryTime,
		ContactEmail:           configuration.ContactEmail,
	}
}

// Run starts configures the HTTP server and starts listening and serving.
func (conf Configuration) Run() error {
	// Set default timeout time in seconds
	const readTimeout = 1 * time.Second
	const WriteTimeout = 1 * time.Second
	const IdleTimeout = 30 * time.Second
	const ReadHeaderTimeout = 2 * time.Second

	// Make use of the assets
	fs := http.FileServer(http.Dir("static/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	http.HandleFunc("/status", HandlerReadiness)               // Check the status of the server
	http.HandleFunc("/error", HandlerErr)                      // Check if errors work as intended
	http.HandleFunc("/add", conf.FrontHandlerAdd)              // Add a link
	http.HandleFunc("/access", conf.FrontHandlerRedirectToURL) // Access password protected link
	http.HandleFunc("/privacy", conf.FrontHandlerPrivacyPage)  // Privacy policy information
	http.HandleFunc("/", conf.APIHandlerRoot)                  // UI for link creation

	// Set the settings for the http servers
	srv := &http.Server{
		Addr:              ":" + conf.PortSTR,
		ReadTimeout:       readTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	// Start to listen
	log.Printf("Listening on port : '%s'.", conf.PortSTR)
	err := srv.ListenAndServe()

	return err
}
