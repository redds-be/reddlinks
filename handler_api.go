package main

import (
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
	"log"
	"net/http"
	"regexp"
	"time"
)

func (info sendToHandlers) apiRedirectToUrl(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(info.db, trimFirstRune(r.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 404, "There is no link associated with this path, it is probably invalid or expired.")
		return
	}

	if hash != "" {
		// Decode the JSON, client error if it can't, most likely an invalid syntax or no password given at all
		isJson := false
		for _, contentType := range r.Header["Content-Type"] {
			if contentType == "application/json" {
				isJson = true
				break
			}
		}

		// If the client is sending json, decode it and set it as the password,
		// Else if the client is sending a parameter, use its value as the password,
		// Else if the client gives nothing, probably a browser, let a front handler handle that.
		password := ""
		if isJson {
			params, err := decodeJSON(r)
			if err != nil {
				fmt.Println(err)
				respondWithError(w, r, 400, "Wrong JSON or no password has been given. This link requires a password to access it.")
				return
			}
			password = params.Password
		} else if r.URL.Query().Get("pass") != "" {
			password = r.URL.Query().Get("pass")
		} else {
			info.frontAskForPassword(w, r)
			return
		}

		// Check if the password matches the hash
		if match, err := argon2id.ComparePasswordAndHash(password, hash); err == nil && !match {
			respondWithError(w, r, 400, "Wrong password has been given.")
			return
		} else if err != nil {
			log.Println(err)
			respondWithError(w, r, 500, "Could not compare the password against corresponding hash.")
			return
		}
	}

	// Get the URL
	url, err := database.GetUrlByShort(info.db, trimFirstRune(r.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 404, "There is no link associated with this path, it is probably invalid or expired.")
		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (info sendToHandlers) apiCreateLink(w http.ResponseWriter, params parameters) (database.Link, int, string) {
	// Check if the url is valid
	isValid, err := regexp.MatchString(`^https?://.*\..*$`, params.Url)
	if err != nil {
		fmt.Println(err)
		return database.Link{}, 500, "Unable to check the URL."
	}
	if !isValid {
		return database.Link{}, 400, "The URL is invalid."
	}

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
	autoGen := false
	if params.Path == "" {
		autoGen = true
		params.Path = uniuri.NewLen(params.Length)
	}
	if len(params.Path) > 255 {
		params.Path = params.Path[:255]
	}

	// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access and assets are used for the front.
	reservedMatch, err := regexp.MatchString(`^status$|^error$|^add$|^access$|^assets.*$`, params.Path)
	if err != nil {
		fmt.Println(err)
		return database.Link{}, 500, "Could not check the path."
	}
	if reservedMatch {
		return database.Link{}, 400, fmt.Sprintf("The path '/%s' is reserved.", params.Path)
	}

	// If the password given to by the request isn't null (meaning no password), generate an argon2 hash from it
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			return database.Link{}, 500, "Could not hash the password."
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	link, err := database.CreateLink(info.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.Path, hash)
	if err != nil && !autoGen {
		log.Println(err)
		return database.Link{}, 400, "Could not add link: the path is probably already in use."
	} else if err != nil && autoGen {
		for i := 6; i <= 16; i++ {
			params.Path = uniuri.NewLen(i)
			link, err = database.CreateLink(info.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.Path, hash)
			if err != nil {
				log.Println(err)
			} else if err != nil && i == 16 {
				return database.Link{}, 500, "No more space left in the database."
			} else if err == nil && i != params.Length {
				type informationResponse struct {
					Information string `json:"information"`
				}
				respondWithJSON(w, 100, informationResponse{Information: "The length of your auto-generated path had to be changed due to space limitations in the database."})
				break
			} else if err == nil {
				break
			}
		}
	}

	// Return the expiry time, the url and the short to the user
	return link, 0, ""
}

func (info sendToHandlers) apiHandlerRoot(w http.ResponseWriter, r *http.Request) {
	// Check method and decide whether to create or redirect to link
	if r.Method == http.MethodGet && r.URL.Path == "/favicon.ico" {
		return
	} else if r.Method == http.MethodGet && r.URL.Path == "/" {
		info.frontHandlerMainPage(w, r)
	} else if r.Method == http.MethodGet {
		info.apiRedirectToUrl(w, r)
	} else if r.Method == http.MethodPost {
		log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
		params, err := decodeJSON(r)
		if err != nil {
			respondWithError(w, r, 400, "Invalid JSON syntax.")
			return
		}
		link, code, errMsg := info.apiCreateLink(w, params)
		if errMsg != "" {
			respondWithError(w, r, code, errMsg)
			return
		}
		// Send back the expiry time, the url and the short to the user
		respondWithJSON(w, 201, link)
	} else {
		respondWithError(w, r, 405, "Method Not Allowed.")
		return
	}
}
