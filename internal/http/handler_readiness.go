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

package http

import (
	"net/http"

	"github.com/redds-be/reddlinks/internal/json"
)

// HandlerReadiness sends a positive JSON response to indicate its readiness.
func HandlerReadiness(writer http.ResponseWriter, _ *http.Request) {
	// Define a JSON structure for the status
	type statusResponse struct {
		Status string `json:"status"`
	}

	// Respond to the client with the 'Alive.' message at '/status'
	json.RespondWithJSON(writer, http.StatusOK, statusResponse{Status: "Alive."})
}
