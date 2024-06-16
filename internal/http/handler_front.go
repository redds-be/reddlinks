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
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/json"
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
	DefaultExpiryTime      int
	DefaultExpiryDate      string
	ContactEmail           string
}

// RenderTemplate renders the templates using a given Page struct.
//
// It starts by setting the appropriate headers using [http.Header.Set] and [http.WriteHeader], then
// the requested template is rendered using a given page struct using [template.ExecuteTemplate].
func RenderTemplate(writer http.ResponseWriter, tmpl string, page *Page, code int) {
	// Tell that we serve HTML in UTF-8.
	writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	// Tell that all resources comes from here and that only this site can frame itself
	writer.Header().
		Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline';"+
			" style-src 'self'; img-src 'self'; connect-src 'self'; frame-src 'self'; font-src 'self'; media-src 'self';"+
			" object-src 'self'; manifest-src 'self'; worker-src 'self'; form-action 'self'; frame-ancestors 'self'")
	// Block access to styles and scripts
	writer.Header().Set("X-Content-Type-Options", "nosniff")

	// Write the header giving a code
	writer.WriteHeader(code)

	// Render a given template, json error if it can't
	err := Templates.ExecuteTemplate(writer, tmpl+".tmpl", page)
	if err != nil {
		json.RespondWithError(writer, http.StatusInternalServerError, "Unable to load the page.")

		return
	}
}

// FrontErrorPage returns an error page to the user using a given code and message with [RenderTemplate].
func (conf Configuration) FrontErrorPage(
	writer http.ResponseWriter,
	_ *http.Request,
	code int,
	errMsg string,
) {
	// Set what is going to be displayed on the error page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Error:         fmt.Sprintf("Error %d: %s", code, errMsg),
		Version:       conf.Version,
	}

	// Display the error page
	RenderTemplate(writer, "error", page, code)
}

// FrontHandlerMainPage displays the main page with a form used to shorten a link.
//
// An expiry date is created by adding DefaultExpiryTime to now, this date will be used as the default expiry date in the form.
func (conf Configuration) FrontHandlerMainPage(writer http.ResponseWriter, _ *http.Request) {
	// Convert default expiry time into date
	defaultExpiryDate := time.Now().UTC().Add(time.Minute * time.Duration(conf.DefaultExpiryTime))

	// Set what is going to be displayed on the main page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(conf.InstanceURL, ""),
		DefaultShortLength:     conf.DefaultShortLength,
		DefaultMaxShortLength:  conf.DefaultMaxShortLength,
		DefaultMaxCustomLength: conf.DefaultMaxCustomLength,
		DefaultExpiryTime:      conf.DefaultExpiryTime,
		DefaultExpiryDate:      defaultExpiryDate.Format("2006-01-02T15:04"),
		Version:                conf.Version,
	}

	// Display the front page
	RenderTemplate(writer, "index", page, http.StatusOK)
}

// FrontHandlerPrivacyPage displays the Privacy Policy page.
func (conf Configuration) FrontHandlerPrivacyPage(writer http.ResponseWriter, _ *http.Request) {
	// Set what is going to be displayed on the privacy page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Version:       conf.Version,
		ContactEmail:  conf.ContactEmail,
	}

	// Display the front page
	RenderTemplate(writer, "privacy", page, http.StatusOK)
}

