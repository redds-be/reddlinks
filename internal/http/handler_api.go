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

package http

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/redds-be/reddlinks/internal/database" // Local database package
	"github.com/redds-be/reddlinks/internal/json"
	"github.com/redds-be/reddlinks/internal/links"
	"github.com/redds-be/reddlinks/internal/utils"
)

// APIRedirectToURL redirects the client to the URL corresponding to given shortened link.
//
// It first starts by getting the short from the request (GET /{short}),
// then it checks if there's a '+' at the end, meaning an info request. If it is, the '+' is trimmed and then
// its password's hash using [database.GetHashByShort], if there is one,
// it firsts checks if there's a json payload to get a password from,
// if not, redirect to /access handled by FrontAskForPassword which is going to ask for a password using a form.
// Once the JSON payload is decoded using [utils.DecodeJSON], if there's a password, its hash will be compared to the hash corresponding
// to the short using [argon2id.ComparePasswordAndHash], if it's the case,
// if it's an info request, the information associated with the URL is sent to the client, without redirection, if not, the client will be redirected.
// If there's no hash associated with the short,
// if it's an info request, the information associated with the URL is sent to the client, without redirection, if not, the client will be redirected.
func (conf Configuration) APIRedirectToURL( //nolint:funlen,cyclop
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the locale
	locale := utils.GetLocale(req, utils.Configuration(conf))

	// Get the requested short
	requestedShort := req.PathValue("short")

	// Check for a '+' at the end of the short, indicating an info request.
	infoRequest := strings.HasSuffix(requestedShort, "+")

	// Remove the '+' at the end so that it can be
	if infoRequest {
		requestedShort = requestedShort[:len(requestedShort)-1]
	}

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(conf.DB, requestedShort)
	if err != nil {
		conf.RespondWithError(
			writer,
			req,
			http.StatusNotFound,
			locale.ErrNotFound,
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
				conf.RespondWithError(
					writer,
					req,
					http.StatusBadRequest,
					locale.ErrPassAccess,
				)

				return
			}
			password = params.Password
		case req.URL.Query().Get("pass") != "":
			password = req.URL.Query().Get("pass")
		default:
			conf.FrontAskForPassword(writer, req, infoRequest)

			return
		}

		// Check if the password matches the hash
		if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
			!match {
			conf.RespondWithError(writer, req, http.StatusBadRequest, "Wrong password has been given.")

			return
		} else if err != nil {
			conf.RespondWithError(writer, req, http.StatusInternalServerError, "Could not compare the password against corresponding hash.")

			return
		}
	}

	// If it's an info request, we send the info
	if infoRequest {
		// CLI clients usually use only '*/*' whilst web browser typically uses a list containing both '*/*' and "text/html".
		// if "text/html" is present, it is safe to assume it's a web browser
		if strings.Contains(req.Header.Get("Accept"), "text/html") {
			conf.FrontHandlerURLInfo(writer, req, requestedShort)

			return
		}

		// Get the information
		url, createdAt, expireAt, err := database.GetURLInfo(conf.DB, requestedShort)
		if err != nil {
			conf.RespondWithError(
				writer,
				req,
				http.StatusInternalServerError,
				locale.ErrGetInfo,
			)

			return
		}

		// Send the information to the client
		json.RespondWithJSON(writer, http.StatusOK, json.InfoResponse{
			DstURL:    url,
			Short:     requestedShort,
			CreatedAt: createdAt.Format(time.RFC822),
			ExpiresAt: expireAt.Format(time.RFC822),
		})

		return
	}

	// Get the URL
	url, err := database.GetURLByShort(conf.DB, requestedShort)
	if err != nil {
		conf.RespondWithError(
			writer,
			req,
			http.StatusNotFound,
			locale.ErrNotFound,
		)

		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(writer, req, url, http.StatusSeeOther)
}

