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
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database" // Local database package
	"github.com/redds-be/reddlinks/internal/json"
	"github.com/redds-be/reddlinks/internal/utils"
	"gitlab.gnous.eu/ada/atp"
)

// APIRedirectToURL redirects the client to the URL corresponding to given shortened link.
func (conf Configuration) APIRedirectToURL( //nolint:funlen,cyclop
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the requested short
	requestedShort := req.PathValue("short")

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(conf.DB, requestedShort)
	if err != nil {
		json.RespondWithError(
			writer,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
		)

		return
	}

	if hash != "" {
		// Decode the JSON, client error if it can't, most likely an invalid syntax or no password given at all
		isJSON := false
		for _, contentType := range req.Header["Content-Type"] {
			if contentType == "application/json" {
				isJSON = true

				break
			}
		}

		// If the client is sending json, decode it and set it as the password,
		// Else if the client is sending a parameter, use its value as the password,
		// Else if the client gives nothing, probably a browser, let a front handler handle that.
		var password string
		switch {
		case isJSON:
			params, err := utils.DecodeJSON(req)
			if err != nil {
				json.RespondWithError(
					writer,
					http.StatusBadRequest,
					"Wrong JSON or no password has been given. This link requires a password to access it.",
				)

				return
			}
			password = params.Password
		case req.URL.Query().Get("pass") != "":
			password = req.URL.Query().Get("pass")
		default:
			conf.FrontAskForPassword(writer, req)

			return
		}

		// Check if the password matches the hash
		if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
			!match {
			json.RespondWithError(writer, http.StatusBadRequest, "Wrong password has been given.")

			return
		} else if err != nil {
			json.RespondWithError(writer, http.StatusInternalServerError, "Could not compare the password against corresponding hash.")

			return
		}
	}

	// Get the URL
	url, err := database.GetURLByShort(conf.DB, requestedShort)
	if err != nil {
		json.RespondWithError(
			writer,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
		)

		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(writer, req, url, http.StatusSeeOther)
}

// APICreateLink create a link entry in the database using given json parameters.
func (conf Configuration) APICreateLink( //nolint:funlen,cyclop,gocognit
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the JSON parameters
	params, err := utils.DecodeJSON(req)
	if err != nil {
		json.RespondWithError(writer, http.StatusBadRequest, "Invalid JSON syntax.")

		return
	}

	// Check if the url is valid
	isValid, err := regexp.MatchString(`^https?://.*\..*$`, params.URL)
	if err != nil {
		json.RespondWithError(writer, http.StatusInternalServerError, "Unable to check the URL.")

		return
	}
	if !isValid {
		json.RespondWithError(writer, http.StatusBadRequest, "The URL is invalid.")

		return
	}

	// Check if the expiry time, defaults to now + default, if it's not empty, parse the time and add to now
	// ex: '1d1m' = now + 1day + 1 minute
	var expireAt time.Time
	if params.ExpireAfter == "" {
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.DefaultExpiryTime))
	} else {
		expireDuration, err := atp.ParseDuration(params.ExpireAfter)
		if err != nil {
			json.RespondWithError(writer, http.StatusInternalServerError, "Could not parse the given time. Should look like '1d2H3M4S'")

			return
		}
		expireAt = time.Now().UTC().Add(expireDuration)
	}

	// Check the length, will default to 6 if it's inferior or equal to 0 or will default to 16 if it's over 16
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
			json.RespondWithError(
				writer,
				http.StatusInternalServerError,
				"Could not check the path.",
			)

			return
		}
		if reservedMatch {
			json.RespondWithError(writer, http.StatusBadRequest, fmt.Sprintf(
				"The path '/%s' is reserved.",
				params.Path,
			))

			return
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
				json.RespondWithError(writer, http.StatusBadRequest, fmt.Sprintf(
					"The character '%s' is not allowed.",
					char,
				))

				return
			}
		}
	}

	// Check the path, will default to a randomly generated one with specified length,
	// if its length is over 16, it will be trimmed
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
			json.RespondWithError(
				writer,
				http.StatusInternalServerError,
				"Could not hash the password.",
			)

			return
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
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
		json.RespondWithError(
			writer,
			http.StatusBadRequest,
			"Could not add link: the path is probably already in use.",
		)

		return
	} else if err != nil && autoGen {
	loop:
		for index := conf.DefaultShortLength; index <= conf.DefaultMaxShortLength; index++ {
			params.Path = utils.GenStr(index, allowedChars)
			err = database.CreateLink(conf.DB, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
			switch {
			case err != nil && index == conf.DefaultMaxShortLength:
				json.RespondWithError(writer, http.StatusInternalServerError, "No more space left in the database.")

				return
			case err == nil && index != params.Length:
				type informationResponse struct {
					Information string `json:"information"`
				}
				json.RespondWithJSON(writer, http.StatusContinue, informationResponse{Information: "The length of your auto-generated path" +
					" had to be changed due to space limitations in the database."})

				break loop
			case err == nil:
				break loop
			}
		}
	}

	// Define what is to be returned to the user as a successful response
	link := database.Link{
		ExpireAt: expireAt,
		URL:      params.URL,
		Short:    params.Path,
	}

	// Return the expiry time, the url and the short to the user
	json.RespondWithJSON(writer, http.StatusCreated, link)
}
