package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func collectGarbage(portStr string) {
	// Just some kind of hack to call the manual garbage collecting function every minute
	garbageURL := fmt.Sprintf("http://localhost:%s/garbage", portStr)
	for {
		time.Sleep(1 * time.Minute)
		_, err := http.Get(garbageURL)
		if err != nil {
			log.Println("There was an error when trying to collect garbage.")
			continue
		}
	}
}

func (apiCfg apiConfig) handlerGarbage(_ http.ResponseWriter, r *http.Request) {
	// Manual garbage collecting when accessing '/garbage', it will go through all the link entries in the database and check if the current time is after or equal the expiry time
	log.Println("Collecting garbage...")
	links, err := apiCfg.DB.GetLinks(r.Context())
	if err != nil {
		log.Println(err)
		return
	}

	// Go through the link and delete expired ones
	now := time.Now().UTC()
	for _, link := range links {
		if now.After(link.ExpireAt) || now.Equal(link.ExpireAt) {
			log.Printf("URL : %s (%s) is expired, deleting it...", link.Url, link.Short)
			err := apiCfg.DB.RemoveLink(r.Context(), link.Short)
			if err != nil {
				log.Printf("Could not remove URL : %s (%s).", link.Url, link.Short)
				return
			}
		}
	}
}
