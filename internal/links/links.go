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

// Package links handles links.
package links

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/utils"
	"gitlab.gnous.eu/ada/atp"
)

// Link defines the structure of a link entry.
//
// ExpireAt is the date at which the link will expire,
// URL is the original URL,
// Short is the shortened path.
type Link struct {
	ExpireAt time.Time `json:"expireAt"`
	URL      string    `json:"url"`
	Short    string    `json:"short"`
}

// SimpleJSONLink defines the structure of a link entry that will be served to the client in json.
//
// ShortenedLink is the full shortened link,
// ExpireAt is the formatted date at which the link will expire,
// URL is the original URL.
type SimpleJSONLink struct {
	ShortenedLink string `json:"shortenedLink"`
	ExpireAt      string `json:"expireAt"`
	URL           string `json:"url"`
}

// Link defines the structure of a link entry that will be served to the client in json.
//
// ShortenedLink is the full shortened link,
// Password is the password needed to access the url,
// ExpireAt is the formatted date at which the link will expire,
// URL is the original URL.
type PassJSONLink struct {
	ShortenedLink string `json:"shortenedLink"`
	Password      string `json:"password"`
	ExpireAt      string `json:"expireAt"`
	URL           string `json:"url"`
}

// Configuration redefines [utils.Configuration] to be used for methods within the package.
type Configuration utils.Configuration

// NewAdapter returns a configuration to be used by the link handling functions.
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

// CreateLink returns a [Link] struct along with an HTTP code, optional information and an optional error code.
//
// It checks using a regexp if the URL from the payload has http/https as its protocol, it then checks the expiration time,
// if there is none, DefaultExpiryTime will be added to now, if there's one, the time will be parsed using
// [atp.ParseDuration] and this time will be added to now. If there's a specific expiration date provided, it will be used in priority.
// The length provided will be checked and fixed according to min and max settings. the custom path provided will be checked if there's one,
// endpoints and some characters are blacklisted, if the path exceeds the length of DefaultMaxCustomLength,
// it will be trimmed. If there's no custom path provided, a random one will generated using either DefaultShortLength or
// the provided length with [utils.GenStr]. If there's a password provided, it will be hashed using [argon2id.CreateHash].
// After all is done, a link entry will be created in the database using [database.CreateLink].
// If there's an error when creating a link entry using a generated short, it will be re-generated again and again until it works.
func (conf Configuration) CreateLink( //nolint:funlen,gocognit,cyclop,gocyclo
	params utils.Parameters,
) (Link, int, string, string) {
	// Check if the url is valid
	isValid, err := regexp.MatchString(`^https?://.*\..*$`, params.URL)
	if err != nil {
		return Link{}, http.StatusInternalServerError, "", "Unable to check the URL."
	}
	if !isValid {
		return Link{}, http.StatusBadRequest, "", "The URL is invalid."
	}

	// Set the expiry date, if there is none and the default expiry time is 0, the time will be set to the max for sqlite,
	// if there is none and there is a default expiry time, add the default expiry time to now, if there is a "1d2h3m4s" time format,
	// parse it and add it to now, if there's a date, parse the date ans set it as the expiry date
	var expireAt time.Time
	switch {
	case params.ExpireAfter == "" && conf.DefaultExpiryTime == 0 && params.ExpireDate == "":
		expireAt, err = time.Parse("2006-01-02", "9999-12-31")
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Unable to tell when the end of the world will be."
		}
	case params.ExpireAfter == "" && params.ExpireDate == "":
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.DefaultExpiryTime))
	case params.ExpireAfter != "" && params.ExpireDate == "":
		expireDuration, err := atp.ParseDuration(params.ExpireAfter)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Could not parse the given time. Should look like '1d2h3m4s'."
		}
		expireAt = time.Now().UTC().Add(expireDuration)
	case params.ExpireDate != "" && params.ExpireAfter == "":
		expireAt, err = time.Parse("2006-01-02T15:04", params.ExpireDate)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Unable to parse the expiry date."
		}
	case params.ExpireDate != "" && params.ExpireAfter != "":
		expireAt, err = time.Parse("2006-01-02T15:04", params.ExpireDate)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Unable to parse the expiry date."
		}
	}

	// Check the length, will default to DefaultShortLength,
	// if it's inferior or equal to 0 or will default to DefaultMaxShortLength if it's over DefaultMaxShortLength
	if params.Length <= 0 {
		params.Length = conf.DefaultShortLength
	} else if params.Length > conf.DefaultMaxShortLength {
		params.Length = conf.DefaultMaxShortLength
	}

	// Check the validity of a custom path
	if params.Path != "" {
		// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access, privacy and assets are used for the front.
		reservedMatch, err := regexp.MatchString(
			`^status$|^error$|^add$|^access$|^privacy$|^assets.*$`,
			params.Path,
		)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Could not check the path."
		}
		if reservedMatch {
			return Link{}, http.StatusBadRequest, "", fmt.Sprintf(
				"The path '/%s' is reserved.",
				params.Path,
			)
		}

		// Check the validity of the custom path
		reservedChars := []string{
			":",
			"/",
			"?",
			"#",
			"[",
			"]",
			"@",
			"!",
			"$",
			"&",
			"'",
			"(",
			")",
			"*",
			"+",
			",",
			";",
			"=",
		}
		for _, char := range reservedChars {
			if match := strings.Contains(params.Path, char); match {
				return Link{}, http.StatusBadRequest, "", fmt.Sprintf(
					"The character '%s' is not allowed.",
					char,
				)
			}
		}
	}

	// Check the path, will default to a randomly generated one with specified length,
	// if its length is over DefaultMaxCustomLength, it will be trimmed
	autoGen := false
	allowedChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	if params.Path == "" {
		autoGen = true
		params.Path = utils.GenStr(params.Length, allowedChars)
	}
	if len(params.Path) > conf.DefaultMaxCustomLength {
		params.Path = params.Path[:conf.DefaultMaxCustomLength]
	}

	// If the password given to by the request isn't null (meaning no password), generate an argon2 hash from it
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", "Could not hash the password."
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	addInfo := ""
	err = database.CreateLink(
		conf.DB,
		uuid.New(),
		time.Now().UTC(),
		expireAt,
		params.URL,
		params.Path,
		hash,
	)
	if err != nil && !autoGen {
		return Link{}, http.StatusBadRequest, "", "Could not shorten the URL: the path is probably already in use."
	} else if err != nil && autoGen {
	loop:
		for index := conf.DefaultShortLength; index <= conf.DefaultMaxShortLength; index++ {
			params.Path = utils.GenStr(index, allowedChars)
			err = database.CreateLink(conf.DB, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
			switch {
			case err != nil && index == conf.DefaultMaxShortLength:
				return Link{}, http.StatusInternalServerError, "", "No more space left in the database."
			case err == nil && index != params.Length:
				addInfo = "The length of your auto-generated path had to be changed due to space limitations in the database."

				break loop
			case err == nil:
				break loop
			}
		}
	}

	// Return the necessary information
	link := Link{
		ExpireAt: expireAt,
		URL:      params.URL,
		Short:    params.Path,
	}

	return link, http.StatusCreated, addInfo, ""
}
