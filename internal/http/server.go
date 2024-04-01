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
	"crypto/tls"
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
		ContactEmail:           configuration.ContactEmail,
		Version:                configuration.Version,
		PortSTR:                configuration.PortSTR,
		TLSPortSTR:             configuration.TLSPortSTR,
		CertFile:               configuration.CertFile,
		KeyFile:                configuration.KeyFile,
		TLSEnabled:             configuration.TLSEnabled,
		DefaultShortLength:     configuration.DefaultShortLength,
		DefaultMaxShortLength:  configuration.DefaultMaxShortLength,
		DefaultMaxCustomLength: configuration.DefaultMaxCustomLength,
		DefaultExpiryTime:      configuration.DefaultExpiryTime,
	}
}

// redirectToHTTPS redirects incoming http requests to https, only used when TLS is enabled.
func (conf Configuration) redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request
	http.Redirect(w, r, conf.InstanceURL+r.RequestURI, http.StatusMovedPermanently)
}

// Run starts configures the HTTP server and starts listening and serving.
func (conf Configuration) Run() error { //nolint:funlen
	// Set default timeout time in seconds
	const readTimeout = 1 * time.Second
	const WriteTimeout = 1 * time.Second
	const IdleTimeout = 30 * time.Second
	const ReadHeaderTimeout = 2 * time.Second

	// Create a multiplexer
	mux := http.NewServeMux()

	// Make use of the assets
	fs := http.FileServer(http.Dir("static/assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	// Assign a handler to these different paths
	mux.HandleFunc("GET /status", HandlerReadiness) // Check the status of the server
	mux.HandleFunc(
		"GET /error",
		HandlerErr,
	) // Check if errors work as intended
	mux.HandleFunc(
		"POST /add",
		conf.FrontHandlerAdd,
	) // Front page for adding a link that returns the basic info
	mux.HandleFunc("POST /access", conf.FrontHandlerRedirectToURL) // Access a password protected link
	mux.HandleFunc("GET /privacy", conf.FrontHandlerPrivacyPage)   // Display Privacy policy information page
	mux.HandleFunc("GET /", conf.FrontHandlerMainPage)             // Main page with the form to create a link
	mux.HandleFunc("GET /{short}", conf.APIRedirectToURL)          // Access an url
	mux.HandleFunc("POST /", conf.APICreateLink)                   // Create a link

	var srv *http.Server
	var redirectSrv *http.Server
	if conf.TLSEnabled {
		// Configure TLS
		TLSConf := &tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				// TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA is "secure" (as of writing this), gosec is giving a false positive.
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, //nolint:gosec
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		// Set the settings for the https server
		srv = &http.Server{
			Addr:              ":" + conf.TLSPortSTR,
			ReadTimeout:       readTimeout,
			WriteTimeout:      WriteTimeout,
			IdleTimeout:       IdleTimeout,
			ReadHeaderTimeout: ReadHeaderTimeout,
			Handler:           mux,
			TLSConfig:         TLSConf,
			TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}

		// Set the settings for the http redirection server
		redirectSrv = &http.Server{
			Addr:              ":" + conf.PortSTR,
			ReadTimeout:       readTimeout,
			WriteTimeout:      WriteTimeout,
			IdleTimeout:       IdleTimeout,
			ReadHeaderTimeout: ReadHeaderTimeout,
		}
	} else {
		// Set the settings for the http server
		srv = &http.Server{
			Addr:              ":" + conf.PortSTR,
			ReadTimeout:       readTimeout,
			WriteTimeout:      WriteTimeout,
			IdleTimeout:       IdleTimeout,
			ReadHeaderTimeout: ReadHeaderTimeout,
			Handler:           mux,
		}
	}

	// Start to listen
	var err error
	if conf.TLSEnabled {
		log.Printf("Listening on port : '%s'.", conf.TLSPortSTR)
		go func() {
			err := srv.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
			if err != nil {
				log.Panic(err)
			}
		}()
		http.HandleFunc("/*", conf.redirectToHTTPS)
		err = redirectSrv.ListenAndServe()
	} else {
		log.Printf("Listening on port : '%s'.", conf.PortSTR)
		err = srv.ListenAndServe()
	}

	return err
}
