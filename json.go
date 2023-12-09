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

func respondWithError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	// Respond the client with a JSON error
	log.Printf("Responding with an error to %s (%s) at '%s' with method '%s':\nError: %s (%d)\n", r.RemoteAddr, r.UserAgent(), r.URL.Path, r.Method, msg, code)

	// Define a JSON structure for the error
	type errResponse struct {
		Error string `json:"error"`
	}

	// Send the JSON the client along with the error code
	respondWithJSON(w, code, errResponse{Error: fmt.Sprintf("%d %s", code, msg)})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Create a JSON response, internal error if it can't
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v\n", payload)
		w.WriteHeader(500)
		return
	}

	// Add the json header to the response so that the client can interpret it as JSON, internal error if it can't
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(dat)
	if err != nil {
		log.Printf("Failed to write JSON response: %v\n", dat)
		w.WriteHeader(500)
		return
	}
}
