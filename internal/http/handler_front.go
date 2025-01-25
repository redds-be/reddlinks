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
	"html/template"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/json"
	"github.com/redds-be/reddlinks/internal/links"
	"github.com/redds-be/reddlinks/internal/utils"
)

// Templates is a global variables for the HTML templates.
var Templates *template.Template //nolint:gochecknoglobals

// Page defines the structure of what can be displayed on a page.
//
// InstanceTitle is the title of the instance,
// InstanceURL is the URL of the instance,
// ShortenedLink is the shortened URL,
// Short is the short path linked to the URL,
// URL is the URL to be shortened,
// ExpireAt is the formatted date of expiration of a link,
// Password is the password used by the user to create a link,
// Error is an error message,
// AddInfo is an information to be displayed to the user after link creation,
// Version is the version of reddlinks used by this instance,
// DefaultShortLength refers to the default length of generated strings for a short URL,
// DefaultMaxShortLength refers to the maximum length of generated strings for a short URL,
// DefaultMaxCustomLength refers to the maximum length of custom strings for a short URL,
// DefaultExpiryTime refers to the default expiry time of links records,
// DefaultExpiryDate refers to the default expiry date,
// ContactEmail refers to an optional admin's contact email.
type Page struct {
	InstanceTitle          string
	InstanceURL            string
	ShortenedLink          string
	ShortenedQR            string
	Short                  string
	URL                    string
	ExpireAt               string
	Password               string
	Error                  string
	AddInfo                string
	Version                string
	DefaultShortLength     int
	DefaultMaxShortLength  int
	DefaultMaxCustomLength int
	DefaultExpiryDate      string
	ContactEmail           string
}

// RenderTemplate renders the templates using a given Page struct.
//
// It starts by setting the appropriate headers using [http.Header.Set] and [http.WriteHeader], then
// the requested template is rendered using a given page struct using [template.ExecuteTemplate].
func RenderTemplate(writer http.ResponseWriter, tmpl string, page *Page, code int, lang string) {
	// Tell that we serve HTML in UTF-8.
	writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	// Tell that all resources comes from here and that only this site can frame itself
	writer.Header().
		Set("Content-Security-Policy", "default-src 'none'; script-src 'self';"+
			"style-src 'self'; img-src 'self' data: ;")
	// Block access to styles and scripts
	writer.Header().Set("X-Content-Type-Options", "nosniff")

	// Write the header giving a code
	writer.WriteHeader(code)

	// Check if lang is supported, defaults to english if it's not the case
	supportedLangs := []string{"en", "fr"}

	if !slices.Contains(supportedLangs, lang) {
		lang = "en"
	}

	// Render a given template, json error if it can't
	err := Templates.ExecuteTemplate(writer, tmpl+"."+lang+".tmpl", page)
	if err != nil {
		json.RespondWithError(writer, http.StatusInternalServerError, "Unable to load the page.")

		return
	}
}

// FrontErrorPage returns an error page to the user using a given code and message with [RenderTemplate].
func (conf Configuration) FrontErrorPage(
	writer http.ResponseWriter,
	req *http.Request,
	code int,
	errMsg string,
	url string,
) {
	// Get the client's main language
	lang := req.Header.Get("Accept-Language")
	if len(lang) >= 2 { //nolint:mnd
		lang = lang[:2]
	} else {
		lang = "en"
	}

	// Set what is going to be displayed on the error page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Error:         fmt.Sprintf("Error %d: %s", code, errMsg),
		Version:       conf.Version,
		URL:           url,
	}

	// Display the error page
	RenderTemplate(writer, "error", page, code, lang)
}

// FrontHandlerMainPage displays the main page with a form used to shorten a link.
//
// An expiry date is created by adding DefaultExpiryTime to now, this date will be used as the default expiry date in the form.
func (conf Configuration) FrontHandlerMainPage(writer http.ResponseWriter, req *http.Request) {
	// Get the client's main language
	lang := req.Header.Get("Accept-Language")
	if len(lang) >= 2 { //nolint:mnd
		lang = lang[:2]
	} else {
		lang = "en"
	}

	var defaultExpiryDate string
	if conf.DefaultExpiryTime != 0 {
		// Convert default expiry time into date
		defaultExpiryDate = time.Now().
			UTC().
			Add(time.Minute * time.Duration(conf.DefaultExpiryTime)).
			Format("2006-01-02T15:04")
	}

	// Set what is going to be displayed on the main page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(conf.InstanceURL, ""),
		DefaultShortLength:     conf.DefaultShortLength,
		DefaultMaxShortLength:  conf.DefaultMaxShortLength,
		DefaultMaxCustomLength: conf.DefaultMaxCustomLength,
		DefaultExpiryDate:      defaultExpiryDate,
		Version:                conf.Version,
	}

	// Display the front page
	RenderTemplate(writer, "index", page, http.StatusOK, lang)
}

// FrontHandlerPrivacyPage displays the Privacy Policy page.
func (conf Configuration) FrontHandlerPrivacyPage(writer http.ResponseWriter, req *http.Request) {
	// Get the client's main language
	lang := req.Header.Get("Accept-Language")
	if len(lang) >= 2 { //nolint:mnd
		lang = lang[:2]
	} else {
		lang = "en"
	}

	// Set what is going to be displayed on the privacy page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Version:       conf.Version,
		ContactEmail:  conf.ContactEmail,
	}

	// Display the front page
	RenderTemplate(writer, "privacy", page, http.StatusOK, lang)
}

