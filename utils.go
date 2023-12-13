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
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/redds-be/rlinks/database"
)

type configuration struct {
	// Define what is going to be sent to the handlers
	db                     *sql.DB
	instanceName           string
	instanceURL            string
	defaultShortLength     int
	defaultMaxShortLength  int
	defaultMaxCustomLength int
	defaultExpiryTime      int
}

type parameters struct {
	// Define the structure of the JSON payload that will be read from the user
	URL         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"customPath"`
	ExpireAfter int    `json:"expireAfter"`
	Password    string `json:"password"`
}

func trimFirstRune(s string) string {
	// Remove the first letter of a string (https://go.dev/play/p/ZOZyRORkK82)
	_, i := utf8.DecodeRuneInString(s)

	return s[i:]
}

func (conf configuration) collectGarbage(timeBetweenCleanups int) {
	// Just some kind of hack to call the manual garbage collecting function every minute
	for {
		log.Println("Collecting garbage...")
		// Get the links
		links, err := database.GetLinks(conf.db)
		if err != nil {
			log.Println(err)

			return
		}

		// Go through the link and delete expired ones
		now := time.Now().UTC()
		for _, link := range links {
			if now.After(link.ExpireAt) || now.Equal(link.ExpireAt) {
				log.Printf(
					"Link : %s is expired, deleting it...", link.Short)
				err := database.RemoveLink(conf.db, link.Short)
				if err != nil {
					log.Printf(
						"Could not remove Link: %s", link.Short)

					return
				}
			}
		}
		// Wait for length of time in minutes specified in .env
		time.Sleep(time.Duration(timeBetweenCleanups) * time.Minute)
	}
}

func decodeJSON(r *http.Request) (parameters, error) {
	// Decode the JSON from the client's request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	return params, err
}
