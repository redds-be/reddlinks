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
	"strings"
	"time"
)

func (conf configuration) apiRedirectToUrl(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(conf.db, trimFirstRune(r.URL.Path))
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
			conf.frontAskForPassword(w, r)
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
	url, err := database.GetUrlByShort(conf.db, trimFirstRune(r.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 404, "There is no link associated with this path, it is probably invalid or expired.")
		return
	}

	// Redirect the client to the URL associated with the short of the database
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (conf configuration) apiCreateLink(w http.ResponseWriter, params parameters) (database.Link, int, string) {
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
	} else if params.ExpireAfter <= 0 {
		expireAt = time.Now().UTC().Add(time.Minute * time.Duration(conf.defaultExpiryTime))
	} else {
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
		// Check if the path is a reserved one, 'status' and 'error' are used to debug. add, access and assets are used for the front.
		reservedMatch, err := regexp.MatchString(`^status$|^error$|^add$|^access$|^assets.*$`, params.Path)
		if err != nil {
			fmt.Println(err)
			return database.Link{}, 500, "Could not check the path."
		}
		if reservedMatch {
			return database.Link{}, 400, fmt.Sprintf("The path '/%s' is reserved.", params.Path)
		}

		// Check the validity of the custom path
		reservedChars := []string{":", "/", "?", "#", "[", "]", "@", "!", "$", "&", "'", "(", ")", "*", "+", ",", ";", "="}
		for _, char := range reservedChars {
			if match := strings.Contains(params.Path, char); match {
				return database.Link{}, 400, fmt.Sprintf("The character '%s' is not allowed.", char)
			}
		}
	}

	// Check the path, will default to a randomly generated one with specified length, if its length is over 16, it will be trimmed
	autoGen := false
	allowedChars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~")
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
			return database.Link{}, 500, "Could not hash the password."
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	link, err := database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.Path, hash)
	if err != nil && !autoGen {
		log.Println(err)
		return database.Link{}, 400, "Could not add link: the path is probably already in use."
	} else if err != nil && autoGen {
		for i := conf.defaultShortLength; i <= conf.defaultMaxShortLength; i++ {
			params.Path = uniuri.NewLenChars(i, allowedChars)
			link, err = database.CreateLink(conf.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.Path, hash)
			if err != nil {
				log.Println(err)
			} else if err != nil && i == conf.defaultMaxShortLength {
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

func (conf configuration) apiHandlerRoot(w http.ResponseWriter, r *http.Request) {
	// Check method and decide whether to create or redirect to link
	if r.Method == http.MethodGet && r.URL.Path == "/favicon.ico" {
		return
	} else if r.Method == http.MethodGet && r.URL.Path == "/" {
		conf.frontHandlerMainPage(w, r)
	} else if r.Method == http.MethodGet {
		conf.apiRedirectToUrl(w, r)
	} else if r.Method == http.MethodPost {
		log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
		params, err := decodeJSON(r)
		if err != nil {
			respondWithError(w, r, 400, "Invalid JSON syntax.")
			return
		}
		link, code, errMsg := conf.apiCreateLink(w, params)
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
