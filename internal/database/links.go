//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2025 redd
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
	"log"
	"time"

	"github.com/google/uuid"
)

// indexLinks creates indexes on the links table to improve query performance.
//
// Parameters:
//   - dbase: A pointer to the SQL database connection
//
// Returns:
//   - error: Any error encountered during index creation
func indexLinks(dbase *sql.DB) error {
	// Index on the short column for quick lookups
	if _, err := dbase.Exec("CREATE INDEX IF NOT EXISTS idx_links_short ON links(short);"); err != nil {
		return fmt.Errorf("failed to create short index: %w", err)
	}

	// Index on expiration time for efficient cleanup queries
	if _, err := dbase.Exec("CREATE INDEX IF NOT EXISTS idx_links_expire ON links(expire_at);"); err != nil {
		return fmt.Errorf("failed to create expiration index: %w", err)
	}

	return nil
}

// CreateLinksTable creates the links table in the database if it doesn't exist.
//
// The function creates a links table with columns for unique identifiers, timestamps,
// URL data, and optional password protection. It dynamically sets the maximum
// length for the short string based on the provided configuration.
//
// Parameters:
//   - database: A pointer to the SQL database connection
//   - dbType: The type of database being used ("postgres" or "sqlite")
//   - maxShort: The maximum allowed length for short strings
//
// Returns:
//   - error: Any error encountered during table creation or update
func CreateLinksTable(dbase *sql.DB, dbType string, maxShort int) error {
	// Creating the table with a parameterized short column length
	sqlCreateTable := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS links ("+
			"id UUID PRIMARY KEY, "+
			"created_at TIMESTAMP NOT NULL, "+
			"expire_at TIMESTAMP NOT NULL, "+
			"url TEXT NOT NULL, "+
			"short varchar(%d) UNIQUE NOT NULL, "+
			"password TEXT);",
		maxShort,
	)

	if _, err := dbase.Exec(sqlCreateTable); err != nil {
		return fmt.Errorf("failed to create links table: %w", err)
	}

	// Update table structure for backward compatibility
	if err := updateLinksTable(dbase, dbType, maxShort); err != nil {
		return fmt.Errorf("failed to update links table: %w", err)
	}

	// Create indexes
	if err := indexLinks(dbase); err != nil {
		return fmt.Errorf("failed to index links table: %w", err)
	}

	return nil
}

// updateLinksTable updates the structure of an existing links table to ensure
// compatibility with the current schema definition.
//
// This function handles database-specific operations for modifying table structure
// without losing data. For PostgreSQL, it alters column types directly. For SQLite,
// which doesn't support ALTER COLUMN, it uses a temporary table to reconstruct the data.
//
// Parameters:
//   - database: A pointer to the SQL database connection
//   - dbType: The type of database being used ("postgres" or "sqlite")
//   - maxShort: The maximum allowed length for short strings
//
// Returns:
//   - error: Any error encountered during the update process
func updateLinksTable(database *sql.DB, dbType string, maxShort int) error { //nolint:cyclop,funlen
	switch dbType {
	case "postgres":
		// PostgreSQL supports direct column type modifications
		sqlUpdateMaxShort := fmt.Sprintf(
			"ALTER TABLE links ALTER COLUMN short TYPE varchar(%d);",
			maxShort,
		)
		if _, err := database.Exec(sqlUpdateMaxShort); err != nil {
			return fmt.Errorf("failed to alter short column type: %w", err)
		}

	case "sqlite":
		// SQLite requires table recreation for schema changes
		// Begin transaction for atomicity
		trans, err := database.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		defer func() {
			// Ensure transaction is properly handled on any return path
			if err != nil {
				err := trans.Rollback()
				if err != nil {
					log.Println("failed to rollback transaction:", err)
				}
			}
		}()

		// Create temporary table with the new schema
		sqlCreateTempTable := fmt.Sprintf(
			"CREATE TABLE tmp_links ("+
				"id UUID PRIMARY KEY, "+
				"created_at TIMESTAMP NOT NULL, "+
				"expire_at TIMESTAMP NOT NULL, "+
				"url TEXT NOT NULL, "+
				"short varchar(%d) UNIQUE NOT NULL, "+
				"password varchar(97));",
			maxShort,
		)

		if _, err = trans.Exec(sqlCreateTempTable); err != nil {
			return fmt.Errorf("failed to create temporary table: %w", err)
		}

		// Copy data from old table to new table
		const sqlCopyOldToNew = `
			INSERT INTO tmp_links 
				(id, created_at, expire_at, url, short, password) 
			SELECT 
				id, created_at, expire_at, url, short, password 
			FROM links;`

		if _, err = trans.Exec(sqlCopyOldToNew); err != nil {
			return fmt.Errorf("failed to copy data to temporary table: %w", err)
		}

		// Drop old table and rename new table
		if _, err = trans.Exec("DROP TABLE links;"); err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}

		if _, err = trans.Exec("ALTER TABLE tmp_links RENAME TO links;"); err != nil {
			return fmt.Errorf("failed to rename temporary table: %w", err)
		}

		// Commit the transaction
		if err = trans.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Create indexes
		if err = indexLinks(database); err != nil {
			return fmt.Errorf("failed to index links table: %w", err)
		}
	}

	return nil
}

