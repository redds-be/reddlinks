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
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	HTTP "github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/stretchr/testify/suite"
)

func (s *APISuite) TestReadiness() {
	// Test a GET request on the readiness handler
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	resp := httptest.NewRecorder()
	HTTP.HandlerReadiness(resp, req)
	s.Equal(http.StatusOK, resp.Code)
	s.Equal("{\"status\":\"Alive.\"}", resp.Body.String())
}

func (s *APISuite) TestErr() {
	// Test a GET request on the generic error handler
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	resp := httptest.NewRecorder()
	HTTP.HandlerErr(resp, req)
	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 Something went wrong.\"}", resp.Body.String())
}

func (s *APISuite) TestMainAPIHandlers() { //nolint:funlen,maintidx
	testEnv := env.GetEnv(".env.test")
	testEnv.DBURL = "api_test.db"

	// If the test db already exists, delete it as it will cause errors
	if _, err := os.Stat(testEnv.DBURL); !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(testEnv.DBURL)
		s.Require().NoError(err)
	}

	// Prep everything
	dataBase, err := database.DBConnect(testEnv.DBType, testEnv.DBURL)
	s.Require().NoError(err)

	err = database.CreateLinksTable(dataBase, testEnv.DefaultMaxLength)
	s.Require().NoError(err)

	conf := &utils.Configuration{
		DB:                     dataBase,
		InstanceName:           testEnv.InstanceName,
		InstanceURL:            testEnv.InstanceURL,
		Version:                "noVersion",
		PortSTR:                testEnv.PortStr,
		DefaultShortLength:     testEnv.DefaultLength,
		DefaultMaxShortLength:  testEnv.DefaultMaxLength,
		DefaultMaxCustomLength: testEnv.DefaultMaxCustomLength,
		DefaultExpiryTime:      testEnv.DefaultExpiryTime,
		ContactEmail:           testEnv.ContactEmail,
	}

	httpAdapter := HTTP.NewAdapter(*conf)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", httpAdapter.APICreateLink)
	mux.HandleFunc("GET /{short}", httpAdapter.APIRedirectToURL)

	// Test link creation with default values
	params := utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 0,
		Password:    "",
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder := json.NewDecoder(resp.Body)
	returnedLink := database.Link{}
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link redirection with default values
	req = httptest.NewRequest(http.MethodGet, "/"+returnedLink.Short, nil)
	resp = httptest.NewRecorder()

	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)
	// Test link creation with custom length for random short
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      12,
		Path:        "",
		ExpireAfter: 0,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)
	// s.Equal(params.Length, len(returnedLink.Short))
	s.Len(returnedLink.Short, params.Length)

	// Test link redirection with custom length for random short
	req = httptest.NewRequest(http.MethodGet, "/"+returnedLink.Short, nil)
	resp = httptest.NewRecorder()

	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)

	// Test link creation with a custom short
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "custom",
		ExpireAfter: 0,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)
	s.Equal(len(params.Path), len(returnedLink.Short))
	s.Equal(params.Path, returnedLink.Short)

	// Test link redirection with a custom short
	req = httptest.NewRequest(http.MethodGet, "/"+returnedLink.Short, nil)
	resp = httptest.NewRecorder()

	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)

	// Test link creation with custom expiration time
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 5,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(params.ExpireAfter)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link redirection with custom expiration time
	req = httptest.NewRequest(http.MethodGet, "/"+returnedLink.Short, nil)
	resp = httptest.NewRecorder()

	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)

	// Test link creation with a password
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 0,
		Password:    "secret",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link redirection with a password
	params = utils.Parameters{
		Password: "secret",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodGet, "/"+returnedLink.Short, &buf)
	resp = httptest.NewRecorder()
	req.Header.Add("Content-Type", "application/json")

	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)

	// Test link creation with custom expiration time
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 5,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&returnedLink)
	s.Require().NoError(err)

	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(params.ExpireAfter)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link creation with an invalid url
	params = utils.Parameters{
		URL:         "gopher://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 0,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 The URL is invalid.\"}", resp.Body.String())

	// Test link creation with an invalid custom path
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "cust*m",
		ExpireAfter: 0,
		Password:    "",
	}

	err = json.NewEncoder(&buf).Encode(params)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/", &buf)
	resp = httptest.NewRecorder()
	mux.ServeHTTP(resp, req)

	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 The character '*' is not allowed.\"}", resp.Body.String())
}

type APISuite struct {
	suite.Suite
}

func TestAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APISuite))
}
