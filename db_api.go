package main

import (
	"database/sql"
	"github.com/redds-be/rlinks/internal/database"
	"log"
)

type apiConfig struct {
	// Define the apiConfig struct giving it the queries functions
	DB *database.Queries
}

func dbConnect(dbURL string) apiConfig {
	// Connect to the database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connect to the database, please check the database URL.")
	}

	queries := database.New(conn)

	// Create the apiConfig struct giving it the queries functions
	apiCfg := apiConfig{DB: queries}

	return apiCfg
}
