package main

import (
	"github.com/redds-be/rlinks/database"
	"log"
	"time"
	"unicode/utf8"
)

func trimFirstRune(s string) string {
	// Remove the first letter of a string (https://go.dev/play/p/ZOZyRORkK82)
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func (db *Database) collectGarbage() {
	// Just some kind of hack to call the manual garbage collecting function every minute
	for {
		log.Println("Collecting garbage...")
		// Get the links
		links, err := database.GetLinks(db.db)
		if err != nil {
			log.Println(err)
			return
		}

		// Go through the link and delete expired ones
		now := time.Now().UTC()
		for _, link := range links {
			if now.After(link.ExpireAt) || now.Equal(link.ExpireAt) {
				log.Printf("URL : %s (%s) is expired, deleting it...", link.Url, link.Short)
				err := database.RemoveLink(db.db, link.Short)
				if err != nil {
					log.Printf("Could not remove URL : %s (%s): %s", link.Url, link.Short, err)
					return
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}
