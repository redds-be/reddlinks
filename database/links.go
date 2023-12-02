package database

import (
	"database/sql"
	"github.com/google/uuid"
	"log"
	"time"
)

func checkTable(db *sql.DB, table string) bool {
	sqlCheckTable := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1);`
	var tableExists bool
	err := db.QueryRow(sqlCheckTable, table).Scan(&tableExists)
	if err != nil {
		log.Fatal("Unable to check if the table 'links' exists:", err)
	}

	return tableExists
}

func CreateLinksTable(db *sql.DB) {
	doTableExists := checkTable(db, "links")
	if !doTableExists {
		sqlCreateTable := `CREATE TABLE links (id UUID PRIMARY KEY, created_at TIMESTAMP NOT NULL, expire_at TIMESTAMP NOT NULL, url varchar NOT NULL, short varchar(255) UNIQUE NOT NULL, password varchar(97));`
		_, err := db.Exec(sqlCreateTable)
		if err != nil {
			log.Fatal("Unable to create the 'links' table:", err)
		}
	}
}

func CreateLink(db *sql.DB, id uuid.UUID, createdAt time.Time, expireAt time.Time, url, short, password string) (Link, error) {
	sqlCreateLink := `INSERT INTO links (id, created_at, expire_at, url, short, password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING expire_at, url, short;`
	var returnValues Link
	err := db.QueryRow(sqlCreateLink, id, createdAt, expireAt, url, short, password).Scan(
		&returnValues.ExpireAt,
		&returnValues.Url,
		&returnValues.Short,
	)

	return returnValues, err
}

func GetUrlByShort(db *sql.DB, short string) (string, error) {
	sqlGetUrlByShort := `SELECT url FROM links WHERE short = $1;`
	var url string
	err := db.QueryRow(sqlGetUrlByShort, short).Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

func GetHashByShort(db *sql.DB, short string) (string, error) {
	sqlGetPasswordByShort := `SELECT password FROM links WHERE short = $1;`
	var password string
	err := db.QueryRow(sqlGetPasswordByShort, short).Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil
}

func GetLinks(db *sql.DB) ([]Link, error) {
	sqlGetLinks := `SELECT expire_at, url, short FROM links;`
	var links []Link
	rows, err := db.Query(sqlGetLinks)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var i Link
		err := rows.Scan(
			&i.ExpireAt,
			&i.Url,
			&i.Short,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, i)
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}
	return links, nil
}

func RemoveLink(db *sql.DB, short string) error {
	sqlRemoveLink := `DELETE FROM links WHERE short = $1;`
	_, err := db.Exec(sqlRemoveLink, short)
	if err != nil {
		return err
	}

	return nil
}