// FrontCreateLink creates a link entry into the database using the values of the form from the front page.
//
// It firsts checks using a regexp if the URL from the payload has http/https as its protocol, it then checks the expiration time,
// if there is none, DefaultExpiryTime will be added to now, if there's one, the time will be parsed using
// [atp.ParseDuration] and this time will be added to now. the length provided will be checked and fixed
// according to min and max settings. the custom path provided will be checked if there's one,
// endpoints and some characters are blacklisted, if the path exceeds the length of DefaultMaxCustomLength,
// it will be trimmed. If there's no custom path provided, a random one will generated using either DefaultShortLength or
// the provided length with [utils.GenStr]. If there's a password provided, it will be hashed using [argon2id.CreateHash].
// After all is done, a link entry will be created in the database using [database.CreateLink].
// If there's an error when creating a link entry using a generated short, it will be re-generated again and again until it works.
func (conf Configuration) FrontCreateLink( //nolint:cyclop,funlen,gocognit
	params utils.Parameters,
) (string, int, string, database.Link) {
	// Check if the url is valid
	isValid, err := regexp.MatchString(`^https?://.*\..*$`, params.URL)
	if err != nil {
		return "Unable to check the given URL", http.StatusInternalServerError, "", database.Link{}
	}

	if !isValid {
		return fmt.Sprintf(
			"'%s' is not a valid url. (only http and https are supported)",
			params.URL,
		), http.StatusBadRequest, "", database.Link{}
	}

	// Convert the datetime into a date
	var expireAt time.Time
	if params.ExpireDate == "" {
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.DefaultExpiryTime))
	} else {
		expireAt, err = time.Parse("2006-01-02T15:04", params.ExpireDate)
		if err != nil {
			return "Unable to parse the expiry date", http.StatusInternalServerError, "", database.Link{}
		}
	}

	// Check the length, will default to 6 if it's inferior or equal to 0 or will default to 16 if it's over 16
	if params.Length <= 0 {
		params.Length = conf.DefaultShortLength
	} else if params.Length > conf.DefaultMaxShortLength {
		params.Length = conf.DefaultMaxShortLength
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
			return "Could not hash the password.", http.StatusInternalServerError, "", database.Link{}
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
		return "Could not add link: the path is probably already in use.", http.StatusBadRequest, "", database.Link{}
	} else if err != nil && autoGen {
	loop:
		for index := conf.DefaultShortLength; index <= conf.DefaultMaxShortLength; index++ {
			params.Path = utils.GenStr(index, allowedChars)
			err = database.CreateLink(conf.DB, uuid.New(), time.Now().UTC(), expireAt, params.URL, params.Path, hash)
			switch {
			case err != nil && index == conf.DefaultMaxShortLength:
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

	return "", http.StatusCreated, addInfo, link
}

// FrontHandlerAdd displays the information about the newly added link to the user.
//
// It starts by gathering the form values given by the front page into a [utils.Parameters] struct
// and uses that to call [FrontCreateLink], after the link is created, the useful informations will be
// displayed to the user.
func (conf Configuration) FrontHandlerAdd(
	writer http.ResponseWriter,
	req *http.Request,
) {
	// What to if the form is correct, i.e. the front page form was posted.
	// If this isn't the case, display an error page
	if req.FormValue("add") != "Add" {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the form.")

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
		)

		return
	}

	// Set the values that will be used for the link creation
	params := utils.Parameters{
		URL:        req.FormValue("url"),
		Length:     length,
		Path:       req.FormValue("short"),
		ExpireDate: req.FormValue("expire_datetime"),
		Password:   req.FormValue("password"),
	}

	// Create a link entry into the database, display an error page if it can't
	errMsg, code, addInfo, link := conf.FrontCreateLink(params)
	if errMsg != "" {
		conf.FrontErrorPage(writer, req, code, errMsg)

		return
	}

	// Format the expiration date that will be displayed to the user
	expireAt := link.ExpireAt.Format(time.ANSIC)

	// Set what is going to be displayed on the add page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").
			ReplaceAllString(fmt.Sprintf("%s%s", conf.InstanceURL, link.Short), ""),
		Short:    link.Short,
		URL:      link.URL,
		ExpireAt: expireAt,
		Password: params.Password,
		AddInfo:  addInfo,
		Version:  conf.Version,
	}

	// Display the add page which will display the information about the added link
	RenderTemplate(writer, "add", page, http.StatusCreated)
}

// FrontAskForPassword asks for a password to access a given shortened link.
func (conf Configuration) FrontAskForPassword(writer http.ResponseWriter, req *http.Request) {
	// Set what is going to be displayed on the pass page
	page := &Page{
		InstanceTitle: conf.InstanceName,
		InstanceURL:   conf.InstanceURL,
		Short:         req.PathValue("short"),
		Version:       conf.Version,
	}

	// Display the pass page which will ask the user for a password
	RenderTemplate(writer, "pass", page, http.StatusOK)
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
		)

		return
	}

	// Get the password from the form, throw an error page if the form doesn't have a value
	var password string
	if req.FormValue("access") == "Access" {
		password = req.FormValue("password")
	} else {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to read the password.")

		return
	}

	// Check if the password matches the hash
	if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil &&
		!match {
		conf.FrontErrorPage(writer, req, http.StatusBadRequest, "The password is incorrect.")

		return
	} else if err != nil {
		conf.FrontErrorPage(writer, req, http.StatusInternalServerError, "Unable to compare the password against the hash.")

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
		)

		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(writer, req, url, http.StatusSeeOther)
}
