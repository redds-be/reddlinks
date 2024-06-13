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
	"database/sql"
	"embed"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
)

// Configuration defines what is going to be sent to the handlers.
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
type Parameters struct {
	URL         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"customPath"`
	ExpireAfter string `json:"expireAfter"`
	ExpireDate  string `json:"expireDate"`
	Password    string `json:"password"`
}

// CollectGarbage deletes old expired entries in the database.
func (conf Configuration) CollectGarbage() error {
	log.Println("Collecting garbage...")

	// Delete expired links
	err := database.RemoveExpiredLinks(conf.DB)
	if err != nil {
		return err
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

// GenStr generates strings of X length compose of Y characters.
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
