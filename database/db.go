//    rlinks, a simple link shortener written in Go.
//    Copyright (C) 2023 redd
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
	_ "github.com/lib/pq"           // Driver for postgresql
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite
)

func DBConnect(dbType, dbURL string) (*sql.DB, error) {
	db, err := sql.Open(dbType, dbURL)

	return db, err
}
