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

package test_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redds-be/reddlinks/internal/json"
	"github.com/stretchr/testify/suite"
)

func (s *JSONSuite) TestRespondWithError() {
	resp := httptest.NewRecorder()
	json.RespondWithError(resp, http.StatusBadRequest, "An error.")
	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 An error.\"}", resp.Body.String())
}

func (s *JSONSuite) TestRespondWithJSON() {
	// Testing a JSON response
	type msg struct {
		Msg string `json:"msg"`
	}
	resp := httptest.NewRecorder()
	json.RespondWithJSON(resp, http.StatusOK, msg{Msg: "OK"})
	s.Equal(http.StatusOK, resp.Code)
	s.Equal("{\"msg\":\"OK\"}", resp.Body.String())
}

type JSONSuite struct {
	suite.Suite
}

func TestJSONSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(JSONSuite))
}