// FrontHandlerAdd displays the information about the newly added link to the user.
//
// It starts by gathering the form values given by the front page into a [utils.Parameters] struct,
// it then creates an adapter for links using [links.NewAdapter] then calls [links.CreateLink]
// to create a link entry giving it the deocded params, the information is then formatted and returned to the client
// on a web page.
func (conf Configuration) FrontHandlerAdd( //nolint:funlen
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the client's main language
	lang := req.Header.Get("Accept-Language")
	if len(lang) >= 2 { //nolint:mnd
		lang = lang[:2]
	} else {
		lang = "en"
	}

	// What to if the form is correct, i.e. the front page form was posted.
	// If this isn't the case, display an error page
	if req.FormValue("add") != "Add" {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the form.", req.URL.Path)

		return
	}

	// Convert the length to an int, display an error page if it can't
	length, err := strconv.Atoi(req.FormValue("length"))
	if err != nil {
		conf.FrontErrorPage(
			writer,
			req,
			http.StatusInternalServerError,
			"There was an error trying to read the length.",
			req.URL.Path,
		)

		return
	}

	// Set the values that will be used for the link creation
	params := utils.Parameters{
		URL:         req.FormValue("url"),
		Length:      length,
		Path:        req.FormValue("short"),
		ExpireDate:  req.FormValue("expire_datetime"),
		ExpireAfter: req.FormValue("expire_after"),
		Password:    req.FormValue("password"),
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

	// Create a link entry into the database, display an error page if it can't
	link, code, addInfo, errMsg := linksAdapter.CreateLink(params)
	if errMsg != "" {
		conf.FrontErrorPage(writer, req, code, errMsg, req.URL.Path)

		return
	}

	// Format the expiration date that will be displayed to the user
	var expireAt string
	if params.ExpireDate == "" && params.ExpireAfter == "" && conf.DefaultExpiryTime == 0 {
		expireAt = "Never"
	} else {
		expireAt = link.ExpireAt.Format(time.RFC822)
	}

	qr, err := utils.TextToB64QR(conf.InstanceURL + link.Short) //nolint:varnamelen // The name is self-explanatory
	if err != nil {
		conf.FrontErrorPage(writer, req, code, errMsg, "/")

		return
	}

	// Set what is going to be displayed on the add page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(fmt.Sprintf("%s%s", conf.InstanceURL, link.Short), ""),
		Short:       link.Short,
		URL:         link.URL,
		ExpireAt:    expireAt,
		Password:    params.Password,
		AddInfo:     addInfo,
		Version:     conf.Version,
		ShortenedQR: qr,
	}

	// Display the add page which will display the information about the added link
	RenderTemplate(writer, "add", page, http.StatusCreated, lang)
}

// FrontAskForPassword asks for a password to access a given shortened link.
func (conf Configuration) FrontAskForPassword(writer http.ResponseWriter, req *http.Request) {
	// Get the client's main language
	lang := req.Header.Get("Accept-Language")
	if len(lang) >= 2 { //nolint:mnd
		lang = lang[:2]
	} else {
		lang = "en"
	}

	// Set what is going to be displayed on the pass page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Short:         req.PathValue("short"),
		Version:       conf.Version,
	}

	// Display the pass page which will ask the user for a password
	RenderTemplate(writer, "pass", page, http.StatusOK, lang)
}

// FrontHandlerRedirectToURL redirects the client to the URL corresponding to given shortened link.
//
// It starts by getting the hash of the short using [database.GetHashByShort],
// then it gets the password from [FrontAskForPassword],
// it then compares the hash of the given password with the short's hash using [argon2id.ComparePasswordAndHash],
// if the password matches, it uses [database.GetURLByShort] to get the URL to redirect to before redirect to said URL.
func (conf Configuration) FrontHandlerRedirectToURL(
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Get the hash corresponding to the short
	hash, err := database.GetHashByShort(conf.DB, req.FormValue("short"))
	if err != nil {
		conf.FrontErrorPage(
			writer,
			req,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
			"/",
		)

		return
	}

	returnURL := req.FormValue("short")

	// Get the password from the form, throw an error page if the form doesn't have a value
	var password string
	if req.FormValue("access") == "Access" {
		password = req.FormValue("password")
	} else {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the password.", returnURL)

		return
	}

	// Check if the password matches the hash
	if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
		!match {
		conf.FrontErrorPage(writer, req, http.StatusBadRequest, "The password is incorrect.", returnURL)

		return
	} else if err != nil {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to compare the password against the hash.", returnURL)

		return
	}

	// Get the URL corresponding to the short
	url, err := database.GetURLByShort(conf.DB, req.FormValue("short"))
	if err != nil {
		conf.FrontErrorPage(
			writer,
			req,
			http.StatusNotFound,
			"There is no link associated with this path, it is probably invalid or expired.",
			req.URL.Path,
		)

		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(writer, req, url, http.StatusSeeOther)
}
