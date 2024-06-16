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

// Package utils implements functions and structs that does not need their own package.
package utils

import (
	"database/sql"
	"embed"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
)

// Configuration defines what is going to be sent to the handlers.
//
// DB is a pointer to the database connection,
// InstanceName refers to the name of the reddlinks instance,
// InstanceURL refers to the public URL of the reddlinks instance,
// Version refers to the actual version of the reddlinks instance,
// AddrAndPort refers to the listening port and address of the reddlinks instance,
// DefaultShortLength refers to the default length of generated strings for a short URL,
// DefaultMaxShortLength refers to the maximum length of generated strings for a short URL,
// DefaultMaxCustomLength refers to the maximum length of custom strings for a short URL,
// DefaultExpiryTime refers to the default expiry time of links records,
// ContactEmail refers to an optional admin's contact email,
// Static contains the embedded static filesystem.
type Configuration struct {
	DB                     *sql.DB
	InstanceName           string
	InstanceURL            string
	Version                string
	AddrAndPort            string
	DefaultShortLength     int
	DefaultMaxShortLength  int
	DefaultMaxCustomLength int
	DefaultExpiryTime      int
	ContactEmail           string
	Static                 embed.FS
}

// Parameters defines the structure of the JSON payload that will be read from the user.
//
// URL is the URL to shorten,
// Length is the length of the string that will be generated,
// Path refers to the custom string used in the shortened URL,
// ExpireAfter refers the time from now after which the link will expire,
// ExpireDate refers to the exact expiration date for the link,
// Password refers to a password to protect a link from being accessed by anybody.
type Parameters struct {
	URL         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"customPath"`
	ExpireAfter string `json:"expireAfter"`
	ExpireDate  string `json:"expireDate"`
	Password    string `json:"password"`
}

// CollectGarbage deletes old expired entries in the database.
//
// It calls [database.RemoveExpiredLinks] which will delete expired links.
// As of now, the necessity of this function is questionable.
func (conf Configuration) CollectGarbage() error {
	// Delete expired links
	err := database.RemoveExpiredLinks(conf.DB)
	if err != nil {
		return err
	}

	return nil
}

// DecodeJSON returns a [utils.Parameters] struct that contains the decoded clients's JSON request.
//
// It creates a decoder using [json.NewDecoder], using this decoder,
// the function decodes the client's JSON and store it in the [utils.Parameters] struct to then be returned.
// As of now, the necessity of keeping the function in utils rather json is questionable.
func DecodeJSON(r *http.Request) (Parameters, error) {
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	return params, err
}

// GenStr returns a string of a set length composed of a specific charset.
//
// It first creates a byte map of a set length, then, for the length of the map,
// select a random char from the charset to be added the map at the actual index of the iteration.
// After all is done, the map is converted into a string while being returned.
func GenStr(length int, charset string) string {
	// Create an empty map for the future string
	randomByteStr := make([]byte, length)

	// For the length of the empty string, append a random character within the charset
	for i := range randomByteStr {
		randomByteStr[i] = charset[rand.New( //nolint:gosec
			rand.NewSource(time.Now().UnixNano())).Intn(len(charset))]
	}

	// Convert and return the generated string
	return string(randomByteStr)
}
