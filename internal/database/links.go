//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2024 redd
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <https://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateLinksTable creates the links table in the database.
//
// An SQL statement is prepared using the max short length as the max for the short column,
// the statement is then executed which should create the links table if it doesn't exist.
func CreateLinksTable(database *sql.DB, maxShort int) error {
	sqlCreateTable := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS links ("+
			"id UUID PRIMARY KEY, "+
			"created_at TIMESTAMP NOT NULL, "+
			"expire_at TIMESTAMP NOT NULL, "+
			"url varchar NOT NULL, "+
			"short varchar(%d) UNIQUE NOT NULL, "+
			"password varchar(97));",
		maxShort,
	)
	_, err := database.Exec(sqlCreateTable)

	return err
}

// CreateLink inserts a link entry in the links table.
//
// An SQL statement is prepared, it will insert a record into the links table
// giving a uuid, a creation time, an expiration time, a url, a short and a password.
func CreateLink(
	database *sql.DB,
	identifier uuid.UUID,
	createdAt time.Time,
	expireAt time.Time,
	url, short, password string,
) error {
	sqlCreateLink := `INSERT INTO links (id, created_at, expire_at, url, short, password) 
					  VALUES ($1, $2, $3, $4, $5, $6) RETURNING expire_at, url, short;`
	_, err := database.Exec(sqlCreateLink, identifier, createdAt, expireAt, url, short, password)

	return err
}

// GetURLByShort gets a link entry from the links table by its value of the short column.
func GetURLByShort(db *sql.DB, short string) (string, error) {
	sqlGetURLByShort := `SELECT url FROM links WHERE short = $1;`
	var url string
	err := db.QueryRow(sqlGetURLByShort, short).Scan(&url)

	return url, err
}

// GetHashByShort gets a hash from a link entry by its value of the short column.
func GetHashByShort(db *sql.DB, short string) (string, error) {
	sqlGetPasswordByShort := `SELECT password FROM links WHERE short = $1;`
	var password string
	err := db.QueryRow(sqlGetPasswordByShort, short).Scan(&password)

	return password, err
}

// RemoveExpiredLinks removes an expired link entry from the links table.
func RemoveExpiredLinks(db *sql.DB) error {
	sqlRemoveLink := `DELETE FROM links WHERE expire_at <= CURRENT_TIMESTAMP;`
	_, err := db.Exec(sqlRemoveLink)

	return err
}
