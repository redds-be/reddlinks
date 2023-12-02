package main

import (
	"encoding/json"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
	"log"
	"net/http"
	"time"
)

type parameters struct {
	// Define the structure of the JSON payload that will be read from the user
	Url         string `json:"url"`
	Length      int    `json:"length"`
	CustomPath  string `json:"custom_path"`
	ExpireAfter int    `json:"expire_after"`
	Password    string `json:"password"`
}

func (db *Database) handlerRoot(w http.ResponseWriter, r *http.Request) {
	// Check method and decide whether to create or redirect to link
	if r.Method == http.MethodGet {
		db.redirectToUrl(w, r)
	} else if r.Method == http.MethodPost {
		db.createLink(w, r)
	} else {
		respondWithError(w, r, 405, "405 Method Not Allowed")
		return
	}
}

func (db *Database) createLink(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
	var expireAt time.Time

	// Decode the JSON, client error if it can't, most likely an invalid syntax
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 400, "400 JSON syntax is probably invalid.")
		return
	}

	// Check the expiration time and set it to x minute specified by the user, -1 = never, will default to 48 hours
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
	if params.CustomPath == "" {
		params.CustomPath = uniuri.NewLen(params.Length)
	}
	if len(params.CustomPath) > 255 {
		params.CustomPath = params.CustomPath[:255]
	}

	// Check if the path is a reserved one, 'status' and 'error' are used to debug.
	if params.CustomPath == "status" || params.CustomPath == "error" {
		respondWithJSON(w, 400, fmt.Sprintf("The path '/%s' is reserved.", params.CustomPath))
		return
	}

	// If the password given to by the request isn't null (meaning no password), generate an argon2 hash from it
	hash := ""
	if params.Password != "" {
		hash, err = argon2id.CreateHash(params.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			respondWithError(w, r, 500, "500 Could not hash the password.")
			return
		}
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	link, err := database.CreateLink(db.db, uuid.New(), time.Now().UTC(), expireAt, params.Url, params.CustomPath, hash)
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 400, "400 Could not add link: the path is probably already in use.")
		return
	}

	// Send back the expiry time, the url and the short to the user
	respondWithJSON(w, 201, link)
}

func (db *Database) redirectToUrl(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)

	// Check if there is a hash associated with the short, if there is a hash, we will require a password
	hash, err := database.GetHashByShort(db.db, trimFirstRune(r.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 404, "404 there is no link associated with this path, it is probably invalid or expired.")
		return
	}
	if hash != "" {
		// Decode the JSON, client error if it can't, most likely an invalid syntax or no password given at all
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Println(err)
			respondWithError(w, r, 400, "400 Wrong JSON syntax or no password has been given. A password is required to access this link.")
			return
		}

		// Check if the password matches the hash
		if match, err := argon2id.ComparePasswordAndHash(params.Password, hash); err == nil && !match {
			respondWithError(w, r, 400, "400 Wrong password has been given.")
			return
		} else if err != nil {
			log.Println(err)
			respondWithError(w, r, 500, "500 Could not compare the password against corresponding hash.")
			return
		}
	}

	// Get the URL
	url, err := database.GetUrlByShort(db.db, trimFirstRune(r.URL.Path))
	if err != nil {
		log.Println(err)
		respondWithError(w, r, 404, "404 there is no link associated with this path, it is probably invalid or expired.")
		return
	}
	// Redirect the client to the URL associated with the short of the database
	http.Redirect(w, r, url, http.StatusSeeOther)
}
