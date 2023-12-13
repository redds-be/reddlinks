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
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database" // Local database package
)

func (conf configuration) apiRedirectToURL(writer http.ResponseWriter, req *http.Request) { //nolint:funlen,cyclop
	log.Printf("%s %s", req.Method, req.URL.Path)

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(conf.db, trimFirstRune(req.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(
			writer,
			req,
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
			params, err := decodeJSON(req)
			if err != nil {
				log.Println(err)
				respondWithError(
					writer,
					req,
					http.StatusBadRequest,
					"Wrong JSON or no password has been given. This link requires a password to access it.",
				)

				return
			}
			password = params.Password
		case req.URL.Query().Get("pass") != "":
			password = req.URL.Query().Get("pass")
		default:
			conf.frontAskForPassword(writer, req)

			return
		}

		// Check if the password matches the hash
		if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
			!match {
			respondWithError(writer, req, http.StatusBadRequest, "Wrong password has been given.")

			return
		} else if err != nil {
			log.Println(err)
			respondWithError(writer, req, http.StatusInternalServerError, "Could not compare the password against corresponding hash.")

			return
		}
	}

	// Get the URL
	url, err := database.GetURLByShort(conf.db, trimFirstRune(req.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(
			writer,
			req,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
		)

		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(writer, req, url, http.StatusSeeOther)
}

func (conf configuration) apiCreateLink( //nolint:funlen,cyclop,gocognit
	writer http.ResponseWriter,
	params parameters,
) (database.Link, int, string) {
	// Check if the url is valid
	isValid, err := regexp.MatchString(`^https?://.*\..*$`, params.URL)
	if err != nil {
		log.Println(err)

		return database.Link{}, http.StatusInternalServerError, "Unable to check the URL."
	}
	if !isValid {
		return database.Link{}, http.StatusBadRequest, "The URL is invalid."
	}

	// Check the expiration time and set it to x minute specified by the user, -1 = never, will default to 48 hours
	var expireAt time.Time
	switch {
	case params.ExpireAfter == -1:
		expireAt = time.Date(9999, 12, 31, 23, 59, 59, 59, time.UTC)
	case params.ExpireAfter <= 0:
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.defaultExpiryTime))
	default:
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(params.ExpireAfter))
	}

	// Check the length, will default to 6 if it's inferior or equal to 0 or will default to 16 if it's over 16
	if params.Length <= 0 {
		params.Length = conf.defaultShortLength
	} else if params.Length > conf.defaultMaxShortLength {
		params.Length = conf.defaultMaxShortLength
	}

	// Check the validity of a custom path
	if params.Path != "" {
		// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access, privacy and assets are used for the front.
		reservedMatch, err := regexp.MatchString(
			`^status$|^error$|^add$|^access$|^privacy$|^assets.*$`,
			params.Path,
		)
		if err != nil {
			log.Println(err)

			return database.Link{}, http.StatusInternalServerError, "Could not check the path."
		}
		if reservedMatch {
			return database.Link{}, http.StatusBadRequest, fmt.Sprintf(
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
				return database.Link{}, http.StatusBadRequest, fmt.Sprintf(
					"The character '%s' is not allowed.",
					char,
				)
			}
		}
	}

	// Check the path, will default to a randomly generated one with specified length,
	// if its length is over 16, it will be trimmed
	autoGen := false
	allowedChars := []byte(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~",
	)
	if params.Path == "" {
		autoGen = true
		params.Path = uniuri.NewLenChars(params.Length, allowedChars)
	}
	if len(params.Path) > conf.defaultMaxCustomLength {
		params.Path = params.Path[:conf.defaultMaxCustomLength]
	}

	// If the password given to by the request isn't null (meaning no password), generate an argon2 hash from it
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)

			return database.Link{}, http.StatusInternalServerError, "Could not hash the password."
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	err = database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
	if err != nil && !autoGen {
		log.Println(err)

		return database.Link{}, http.StatusBadRequest, "Could not add link: the path is probably already in use."
	} else if err != nil && autoGen {
	loop:
		for index := conf.defaultShortLength; index <= conf.defaultMaxShortLength; index++ {
			params.Path = uniuri.NewLenChars(index, allowedChars)
			err = database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
			switch {
			case err != nil:
				log.Println(err)
			case err != nil && index == conf.defaultMaxShortLength:
				return database.Link{}, http.StatusInternalServerError, "No more space left in the database."
			case err == nil && index != params.Length:
				type informationResponse struct {
					Information string `json:"information"`
				}
				respondWithJSON(writer, http.StatusContinue, informationResponse{Information: "The length of your auto-generated path" +
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
	return link, 0, ""
}

func (conf configuration) apiHandlerRoot(writer http.ResponseWriter, req *http.Request) {
	// Check method and decide whether to create or redirect to link
	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/favicon.ico":
		return
	case req.Method == http.MethodGet && req.URL.Path == "/":
		conf.frontHandlerMainPage(writer, req)
	case req.Method == http.MethodGet:
		conf.apiRedirectToURL(writer, req)
	case req.Method == http.MethodPost:
		log.Printf("%s %s", req.Method, req.URL.Path)
		params, err := decodeJSON(req)
		if err != nil {
			respondWithError(writer, req, http.StatusBadRequest, "Invalid JSON syntax.")

			return
		}
		link, code, errMsg := conf.apiCreateLink(writer, params)
		if errMsg != "" {
			respondWithError(writer, req, code, errMsg)

			return
		}
		respondWithJSON(writer, http.StatusCreated, link)
	default:
		respondWithError(writer, req, http.StatusMethodNotAllowed, "Method Not Allowed.")

		return
	}
}