// APICreateLink creates a link entry in the database using given json parameters.
//
// It firsts decodes the JSON payload from the client using [utils.DecodeJSON], hen creates an adapter for links using
// [links.NewAdapter] then calls [links.CreateLink] to create a link entry giving it the deocded params,
// the information is then formatted and returned to the client with two versions, [links.PassJSONLink] if there's a password and
// [links.SimpleJSONLink] if there's not a password.
func (conf Configuration) APICreateLink( //nolint:funlen
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the locale
	locale := utils.GetLocale(req, utils.Configuration(conf))

	// Get the JSON parameters
	params, err := utils.DecodeJSON(req)
	if err != nil {
		conf.RespondWithError(writer, req, http.StatusBadRequest, locale.ErrInvalidJSON)

		return
	}

	// Create a configuration struct for the links adapter
	linksConf := utils.Configuration{
		DB:                     conf.DB,
		InstanceName:           conf.InstanceName,
		InstanceURL:            conf.InstanceURL,
		Version:                conf.Version,
		AddrAndPort:            conf.AddrAndPort,
		DefaultShortLength:     conf.DefaultShortLength,
		DefaultMaxShortLength:  conf.DefaultMaxShortLength,
		DefaultMaxCustomLength: conf.DefaultMaxCustomLength,
		DefaultExpiryTime:      conf.DefaultExpiryTime,
		ContactEmail:           conf.ContactEmail,
		Static:                 conf.Static,
	}

	// Create an adapter using the configuration struct
	linksAdapter := links.NewAdapter(linksConf)

	// Create the link entry
	link, code, addInfo, errMsg := linksAdapter.CreateLink(params, locale)
	if errMsg != "" {
		conf.RespondWithError(writer, req, code, errMsg)

		return
	}

	// If there's additional information, display it
	if addInfo != "" {
		type informationResponse struct {
			Information string `json:"information"`
		}
		json.RespondWithJSON(writer, http.StatusContinue, informationResponse{Information: addInfo})
	}

	// Format the shortened link
	shortenedLink := regexp.MustCompile("^https://|http://").
		ReplaceAllString(fmt.Sprintf("%s%s", conf.InstanceURL, link.Short), "")

	// Format the expiration date that will be displayed to the user
	var expireAt string
	if params.ExpireDate == "" && params.ExpireAfter == "" && conf.DefaultExpiryTime == 0 {
		expireAt = "Never"
	} else {
		expireAt = link.ExpireAt.Format(time.RFC822)
	}

	// If there's a password return links.PassJSONLink, if there's none return links.SimpleJSONLink
	if params.Password != "" {
		linkToReturn := links.PassJSONLink{
			ShortenedLink: shortenedLink,
			Password:      params.Password,
			ExpireAt:      expireAt,
			URL:           link.URL,
		}

		// Return the expiry time, the url and the short to the user
		json.RespondWithJSON(writer, code, linkToReturn)
	} else {
		linkToReturn := links.SimpleJSONLink{
			ShortenedLink: shortenedLink,
			ExpireAt:      expireAt,
			URL:           link.URL,
		}

		// Return the expiry time, the url and the short to the user
		json.RespondWithJSON(writer, code, linkToReturn)
	}
}

// RespondWithError sends an appropriate error response to the client based on the
// Accept header in the request. If the Accept header contains "text/html", it renders
// an HTML error page using FrontErrorPage. Otherwise, it responds with a JSON error
// using json.RespondWithError.
//
// Parameters:
//   - writer: The http.ResponseWriter to write the response to
//   - req: The incoming HTTP request
//   - code: The HTTP status code to return
//   - errMsg: The error message to display
func (conf Configuration) RespondWithError(writer http.ResponseWriter, req *http.Request, code int, errMsg string) {
	// CLI clients usually use only '*/*' whilst web browser typically uses a list containing both '*/*' and "text/html".
	// if "text/html" is present, it is safe to assume it's a web browser, so we give render an error page
	if strings.Contains(req.Header.Get("Accept"), "text/html") {
		conf.FrontErrorPage(writer, req, code, errMsg, "/")

		return
	}

	// Otherwise, respond in JSON
	json.RespondWithError(writer, code, errMsg)
}
