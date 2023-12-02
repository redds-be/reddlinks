package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func DbConnect(dbURL string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connect to the database, please check the database URL:", err)
	}

	return db
}
