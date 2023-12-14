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
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
)

type Page struct {
	// Define the structure of what can be displayed on a page
	InstanceTitle          string
	InstanceURL            string
	ShortenedLink          string
	Short                  string
	URL                    string
	ExpireAt               string
	Password               string
	Error                  string
	AddInfo                string
	Version                string
	Token                  string
	TimeOfExecution        string
	DefaultShortLength     int
	DefaultMaxShortLength  int
	DefaultMaxCustomLength int
	DefaultExpiryTime      int
}

// Parse the templates files in advance.
var templates = template.Must(template.ParseFiles( //nolint:gochecknoglobals
	"static/index.html",
	"static/add.html",
	"static/error.html",
	"static/pass.html",
	"static/privacy.html"))

func renderTemplate(writer http.ResponseWriter, tmpl string, page any) {
	// Tell that we serve HTML in UTF-8.
	writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	// Tell that all ressources comes from here and that only this site can frame itself
	writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self';"+
		" style-src 'self'; img-src 'self'; connect-src 'self'; frame-src 'self'; font-src 'self'; media-src 'self';"+
		" object-src 'self'; manifest-src 'self'; worker-src 'self'; form-action 'self'; frame-ancestors 'self'")
	// Block access to styles and scripts
	writer.Header().Set("X-Content-Type-Options", "nosniff")

	// Render a given template, json error if it can't
	err := templates.ExecuteTemplate(writer, tmpl+".html", page)
	if err != nil {
		log.Println(err)
		respondWithError(writer, http.StatusInternalServerError, "Unable to load the page.")

		return
	}
}

func (conf configuration) frontErrorPage(
	writer http.ResponseWriter,
	req *http.Request,
	code int,
	errMsg string,
) {
	log.Printf("%s %s", req.Method, req.URL.Path)
	// Set what is going to be displayed on the error page
	page := &Page{
		InstanceTitle: conf.instanceName,
		InstanceURL:   conf.instanceURL,
		Error:         fmt.Sprintf("Error %d: %s", code, errMsg),
		Version:       conf.version,
	}

	// Display the error page
	renderTemplate(writer, "error", page)
}

func (conf configuration) frontHandlerMainPage(writer http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)

	// Set what is going to be displayed on the main page
	page := &Page{
		InstanceTitle: conf.instanceName,
		InstanceURL:   conf.instanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(conf.instanceURL, ""),
		DefaultShortLength:     conf.defaultShortLength,
		DefaultMaxShortLength:  conf.defaultMaxShortLength,
		DefaultMaxCustomLength: conf.defaultMaxCustomLength,
		DefaultExpiryTime:      conf.defaultExpiryTime,
		Version:                conf.version,
		Token:                  token,
		TimeOfExecution:        time.Now().UTC().Format(time.ANSIC),
	}

	// Display the front page
	renderTemplate(writer, "index", page)
}

func (conf configuration) frontHandlerPrivacyPage(writer http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)

	// Set what is going to be displayed on the privacy page
	page := &Page{
		InstanceTitle: conf.instanceName,
		InstanceURL:   conf.instanceURL,
		Version:       conf.version,
	}

	// Display the front page
	renderTemplate(writer, "privacy", page)
}

