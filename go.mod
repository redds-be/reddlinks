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

module github.com/redds-be/rlinks

go 1.21.4

require (
	github.com/alexedwards/argon2id v1.0.0
	github.com/dchest/uniuri v1.2.0
	github.com/google/uuid v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.18
)

require (
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
