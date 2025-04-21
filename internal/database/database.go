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

// Package database provides utilities for database connection management and query operations.
//
// This package supports both PostgreSQL and SQLite database systems, handling
// connection establishment and maintenance for URL shortening services.
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"    // Driver for PostgreSQL
	_ "modernc.org/sqlite" // Driver for SQLite3
)

// DBConnect establishes a connection to the specified database.
//
// This function initiates a connection to either a PostgreSQL or SQLite database
// using the provided parameters.
//
// The function accepts either a complete connection URL or individual connection
// parameters. If both are provided, the URL takes precedence.
//
// Parameters:
//   - dbType: Database type ("postgres" or "sqlite")
//   - dbURL: Complete database connection URL (optional if individual parameters provided)
//   - dbUser: Database username (used if dbURL is empty)
//   - dbPass: Database password (used if dbURL is empty)
//   - dbHost: Database host address (used if dbURL is empty)
//   - dbPort: Database port number (used if dbURL is empty)
//   - dbName: Database name (used if dbURL is empty)
//
// Returns:
//   - *sql.DB: A database connection object
//   - error: Any error encountered during connection establishment
func DBConnect(dbType, dbURL, dbUser, dbPass, dbHost, dbPort, dbName string) (*sql.DB, error) { //nolint:cyclop
	var (
		dbase *sql.DB
		err   error
	)

	// Determine how to connect based on provided parameters
	if dbURL != "" { //nolint:nestif
		// Connect using the provided URL
		dbase, err = sql.Open(dbType, dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection with URL: %w", err)
		}

		// Test the connection
		err = dbase.Ping()

		// Handle PostgreSQL SSL negotiation if needed
		if dbType == "postgres" && errors.Is(err, pq.ErrSSLNotSupported) {
			// Close the existing connection
			err := dbase.Close()
			if err != nil {
				return nil, err
			}

			// Retry with SSL disabled
			dbase, err = sql.Open(dbType, dbURL+"?sslmode=disable")
			if err != nil {
				return nil, fmt.Errorf("failed to open PostgreSQL connection with SSL disabled: %w", err)
			}

			// Test the connection again
			// golangci-lint doesn't like that but the error value is checked at the end of the function
			err = dbase.Ping() //nolint:ineffassign,staticcheck,wastedassign
		}
	} else {
		// Construct a connection string from individual parameters
		connectionString := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s", dbUser, dbName, dbPass, dbHost, dbPort)

		// Connect to the database with the constructed connection string
		dbase, err = sql.Open(dbType, connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}

		// Test the connection
		err = dbase.Ping()

		// Handle PostgreSQL SSL negotiation if needed
		if errors.Is(err, pq.ErrSSLNotSupported) {
			// Close the existing connection
			err := dbase.Close()
			if err != nil {
				return nil, err
			}

			// Retry with SSL disabled
			dbase, err = sql.Open(dbType, connectionString+" sslmode=disable")
			if err != nil {
				return nil, fmt.Errorf("failed to open PostgreSQL connection with SSL disabled: %w", err)
			}

			// Test the connection again
			// golangci-lint doesn't like that but the error value is checked at the end of the function
			err = dbase.Ping() //nolint:ineffassign,staticcheck,wastedassign
		}
	}

	// Final error check after all connection attempts
	if err != nil {
		// Ensure connection is closed if ping failed
		if dbase != nil {
			err := dbase.Close()
			if err != nil {
				return nil, err
			}
		}

		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	return dbase, nil
}
