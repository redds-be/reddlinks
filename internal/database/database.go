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

// Package database is used to handle the database connection and queries.
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"    // Driver for postgresql
	_ "modernc.org/sqlite" // Driver for sqlite3
)

// DBConnect returns a pointer to a database connection.
//
// It connects to the database using [sql.Open] with the database type and the connection string,
// it then tests the connection before returning it.
func DBConnect(dbType, dbURL, dbUser, dbPass, dbHost, dbPort, dbName string) (*sql.DB, error) {
	var err error
	var dbase *sql.DB
	switch {
	case dbURL != "":
		// Connect to the database
		dbase, err = sql.Open(dbType, dbURL)
		if err != nil {
			return nil, err
		}

		// Test the connection
		err = dbase.Ping()

		// If the postgres server doesn't support ssl, disable it
		if errors.Is(err, pq.ErrSSLNotSupported) {
			dbURL += "?sslmode=disable"
			dbase, err = sql.Open(dbType, dbURL)
			if err != nil {
				return nil, err
			}

			err = dbase.Ping()
		}
	case dbURL == "":
		dbURL = fmt.Sprintf(
			"user=%s dbname=%s password=%s host=%s port=%s",
			dbUser,
			dbName,
			dbPass,
			dbHost,
			dbPort,
		)
		// Connect to the database
		dbase, err = sql.Open(dbType, dbURL)
		if err != nil {
			return nil, err
		}

		// Test the connection
		err = dbase.Ping()

		// If the postgres server doesn't support ssl, disable it
		if errors.Is(err, pq.ErrSSLNotSupported) {
			dbURL += " sslmode=disable"
			dbase, err = sql.Open(dbType, dbURL)
			if err != nil {
				return nil, err
			}

			err = dbase.Ping()
		}
	}

	return dbase, err
}
