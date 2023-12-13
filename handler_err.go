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

package main

import (
	"log"
	"net/http"
)

func handlerErr(writer http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)

	// If the method is wrong, return with an error
	if req.Method != http.MethodGet {
		respondWithError(writer, req, http.StatusMethodNotAllowed, "Method Not Allowed.")

		return
	}

	// Respond with a generic error at '/error'
	respondWithError(writer, req, http.StatusBadRequest, "Something went wrong.")
}
