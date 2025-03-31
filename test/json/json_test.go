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

package json_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redds-be/reddlinks/internal/json"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite jsonTestSuite) TestRespondWithError() {
	resp := httptest.NewRecorder()
	json.RespondWithError(resp, http.StatusBadRequest, "An error.")
	suite.a.Assert(resp.Code, http.StatusBadRequest)
	suite.a.Assert(resp.Body.String(), "{\"error\":\"400 An error.\"}\n")
}

func (suite jsonTestSuite) TestRespondWithJSON() {
	// Testing a JSON response
	type msg struct {
		Msg string `json:"msg"`
	}
	resp := httptest.NewRecorder()
	json.RespondWithJSON(resp, http.StatusOK, msg{Msg: "OK"})
	suite.a.Assert(resp.Code, http.StatusOK)
	suite.a.Assert(resp.Body.String(), "{\"msg\":\"OK\"}\n")
}

// Test suite structure.
type jsonTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestJSONSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := jsonTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestRespondWithError()
	suite.TestRespondWithJSON()
}
