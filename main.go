package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/redds-be/rlinks/database"
	"log"
	"net/http"
)

type Database struct {
	db *sql.DB
}

func main() {
	// Load the env file
	var envFile = ".env"
	portStr, dbURL := getEnv(envFile)

	// Connect to the database
	db := &Database{db: database.DbConnect(dbURL)}
	database.CreateLinksTable(db.db)

	// Assign a handler to these different paths
	http.HandleFunc("/status", handlerReadiness)
	http.HandleFunc("/error", handlerErr)
	http.HandleFunc("/garbage", db.handlerGarbage)
	http.HandleFunc("/", db.handlerRoot)

	// Periodically clean the database
	go collectGarbage(portStr)

	// Start to listen
	log.Printf("Listening on port : '%s'.", portStr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", portStr), nil))
}
