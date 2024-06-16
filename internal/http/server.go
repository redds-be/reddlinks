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

// Package http listen and serves clients using its handlers,
// there is a set of handlers for the REST API and a set of handlers for the front-facing website.
package http

import (
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/redds-be/reddlinks/internal/utils"
)

// Configuration is redefines [utils.Configuration] to be used for methods within the package.
type Configuration utils.Configuration

// NewAdapter returns a configuration to be used by Run and the handlers.
// Check [utils.Configuration] to know about these fields.
func NewAdapter(configuration utils.Configuration) Configuration {
	return Configuration{
		DB:                     configuration.DB,
		InstanceName:           configuration.InstanceName,
		InstanceURL:            configuration.InstanceURL,
		Version:                configuration.Version,
		AddrAndPort:            configuration.AddrAndPort,
		DefaultShortLength:     configuration.DefaultShortLength,
		DefaultMaxShortLength:  configuration.DefaultMaxShortLength,
		DefaultMaxCustomLength: configuration.DefaultMaxCustomLength,
		DefaultExpiryTime:      configuration.DefaultExpiryTime,
		ContactEmail:           configuration.ContactEmail,
		Static:                 configuration.Static,
	}
}

// Run starts configures the HTTP server and starts listening and serving.
//
// It starts by setting constants for the timeouts, after that,
// a filesystem is created for the assets, followed by the creation of
// a file server using the new filesystem. It is followed by the creation of a multiplexer
// which will handle the endpoints.
// GET /assets/ if for serving the assets,
// GET /status calls [HandlerReadiness] for health check,
// POST /add calls FrontHandlerAdd, which creates a link and displays the information in a browser,
// POST /access calls FrontHandlerRedirectToURL, which is used to access a password protected link,
// GET /privacy calls FrontHandlerPrivacyPage, which is used to display the privacy policy,
// GET / calls FrontHandlerMainPage, which is used to serve a form to shorten a link,
// GET /{short} calls APIRedirectToURL, which is used to access a url based on the give short,
// POST / calls APICreateLink, which is used to create a link record in the database.
// After the multiplexer is configured, the HTTP server needs to be configured with the address and port,
// the timeouts constants and the multiplexer as the handler. After the configuration is set,
// [http.ListenAndServe] is called.
func (conf Configuration) Run() error {
	// Set default timeout time in seconds
	const readTimeout = 1 * time.Second
	const WriteTimeout = 1 * time.Second
	const IdleTimeout = 30 * time.Second
	const ReadHeaderTimeout = 2 * time.Second

	// Create the filesystem for the assets
	assetsFS, err := fs.Sub(conf.Static, "static/assets")
	if err != nil {
		log.Panic(err)
	}

	// Create a file server using the assets filesystem
	assetsHTTPFS := http.FileServer(http.FS(assetsFS))

	// Create a multiplexer
	mux := http.NewServeMux()

	// Handle the assets
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetsHTTPFS))

	// Assign a handler to these different paths
	mux.HandleFunc("GET /status", HandlerReadiness) // Check the status of the server
	mux.HandleFunc(
		"POST /add",
		conf.FrontHandlerAdd,
	) // Front page for adding a link that returns the basic info
	mux.HandleFunc(
		"POST /access",
		conf.FrontHandlerRedirectToURL,
	) // Access a password protected link
	mux.HandleFunc(
		"GET /privacy",
		conf.FrontHandlerPrivacyPage,
	) // Display Privacy policy information page
	mux.HandleFunc(
		"GET /",
		conf.FrontHandlerMainPage,
	) // Main page with the form to create a link
	mux.HandleFunc("GET /{short}", conf.APIRedirectToURL) // Access a url
	mux.HandleFunc("POST /", conf.APICreateLink)          // Create a link

	// Set the settings for the http server
	srv := &http.Server{
		Addr:              conf.AddrAndPort,
		ReadTimeout:       readTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
		Handler:           mux,
	}

	// Start to listen
	log.Printf("Listening on: '%s'.", conf.AddrAndPort)
	err = srv.ListenAndServe()

	return err
}
