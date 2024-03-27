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
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	HTTP "github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/stretchr/testify/suite"
)

func (s *frontSuite) TestRenderTemplate() {
	HTTP.Templates = template.Must(template.ParseFiles("test.html"))

	page := HTTP.Page{
		InstanceTitle:          "test",
		InstanceURL:            "test.com",
		ShortenedLink:          "shortenedtest",
		Short:                  "shorttest",
		URL:                    "testurl",
		ExpireAt:               "expireattest",
		Password:               "testpassword",
		Error:                  "testerror",
		AddInfo:                "addtestinfo",
		Version:                "testversion",
		DefaultShortLength:     1,
		DefaultMaxShortLength:  2,
		DefaultMaxCustomLength: 3,
		DefaultExpiryTime:      4,
		ContactEmail:           "test AT test DOT test",
	}

	// Test if template rendering works
	resp := httptest.NewRecorder()

	HTTP.RenderTemplate(resp, "test", &page)

	s.Equal(http.StatusOK, resp.Code)
	s.Equal("nosniff", resp.Header().Get("X-Content-Type-Options"))
	s.Equal("text/html; charset=UTF-8", resp.Header().Get("Content-Type"))
	s.Equal("default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self'; img-src 'self'; "+
		"connect-src 'self'; frame-src 'self'; font-src 'self'; media-src 'self'; object-src 'self'; manifest-src "+
		"'self'; worker-src 'self'; form-action 'self'; frame-ancestors 'self'", resp.Header().Get("Content-Security-Policy"))
	s.Equal("<p>InstanceTitle: test</p>\n"+
		"<p>InstanceURL: test.com</p>\n"+
		"<p>ShortenedLink: shortenedtest</p>\n"+
		"<p>Short: shorttest</p>\n"+
		"<p>URL: testurl</p>\n"+
		"<p>ExpireAt: expireattest</p>\n"+
		"<p>Password: testpassword</p>\n"+
		"<p>Error: testerror</p>\n"+
		"<p>AddInfo: addtestinfo</p>\n"+
		"<p>Version: testversion</p>\n"+
		"<p>DefaultShortLength: 1</p>\n"+
		"<p>DefaultMaxShortLength: 2</p>\n"+
		"<p>DefaultMaxCustomLength: 3</p>\n"+
		"<p>DefaultExpiryTime: 4</p>\n"+
		"<p>ContactEmail: test AT test DOT test</p>", resp.Body.String())
}

func (s *frontSuite) TestMainFrontHandlers() { //nolint:funlen
	HTTP.Templates = template.Must(template.ParseFiles("../static/index.html", "../static/add.html",
		"../static/error.html", "../static/pass.html", "../static/privacy.html"))

	testEnv := env.GetEnv(".env.test")
	testEnv.DBURL = "front_test.db"

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

	// Test if the error page works
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	httpAdapter := HTTP.NewAdapter(*conf)
	httpAdapter.FrontErrorPage(resp, req, 400, "Something went wrong.")

	s.Equal(http.StatusBadRequest, resp.Code)

	// Test if the main page works
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontHandlerMainPage(resp, req)

	s.Equal(http.StatusOK, resp.Code)

	// Test if the privacy page works
	req = httptest.NewRequest(http.MethodGet, "/privacy", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontHandlerPrivacyPage(resp, req)

	s.Equal(http.StatusOK, resp.Code)

	// Test if the password asking page works
	req = httptest.NewRequest(http.MethodGet, "/pass", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontAskForPassword(resp, req)

	s.Equal(http.StatusOK, resp.Code)

	// Test link creation with default values
	params := utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 0,
		Password:    "",
	}

	errMsg, code, _, returnedLink := httpAdapter.FrontCreateLink(params)

	s.Equal("", errMsg)
	s.Equal(http.StatusCreated, code)
	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link creation with custom length for random short
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      12,
		Path:        "",
		ExpireAfter: 0,
		Password:    "",
	}

	errMsg, code, _, returnedLink = httpAdapter.FrontCreateLink(params)

	s.Equal("", errMsg)
	s.Equal(http.StatusCreated, code)
	s.Equal(params.URL, returnedLink.URL)
	s.Len(returnedLink.Short, params.Length)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link creation with a custom short
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "custom",
		ExpireAfter: 0,
		Password:    "",
	}

	errMsg, code, _, returnedLink = httpAdapter.FrontCreateLink(params)

	s.Equal("", errMsg)
	s.Equal(http.StatusCreated, code)
	s.Equal(params.URL, returnedLink.URL)
	s.Equal(params.Path, returnedLink.Short)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link creation with custom expiration time
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 5,
		Password:    "",
	}

	errMsg, code, _, returnedLink = httpAdapter.FrontCreateLink(params)

	s.Equal("", errMsg)
	s.Equal(http.StatusCreated, code)
	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(params.ExpireAfter)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test link creation with a password
	params = utils.Parameters{
		URL:         "http://example.com/",
		Length:      0,
		Path:        "",
		ExpireAfter: 0,
		Password:    "secret",
	}

	errMsg, code, _, returnedLink = httpAdapter.FrontCreateLink(params)

	s.Equal("", errMsg)
	s.Equal(http.StatusCreated, code)
	s.Equal(params.URL, returnedLink.URL)
	s.Equal(
		time.Now().UTC().Add(time.Duration(conf.DefaultExpiryTime)*time.Minute).Format(time.RFC822),
		returnedLink.ExpireAt.Format(time.RFC822),
	)

	// Test if the front link creation page works
	addForm := url.Values{
		"add":          {"Add"},
		"length":       {"6"},
		"expire_after": {"2880"},
		"url":          {"https://example.com"},
		"short":        {"addpagetest"},
		"password":     {"secret"},
	}

	req = httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(addForm.Encode()))
	resp = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpAdapter.FrontHandlerAdd(resp, req)

	s.Equal(http.StatusCreated, resp.Code)

	// Test if the front link redirection
	redirectForm := url.Values{
		"access":   {"Access"},
		"short":    {"addpagetest"},
		"password": {"secret"},
	}

	req = httptest.NewRequest(http.MethodPost, "/pass", strings.NewReader(redirectForm.Encode()))
	resp = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpAdapter.FrontHandlerRedirectToURL(resp, req)

	s.Equal(http.StatusSeeOther, resp.Code)
}

type frontSuite struct {
	suite.Suite
}

func TestFrontSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(frontSuite))
}
