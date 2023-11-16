package main

import (
	"github.com/redds-be/rlinks/internal/database"
	"time"
)

type Link struct {
	// Define the structure of a link entry that will be served to the client in json
	ExpireAt time.Time `json:"expire_at"`
	Url      string    `json:"url"`
	Short    string    `json:"short"`
}

func databaseLinkToLink(dbLink database.Link) Link {
	// Transform the link entry of the database into a simpler form to serve to the client
	return Link{
		ExpireAt: dbLink.ExpireAt,
		Url:      dbLink.Url,
		Short:    dbLink.Short,
	}
}