// CreateLink inserts a new shortened URL entry into the database.
//
// This function stores a complete link record with all necessary metadata including
// creation and expiration timestamps, the original URL, short string, and
// an optional password for protected links.
//
// Parameters:
//   - database: A pointer to the SQL database connection
//   - identifier: A UUID that uniquely identifies this link
//   - createdAt: Timestamp indicating when the link was created
//   - expireAt: Timestamp indicating when the link will expire
//   - url: The original URL that is being shortened
//   - short: The short string to be used in the shortened URL
//   - password: Optional password hash for protected links (empty string if none)
//
// Returns:
//   - error: Any error encountered during the insert operation
func CreateLink(
	database *sql.DB,
	identifier uuid.UUID,
	createdAt time.Time,
	expireAt time.Time,
	url, short, password string,
) error {
	const sqlCreateLink = `
		INSERT INTO links (id, created_at, expire_at, url, short, password) 
		VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := database.Exec(sqlCreateLink, identifier, createdAt, expireAt, url, short, password)
	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	return nil
}

// GetURLInfo retrieves the complete information for a link by its short.
//
// Parameters:
//   - dbase: A pointer to the SQL database connection
//   - short: The shortened URL to look up
//
// Returns:
//   - string: The original URL
//   - time.Time: When the link was created
//   - time.Time: When the link will expire
//   - error: Any error encountered during lookup, including "not found" errors
func GetURLInfo(dbase *sql.DB, short string) (string, time.Time, time.Time, error) {
	const sqlGetURLByShort = `
		SELECT url, created_at, expire_at 
		FROM links 
		WHERE short = $1;`

	var (
		url       string
		createdAt time.Time
		expireAt  time.Time
	)

	err := dbase.QueryRow(sqlGetURLByShort, short).Scan(&url, &createdAt, &expireAt)
	if err != nil {
		return "", time.Time{}, time.Time{}, fmt.Errorf("failed to get URL info: %w", err)
	}

	return url, createdAt, expireAt, nil
}

// GetURLByShort retrieves just the original URL for a given short.
//
// Parameters:
//   - dbase: A pointer to the SQL database connection
//   - short: The shortened URL to look up
//
// Returns:
//   - string: The original URL
//   - error: Any error encountered during lookup, including "not found" errors
func GetURLByShort(dbase *sql.DB, short string) (string, error) {
	const sqlGetURLByShort = `
		SELECT url 
		FROM links 
		WHERE short = $1;`

	var url string
	err := dbase.QueryRow(sqlGetURLByShort, short).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("failed to get URL by short code: %w", err)
	}

	return url, nil
}

// GetHashByShort retrieves the password hash for a given short.
//
// This function is used to check whether a link is password-protected and to
// verify provided passwords against the stored hash.
//
// Parameters:
//   - dbase: A pointer to the SQL database connection
//   - short: The shortened URL to look up
//
// Returns:
//   - string: The stored password hash (empty string if no password)
//   - error: Any error encountered during lookup, including "not found" errors
func GetHashByShort(dbase *sql.DB, short string) (string, error) {
	const sqlGetPasswordByShort = `SELECT password FROM links WHERE short = $1;`

	var password string
	err := dbase.QueryRow(sqlGetPasswordByShort, short).Scan(&password)
	if err != nil {
		return "", fmt.Errorf("failed to get password hash: %w", err)
	}

	return password, nil
}

// RemoveExpiredLinks deletes all links that have passed their expiration date.
//
// Parameters:
//   - dbase: A pointer to the SQL database connection
//
// Returns:
//   - error: Any error encountered during the deletion operation
func RemoveExpiredLinks(dbase *sql.DB) error {
	const sqlRemoveLink = `
		DELETE FROM links 
		WHERE expire_at <= CURRENT_TIMESTAMP;`

	_, err := dbase.Exec(sqlRemoveLink)
	if err != nil {
		return fmt.Errorf("failed to remove expired links: %w", err)
	}

	return nil
}