func (conf configuration) frontCreateLink( //nolint:cyclop,funlen
	params parameters,
) (string, int, string, database.Link) {
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

	if params.Path != "" {
		// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access, privacy and assets are used for the front.
		reservedMatch, err := regexp.MatchString(
			`^status$|^error$|^add$|^access$|^privacy$|^assets.*$`,
			params.Path,
		)
		if err != nil {
			return "The path could not be checked.", http.StatusInternalServerError, "", database.Link{}
		}
		if reservedMatch {
			return fmt.Sprintf(
				"The path '/%s' is reserved.",
				params.Path,
			), http.StatusBadRequest, "", database.Link{}
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
				return fmt.Sprintf(
					"The character '%s' is not allowed.",
					char,
				), http.StatusBadRequest, "", database.Link{}
			}
		}
	}

	// Check the path, will default to a randomly generated one with specified length, if its length is over 16, it will be trimmed
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
	var err error
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)

			return "Could not hash the password.", http.StatusInternalServerError, "", database.Link{}
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	addInfo := ""
	err = database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
	if err != nil && !autoGen {
		log.Println(err)

		return "Could not add link: the path is probably already in use.", http.StatusBadRequest, "", database.Link{}
	} else if err != nil && autoGen {
	loop:
		for index := conf.defaultShortLength; index <= conf.defaultMaxShortLength; index++ {
			params.Path = uniuri.NewLenChars(index, allowedChars)
			err = database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
			switch {
			case err != nil:
				log.Println(err)
			case err != nil && index == conf.defaultMaxShortLength:
				return "No more space left in the database.", http.StatusInternalServerError, "", database.Link{}
			case err == nil && index != params.Length:
				addInfo = "The length of your auto-generated path had to be changed due to space limitations in the database."

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

	return "", 0, addInfo, link
}

func (conf configuration) frontHandlerAdd(writer http.ResponseWriter, req *http.Request) { //nolint:funlen
	log.Printf("%s %s", req.Method, req.URL.Path)

	// What to if the form is correct, i.e. the front page form was posted.
	// If this isn't the case, display an error page
	if req.FormValue("add") != "Add" {
		conf.frontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the form.")

		return
	}

	// Get the time of the request
	execTime, err := time.Parse(time.ANSIC, req.FormValue("_time"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the form.")

		return
	}

	// Check if the given token corresponds to the actual token, if not, probably invalid or expired
	if token != req.FormValue("_token") && time.Now().UTC().Add(15*time.Minute).After(execTime) { //nolint:gomnd
		conf.frontErrorPage(writer, req, http.StatusBadRequest, "Token invalid or expired.")

		return
	}

	// Convert the length to an int, display an error page if it can't
	length, err := strconv.Atoi(req.FormValue("length"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(
			writer,
			req,
			http.StatusInternalServerError,
			"There was an error trying to read the password.",
		)

		return
	}

	// Convert the expiration time to an int, display an error page if it can't
	expireAfter, err := strconv.Atoi(req.FormValue("expire_after"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(
			writer,
			req,
			http.StatusInternalServerError,
			"There was an error trying to read the expiration time.",
		)

		return
	}

	// Set the values that will be used for the link creation
	params := parameters{
		URL:         req.FormValue("url"),
		Length:      length,
		Path:        req.FormValue("short"),
		ExpireAfter: expireAfter,
		Password:    req.FormValue("password"),
	}

	// Create a link entry into the database, display an error page if it can't
	errMsg, code, addInfo, link := conf.frontCreateLink(params)
	if errMsg != "" {
		conf.frontErrorPage(writer, req, code, errMsg)

		return
	}

	// Format the expiration date that will be displayed to the user
	var expireAt string
	if params.ExpireAfter == -1 {
		expireAt = "never"
	} else {
		expireAt = link.ExpireAt.Format(time.ANSIC)
	}

	// Set what is going to be displayed on the add page
	page := &Page{
		InstanceTitle: conf.instanceName,
		InstanceURL:   conf.instanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(fmt.Sprintf("%s%s", conf.instanceURL, link.Short), ""),
		Short:    link.Short,
		URL:      link.URL,
		ExpireAt: expireAt,
		Password: params.Password,
		AddInfo:  addInfo,
		Version:  conf.version,
	}

	// Display the add page which will display the information about the added link
	renderTemplate(writer, "add", page)
}

func (conf configuration) frontAskForPassword(writer http.ResponseWriter, req *http.Request) {
	// Set what is going to be displayed on the pass page
	page := &Page{
		InstanceTitle:   conf.instanceName,
		InstanceURL:     conf.instanceURL,
		Short:           trimFirstRune(req.URL.Path),
		Version:         conf.version,
		Token:           token,
		TimeOfExecution: time.Now().UTC().Format(time.ANSIC),
	}

	// Display the pass page which will ask the user for a password
	renderTemplate(writer, "pass", page)
}

func (conf configuration) frontHandlerRedirectToURL(writer http.ResponseWriter, req *http.Request) { //nolint:funlen
	// Get the time of the request
	execTime, err := time.Parse(time.ANSIC, req.FormValue("_time"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the form.")

		return
	}

	// Check if the given token corresponds to the actual token, if not, probably invalid or expired
	if token != req.FormValue("_token") && time.Now().UTC().Add(15*time.Minute).After(execTime) { //nolint:gomnd
		conf.frontErrorPage(writer, req, http.StatusBadRequest, "Token invalid or expired.")

		return
	}

	// Get the hash corresponding to the short
	hash, err := database.GetHashByShort(conf.db, req.FormValue("short"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(
			writer,
			req,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
		)

		return
	}

	// Get the password from the form, throw an error page if the form doesn't have a value
	var password string
	if req.FormValue("access") == "Access" {
		password = req.FormValue("password")
	} else {
		conf.frontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the password.")

		return
	}

	// Check if the password matches the hash
	if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
		!match {
		log.Println(err)
		conf.frontErrorPage(writer, req, http.StatusBadRequest, "The password is incorrect.")

		return
	} else if err != nil {
		log.Println(err)
		conf.frontErrorPage(writer, req, http.StatusInternalServerError, "Unable to compare the password against the hash.")

		return
	}

	// Get the URL corresponding to the short
	url, err := database.GetURLByShort(conf.db, req.FormValue("short"))
	if err != nil {
		log.Println(err)
		conf.frontErrorPage(
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
