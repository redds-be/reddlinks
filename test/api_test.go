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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	HTTP "github.com/redds-be/reddlinks/internal/http"
	"github.com/stretchr/testify/suite"
)

func (s *APISuite) TestReadiness() {
	// Test a POST request on a handler that only accepts GET requests
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/status", nil)
	s.Require().NoError(err)
	resp := httptest.NewRecorder()
	HTTP.HandlerReadiness(resp, req)
	s.Equal(http.StatusMethodNotAllowed, resp.Code)
	s.Equal("{\"error\":\"405 Method Not Allowed.\"}", resp.Body.String())

	// Test a GET request on the same handler
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/status", nil)
	s.Require().NoError(err)
	resp = httptest.NewRecorder()
	HTTP.HandlerReadiness(resp, req)
	s.Equal(http.StatusOK, resp.Code)
	s.Equal("{\"status\":\"Alive.\"}", resp.Body.String())
}

func (s *APISuite) TestErr() {
	// Test a POST request on a handler that only accepts GET requests
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/error", nil)
	s.Require().NoError(err)
	resp := httptest.NewRecorder()
	HTTP.HandlerErr(resp, req)
	s.Equal(http.StatusMethodNotAllowed, resp.Code)
	s.Equal("{\"error\":\"405 Method Not Allowed.\"}", resp.Body.String())

	// Test a GET request on the same handler
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/error", nil)
	s.Require().NoError(err)
	resp = httptest.NewRecorder()
	HTTP.HandlerErr(resp, req)
	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 Something went wrong.\"}", resp.Body.String())
}

type APISuite struct {
	suite.Suite
}

func TestAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APISuite))
}
