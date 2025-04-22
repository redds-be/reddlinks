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

// Package links handles links.
package links

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/utils"
	"gitlab.gnous.eu/ada/atp"
)

// Common validation patterns compiled once for reuse.
var (
	urlPattern    = regexp.MustCompile(`^https?://.*\..*$`)
	reservedPaths = regexp.MustCompile(`^status$|^error$|^add$|^access$|^privacy$|^assets.*$`)
	alphaNumeric  = regexp.MustCompile(`^[A-Za-z0-9]*$`)
	protocolRegex = regexp.MustCompile(`^https://|http://`)
)

// Link defines the structure of a link entry.
type Link struct {
	// ExpireAt is the date at which the link will expire
	ExpireAt time.Time `json:"expireAt"`
	// URL is the original URL
	URL string `json:"url"`
	// Short is the shortened path
	Short string `json:"short"`
}

// SimpleJSONLink defines the structure of a link entry that will be served to the client in JSON.
type SimpleJSONLink struct {
	// ShortenedLink is the full shortened link
	ShortenedLink string `json:"shortenedLink"`
	// ExpireAt is the formatted date at which the link will expire
	ExpireAt string `json:"expireAt"`
	// URL is the original URL
	URL string `json:"url"`
}

// PassJSONLink defines the structure of a link entry with password that will be served to the client in JSON.
type PassJSONLink struct {
	// ShortenedLink is the full shortened link
	ShortenedLink string `json:"shortenedLink"`
	// Password is the password needed to access the url
	Password string `json:"password"`
	// ExpireAt is the formatted date at which the link will expire
	ExpireAt string `json:"expireAt"`
	// URL is the original URL
	URL string `json:"url"`
}

// Configuration redefines utils.Configuration to be used for methods within the package.
type Configuration utils.Configuration

// NewAdapter returns a configuration to be used by the link handling functions.
//
// It takes a utils.Configuration and returns a Configuration specific to this package.
func NewAdapter(configuration utils.Configuration) Configuration {
	return Configuration(configuration)
}

// CreateLink returns a Link struct along with an HTTP code, optional information and an optional error code.
//
// It performs the following validations and operations:
//   - Validates URL format (must use http/https protocol)
//   - Determines link expiration time based on provided parameters or defaults
//   - Validates or generates a path for the shortened URL
//   - Prevents creation of redirection loops
//   - Hashes passwords if provided for protected links
//   - Creates the link entry in the database, handling collisions for generated paths
//
// Parameters:
//   - params: Contains all link creation parameters (URL, path, expiry, etc.)
//   - locale: Contains localized text messages for error reporting
//
// Returns:
//   - Link: The created link structure (empty if error occurred)
//   - int: HTTP status code
//   - string: Additional information message (if any)
//   - string: Error message (if any)
func (conf *Configuration) CreateLink( //nolint:gocognit,gocyclo,cyclop,funlen
	params utils.Parameters,
	locale utils.PageLocaleTl,
) (Link, int, string, string) {
	// Check if the url is valid using pre-compiled regex
	isValid := urlPattern.MatchString(params.URL)
	if !isValid {
		return Link{}, http.StatusBadRequest, "", locale.ErrInvalidURL
	}

	// Set the expiry date, handling different expiration scenarios
	var expireAt time.Time
	var err error

	switch {
	case params.ExpireAfter == "" && conf.DefaultExpiryTime == 0 && params.ExpireDate == "":
		// No expiration specified and no default - use max date
		expireAt, err = time.Parse("2006-01-02", "9999-12-31")
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", locale.ErrUnableTellEOW
		}
	case params.ExpireAfter == "" && params.ExpireDate == "":
		// Use default expiration time
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.DefaultExpiryTime))
	case params.ExpireAfter != "" && params.ExpireDate == "":
		// Parse and use custom duration
		expireDuration, err := atp.ParseDuration(params.ExpireAfter)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", locale.ErrParseTime
		}
		expireAt = time.Now().UTC().Add(expireDuration)
	case params.ExpireDate != "":
		// Parse and use explicit expiration date (priority over duration)
		expireAt, err = time.Parse("2006-01-02T15:04", params.ExpireDate)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", locale.ErrParseExpiry
		}
	}

	// Adjust length parameter to be within valid bounds
	if params.Length <= 0 {
		params.Length = conf.DefaultShortLength
	} else if params.Length > conf.DefaultMaxShortLength {
		params.Length = conf.DefaultMaxShortLength
	}

	// Process custom path or generate a random one
	autoGen := false
	allowedChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	if params.Path != "" { //nolint:nestif
		// Check if the path is reserved
		reservedMatch := reservedPaths.MatchString(params.Path)
		if reservedMatch {
			return Link{}, http.StatusBadRequest, "", fmt.Sprintf(
				"The path '/%s' is reserved.",
				params.Path,
			)
		}

		// Check if path contains only alphanumeric characters
		specialCharMatch := alphaNumeric.MatchString(params.Path)
		if !specialCharMatch {
			return Link{}, http.StatusBadRequest, "", locale.ErrAlphaNumeric
		}

		// Trim path if it exceeds maximum length
		if len(params.Path) > conf.DefaultMaxCustomLength {
			params.Path = params.Path[:conf.DefaultMaxCustomLength]
		}
	} else {
		// Generate random path
		autoGen = true
		params.Path, err = utils.GenStr(params.Length, allowedChars)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", locale.ErrUnableGen
		}
	}

	// Check for redirection loops
	normalizedOriginal := protocolRegex.ReplaceAllString(params.URL, "")
	normalizedShortened := protocolRegex.ReplaceAllString(fmt.Sprintf("%s%s", conf.InstanceURL, params.Path), "")
	if normalizedOriginal == normalizedShortened {
		return Link{}, http.StatusBadRequest, "", locale.ErrRedirectionLoop
	}

	// Hash password if provided
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			return Link{}, http.StatusInternalServerError, "", locale.ErrPathInUse
		}
	}

	// Create link in database
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

	// Handle collision for custom path
	if err != nil && !autoGen {
		return Link{}, http.StatusBadRequest, "", locale.ErrPathInUse
	} else if err != nil && autoGen {
		// Handle collision for auto-generated path by trying different lengths
	loop:
		for index := conf.DefaultShortLength; index <= conf.DefaultMaxShortLength; index++ {
			params.Path, err = utils.GenStr(index, allowedChars)
			if err != nil {
				return Link{}, http.StatusInternalServerError, "", locale.ErrUnableGen
			}

			err = database.CreateLink(
				conf.DB,
				uuid.New(),
				time.Now().UTC(),
				expireAt,
				params.URL,
				params.Path,
				hash,
			)

			switch {
			case err != nil && index == conf.DefaultMaxShortLength:
				return Link{}, http.StatusInternalServerError, "", locale.ErrNoSpaceLeft
			case err == nil && index != params.Length:
				addInfo = locale.InfoLengthChange

				break loop
			case err == nil:
				break loop
			}
		}
	}

	// Return the created link
	link := Link{
		ExpireAt: expireAt,
		URL:      params.URL,
		Short:    params.Path,
	}

	return link, http.StatusCreated, addInfo, ""
}
