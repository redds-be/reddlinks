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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type errResponse struct {
	// Define a JSON structure for an error
	Error string `json:"error"`
}

func respondWithError(writer http.ResponseWriter, code int, msg string) {
	// Send the JSON the client along with the error code
	respondWithJSON(writer, code, errResponse{Error: fmt.Sprintf("%d %s", code, msg)})
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	// Create a JSON response, internal error if it can't
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v\n", payload)
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Add the json header to the response so that the client can interpret it as JSON, internal error if it can't
	writer.Header().Add("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(code)
	_, err = writer.Write(dat)
	if err != nil {
		log.Printf("Failed to write JSON response: %v\n", dat)
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}
}
