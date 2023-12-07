package main

import (
	"database/sql"
	"encoding/json"
	"github.com/redds-be/rlinks/database"
	"log"
	"net/http"
	"time"
	"unicode/utf8"
)

type configuration struct {
	// Define what is going to be sent to the handlers
	db                     *sql.DB
	instanceName           string
	instanceURL            string
	defaultShortLength     int
	defaultMaxShortLength  int
	defaultMaxCustomLength int
	defaultExpiryTime      int
}

type parameters struct {
	// Define the structure of the JSON payload that will be read from the user
	Url         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"custom_path"`
	ExpireAfter int    `json:"expire_after"`
	Password    string `json:"password"`
}

func trimFirstRune(s string) string {
	// Remove the first letter of a string (https://go.dev/play/p/ZOZyRORkK82)
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func (conf configuration) collectGarbage(timeBetweenCleanups int) {
	// Just some kind of hack to call the manual garbage collecting function every minute
	for {
		log.Println("Collecting garbage...")
		// Get the links
		links, err := database.GetLinks(conf.db)
		if err != nil {
			log.Println(err)
			return
		}

		// Go through the link and delete expired ones
		now := time.Now().UTC()
		for _, link := range links {
			if now.After(link.ExpireAt) || now.Equal(link.ExpireAt) {
				log.Printf("URL : %s (%s) is expired, deleting it...", link.Url, link.Short)
				err := database.RemoveLink(conf.db, link.Short)
				if err != nil {
					log.Printf("Could not remove URL : %s (%s): %s", link.Url, link.Short, err)
					return
				}
			}
		}
		// Wait for length of time in minutes specified in .env
		time.Sleep(time.Duration(timeBetweenCleanups) * time.Minute)
	}
}

func decodeJSON(r *http.Request) (parameters, error) {
	// Decode the JSON from the client's request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return parameters{}, err
	}

	return params, nil
}
