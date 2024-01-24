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

package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/redds-be/reddlinks/internal/database"
)

// Configuration defines what is going to be sent to the handlers.
type Configuration struct {
	DB                     *sql.DB
	InstanceName           string
	InstanceURL            string
	Version                string
	PortSTR                string
	DefaultShortLength     int
	DefaultMaxShortLength  int
	DefaultMaxCustomLength int
	DefaultExpiryTime      int
}

// Parameters defines the structure of the JSON payload that will be read from the user.
type Parameters struct {
	URL         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"customPath"`
	ExpireAfter int    `json:"expireAfter"`
	Password    string `json:"password"`
}

// TrimFirstRune removes the first letter of a string (https://go.dev/play/p/ZOZyRORkK82).
func TrimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)

	return s[i:]
}

// CollectGarbage deletes old expired entries in the database.
func (conf Configuration) CollectGarbage() error {
	log.Println("Collecting garbage...")
	// Get the links
	links, err := database.GetLinks(conf.DB)
	if err != nil {
		return err
	}

	// Go through the link and delete expired ones
	now := time.Now().UTC()
	for _, link := range links {
		if now.After(link.ExpireAt) || now.Equal(link.ExpireAt) {
			log.Printf(
				"Link : %s is expired, deleting it...", link.Short)
			err := database.RemoveLink(conf.DB, link.Short)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DecodeJSON decodes the JSON from the client's request.
func DecodeJSON(r *http.Request) (Parameters, error) {
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	return params, err
}

// RandomToken creates a random token.
func RandomToken() string {
	bytes := make([]byte, 32) //nolint:gomnd
	if _, err := rand.Read(bytes); err != nil {
		log.Println(err)
	}

	return hex.EncodeToString(bytes)
}
