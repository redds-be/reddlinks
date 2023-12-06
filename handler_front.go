package main

import (
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Page struct {
	// Define the structure of what can be displayed on a page
	InstanceTitle string
	InstanceURL   string
	ShortenedLink string
	Short         string
	Url           string
	ExpireAt      string
	Password      string
	Error         string
}

// Parse the templates files in advance
var templates = template.Must(template.ParseFiles(
	"static/index.html",
	"static/add.html",
	"static/error.html",
	"static/pass.html"))

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, p any) {
	// Render a given template, json error if it can't
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 500, "Unable to load the page.")
		return
	}
}

func (info sendToHandlers) frontErrorPage(w http.ResponseWriter, r *http.Request, code int, errMsg string) {
	// Set what is going to be displayed on the error page
	p := &Page{
		InstanceTitle: info.instanceName,
		InstanceURL:   info.instanceURL,
		Error:         fmt.Sprintf("Error %d: %s", code, errMsg),
	}

	// Display the error page
	renderTemplate(w, r, "error", p)
}

func (info sendToHandlers) frontHandlerMainPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
	// Set what is going to be displayed on the main page
	p := &Page{
		InstanceTitle: info.instanceName,
		InstanceURL:   info.instanceURL,
		ShortenedLink: regexp.MustCompile("^https://|http://").ReplaceAllString(info.instanceURL, ""),
	}

	// Display the front page
	renderTemplate(w, r, "index", p)
}

func (info sendToHandlers) frontCreateLink(params parameters) (string, int, database.Link) {
	// Check the expiration time and set it to x minute specified by the user, -1 = never, will default to 48 hours
	var expireAt time.Time
	if params.ExpireAfter == -1 {
		expireAt = time.Date(9999, 12, 31, 23, 59, 59, 59, time.UTC)
		params.Length = 16
	} else if params.ExpireAfter <= 0 {
		expireAt = time.Now().UTC().Add(time.Hour * 24 * 2)
	} else {
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(params.ExpireAfter))
	}

	// Check the length, will default to 6 if it's inferior or equal to 0 or will default to 16 if it's over 16
	if params.Length <= 0 {
		params.Length = 6
	} else if params.Length > 16 {
		params.Length = 16
	}

	// Check the path, will default to a randomly generated one with specified length, if its length is over 16, it will be trimmed
	if params.Path == "" {
		params.Path = uniuri.NewLen(params.Length)
	}
	if len(params.Path) > 255 {
		params.Path = params.Path[:255]
	}

	// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access and assets are used for the front.
	reservedMatch, err := regexp.MatchString(`^status$|^error$|^add$|^access$|^assets.*$`, params.Path)
	if err != nil {
		return "Error 500: The path could not be checked.", 500, database.Link{}
	}
	if reservedMatch {
		return fmt.Sprintf("The path '/%s' is reserved.", params.Path), 400, database.Link{}
	}

	// If the password given to by the request isn't null (meaning no password), generate an argon2 hash from it
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			return "Error 500: Could not hash the password.", 500, database.Link{}
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	link, err := database.CreateLink(info.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.Path, hash)
	if err != nil {
		log.Println(err)
		return "Error 400: Could not add link: the path is probably already in use.", 400, database.Link{}
	}

	return "", 0, link
}

func (info sendToHandlers) frontHandlerAdd(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)

	// What to if the form is correct, i.e. the front page form was posted.
	// If this isn't the case, display an error page
	if r.FormValue("add") == "Add" {
		// Convert the length to an int, display an error page if it can't
		length, err := strconv.Atoi(r.FormValue("length"))
		if err != nil {
			log.Println(err)
			info.frontErrorPage(w, r, 500, "There was an error trying to read the password.")
			return
		}

		// Convert the expiration time to an int, display an error page if it can't
		expireAfter, err := strconv.Atoi(r.FormValue("expire_after"))
		if err != nil {
			log.Println(err)
			info.frontErrorPage(w, r, 500, "There was an error trying to read the expiration time.")
			return
		}

		// Set the values that will be used for the link creation
		params := parameters{
			Url:         r.FormValue("url"),
			Length:      length,
			Path:        r.FormValue("short"),
			ExpireAfter: expireAfter,
			Password:    r.FormValue("password"),
		}

		// Create a link entry into the database, display an error page if it can't
		errMsg, code, link := info.frontCreateLink(params)
		if errMsg != "" {
			info.frontErrorPage(w, r, code, errMsg)
			return
		}

		// Format the expiration date that will be displayed to the user
		expireAt := ""
		if params.ExpireAfter == -1 {
			expireAt = "never"
		} else {
			expireAt = link.ExpireAt.Format(time.ANSIC)
		}

		// Set what is going to be displayed on the add page
		p := &Page{
			InstanceTitle: info.instanceName,
			InstanceURL:   info.instanceURL,
			ShortenedLink: regexp.MustCompile("^https://|http://").ReplaceAllString(fmt.Sprintf("%s%s", info.instanceURL, link.Short), ""),
			Short:         link.Short,
			Url:           link.Url,
			ExpireAt:      expireAt,
			Password:      params.Password,
		}

		// Display the add page which will display the information about the added link
		renderTemplate(w, r, "add", p)
	} else {
		info.frontErrorPage(w, r, 500, "Unable to read the form.")
		return
	}
}

func (info sendToHandlers) frontAskForPassword(w http.ResponseWriter, r *http.Request) {
	// Set what is going to be displayed on the pass page
	p := &Page{
		InstanceTitle: info.instanceName,
		InstanceURL:   info.instanceURL,
		Short:         trimFirstRune(r.URL.Path),
	}

	// Display the pass page which will ask the user for a password
	renderTemplate(w, r, "pass", p)
}

func (info sendToHandlers) frontHandlerRedirectToUrl(w http.ResponseWriter, r *http.Request) {
	// Get the hash corresponding to the short
	hash, err := database.GetHashByShort(info.db, r.FormValue("short"))
	if err != nil {
		log.Println(err)
		info.frontErrorPage(w, r, 404, "There is no link associated with this path, it is probably invalid or expired.")
		return
	}

	// Get the password from the form, throw an error page if the form doesn't have a value
	password := ""
	if r.FormValue("access") == "Access" {
		password = r.FormValue("password")
	} else {
		info.frontErrorPage(w, r, 500, "Unable to read the password.")
		return
	}

	// Check if the password matches the hash
	if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil && !match {
		log.Println(err)
		info.frontErrorPage(w, r, 400, "The password is incorrect.")
		return
	} else if err != nil {
		log.Println(err)
		info.frontErrorPage(w, r, 500, "Unable to compare the password against the hash.")
		return
	}

	// Get the URL corresponding to the short
	url, err := database.GetUrlByShort(info.db, r.FormValue("short"))
	if err != nil {
		log.Println(err)
		info.frontErrorPage(w, r, 404, "There is no link associated with this path, it is probably invalid or expired.")
		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(w, r, url, http.StatusSeeOther)
}
