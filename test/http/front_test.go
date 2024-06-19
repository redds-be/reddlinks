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

package http_test

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	HTTP "github.com/redds-be/reddlinks/internal/http"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite frontTestSuite) TestRenderTemplate() {
	HTTP.Templates = template.Must(template.ParseFiles("test.tmpl"))

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
		DefaultExpiryDate:      "2006-01-02T15:04",
		ContactEmail:           "test AT test DOT test",
	}

	// Test if template rendering works
	resp := httptest.NewRecorder()

	HTTP.RenderTemplate(resp, "test", &page, http.StatusOK)

	suite.a.Assert(resp.Code, http.StatusOK)
	suite.a.Assert(resp.Header().Get("X-Content-Type-Options"), "nosniff")
	suite.a.Assert(resp.Header().Get("Content-Type"), "text/html; charset=UTF-8")
	suite.a.Assert(
		resp.Header().Get("Content-Security-Policy"),
		"default-src 'self'; script-src 'self' 'unsafe-inline'; "+
			"style-src 'self'; img-src 'self'; "+
			"connect-src 'self'; frame-src 'self'; font-src 'self'; media-src 'self'; object-src 'self'; manifest-src "+
			"'self'; worker-src 'self'; form-action 'self'; frame-ancestors 'self'",
	)
	suite.a.Assert(resp.Body.String(), "<p>InstanceTitle: test</p>\n"+
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
		"<p>DefaultExpiryDate: 2006-01-02T15:04</p>\n"+
		"<p>ContactEmail: test AT test DOT test</p>\n")
}

func (suite frontTestSuite) TestMainFrontHandlers() { //nolint:funlen
	HTTP.Templates = template.Must(
		template.ParseFiles("../../static/index.tmpl", "../../static/add.tmpl",
			"../../static/error.tmpl", "../../static/pass.tmpl", "../../static/privacy.tmpl",
			"../../static/footer.tmpl", "../../static/head.tmpl", "../../static/nav.tmpl"),
	)

	testEnv := env.GetEnv("../.env.test")
	testEnv.DBURL = "front_test.db"

	// If the test db already exists, delete it as it will cause errors
	if _, err := os.Stat(testEnv.DBURL); !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(testEnv.DBURL)
		suite.a.AssertNoErr(err)
	}

	// Prep everything
	dataBase, err := database.DBConnect(
		testEnv.DBType,
		testEnv.DBURL,
		testEnv.DBUser,
		testEnv.DBPass,
		testEnv.DBHost,
		testEnv.DBPort,
		testEnv.DBName,
	)
	suite.a.AssertNoErr(err)

	err = database.CreateLinksTable(dataBase, testEnv.DefaultMaxLength)
	suite.a.AssertNoErr(err)

	conf := &utils.Configuration{
		DB:                     dataBase,
		InstanceName:           testEnv.InstanceName,
		InstanceURL:            testEnv.InstanceURL,
		Version:                "noVersion",
		AddrAndPort:            testEnv.AddrAndPort,
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

	suite.a.Assert(resp.Code, http.StatusBadRequest)

	// Test if the main page works
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontHandlerMainPage(resp, req)

	suite.a.Assert(resp.Code, http.StatusOK)

	// Test if the privacy page works
	req = httptest.NewRequest(http.MethodGet, "/privacy", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontHandlerPrivacyPage(resp, req)

	suite.a.Assert(resp.Code, http.StatusOK)

	// Test if the password asking page works
	req = httptest.NewRequest(http.MethodGet, "/pass", nil)
	resp = httptest.NewRecorder()

	httpAdapter.FrontAskForPassword(resp, req)

	suite.a.Assert(resp.Code, http.StatusOK)

	// Test if the front link creation page works
	addForm := url.Values{
		"add":             {"Add"},
		"length":          {"6"},
		"expire_datetime": {"2000-01-02T12:12"},
		"url":             {"https://example.com"},
		"short":           {"addpagetest"},
		"password":        {"secret"},
	}

	req = httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(addForm.Encode()))
	resp = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpAdapter.FrontHandlerAdd(resp, req)

	suite.a.Assert(resp.Code, http.StatusCreated)

	// Test the front link redirection
	redirectForm := url.Values{
		"access":   {"Access"},
		"short":    {"addpagetest"},
		"password": {"secret"},
	}

	req = httptest.NewRequest(http.MethodPost, "/pass", strings.NewReader(redirectForm.Encode()))
	resp = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpAdapter.FrontHandlerRedirectToURL(resp, req)

	suite.a.Assert(resp.Code, http.StatusSeeOther)

	// Test the front link redirection with a short that does not exist
	redirectForm = url.Values{
		"access":   {"Access"},
		"short":    {"idonotexist"},
		"password": {"secret"},
	}

	req = httptest.NewRequest(http.MethodPost, "/pass", strings.NewReader(redirectForm.Encode()))
	resp = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpAdapter.FrontHandlerRedirectToURL(resp, req)

	suite.a.Assert(resp.Code, http.StatusNotFound)
}

// Test suite structure.
type frontTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestFrontSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := frontTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestRenderTemplate()
	suite.TestMainFrontHandlers()
}
