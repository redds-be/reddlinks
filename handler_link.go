package main

import (
	"encoding/json"
	"fmt"
	"github.com/chmike/domain"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"github.com/redds-be/rlinks/internal/database"
	"log"
	"net/http"
	"time"
)

func (apiCfg apiConfig) handlerCreateLink(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
	var expireAt time.Time

	// Define the structure of the JSON payload that will be read from the user
	type parameters struct {
		Url         string `json:"url"`
		Length      int    `json:"length"`
		CustomPath  string `json:"custom_path"`
		ExpireAfter int    `json:"expire_after"`
	}

	// Decode the JSON, client error if it can't, most likely an invalid syntax
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, r, 400, "Error parsing JSON : Syntax is probably invalid.")
		return
	}

	// Check the url before continuing
	err = domain.Check(params.Url)
	if err != nil {
		respondWithError(w, r, 400, fmt.Sprintf("Error reading the url : %s", err))
		return
	}

	// Check the expiration time and set it to x minute specified by the user, -1 = never, will default to 48 hours
	if params.ExpireAfter == -1 {
		expireAt = time.Date(9999, 12, 30, 23, 59, 59, 59, time.UTC)
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
	} else if len(params.CustomPath) > 16 {
		params.CustomPath = params.CustomPath[:16]
	}

	// Check if the path is a reserved one, 'status' and 'error' are used to debug, 'garbage' is used to delete expired links
	if params.CustomPath == "status" || params.CustomPath == "error" || params.CustomPath == "garbage" {
		respondWithJSON(w, 400, fmt.Sprintf("The path '/%s' is reserved.", params.CustomPath))
		return
	}

	// Insert the information to the database, error if it can't, most likely that the short is already in use
	link, err := apiCfg.DB.CreateLink(r.Context(), database.CreateLinkParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		ExpireAt:  expireAt,
		Url:       params.Url,
		Short:     params.CustomPath,
	})
	if err != nil {
		respondWithError(w, r, 400, "Could not add link: The path is probably already in use.")
		return
	}

	// Send back the expiry time, the url and the short to the user
	respondWithJSON(w, 201, databaseLinkToLink(link))
}

func (apiCfg apiConfig) handlerGetLink(w http.ResponseWriter, r *http.Request, link database.Link) {
	log.Printf("Client : %s (%s) accessing '%s' with method '%s'.\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method)
	// Redirect the client to the URL associated with the short of the database
	handlerRedirect(w, r, databaseLinkToLink(link).Url)
}
