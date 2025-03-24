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

package utils_test

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite utilsTestSuite) TestCollectGarbage() {
	// Prepare the database needed for garbage collection
	testEnv := env.GetEnv("../.env.test")
	testEnv.DBURL = "utils_test.db"
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

	err = database.CreateLinksTable(dataBase, testEnv.DBType, testEnv.DefaultMaxLength)
	suite.a.AssertNoErr(err)
	err = database.CreateLink(
		dataBase,
		uuid.New(),
		time.Now().UTC(),
		time.Now().UTC(),
		"http://example.com",
		"garbage",
		"pass",
	)
	suite.a.AssertNoErr(err)

	// Test the execution of collectGarbage()
	conf := &utils.Configuration{DB: dataBase}
	err = conf.CollectGarbage()
	suite.a.AssertNoErr(err)
}

func (suite utilsTestSuite) TestDecodeJSON() {
	// Set the parameters to encode, decodeJSON() will be expected to return exactly the same values
	paramsToEncode := utils.Parameters{
		URL:         "http://example.com",
		Length:      6,
		Path:        "apath",
		ExpireAfter: "2d",
		Password:    "pass",
	}

	// Encore de parameters
	enc, err := json.Marshal(paramsToEncode)
	suite.a.AssertNoErr(err)

	// Mock request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/",
		bytes.NewBuffer(enc),
	)
	suite.a.AssertNoErr(err)

	// Test the decodeJSON() function and compare its return value to the expected values
	decodedParams, err := utils.DecodeJSON(req)
	suite.a.AssertNoErr(err)
	suite.a.Assert(decodedParams, paramsToEncode)
}

func (suite utilsTestSuite) TestGenStr() {
	// Test random char generation
	const testLength = 6

	randStr := utils.GenStr(testLength, "ABC")

	suite.a.Assert(len(randStr), testLength)

	if !strings.Contains(randStr, "A") && !strings.Contains(randStr, "B") &&
		!strings.Contains(randStr, "C") {
		suite.t.Errorf("%s does not contain either A, B, or C.", randStr)
	}
}

func (suite utilsTestSuite) TestIsURL() {
	err := utils.IsURL("http://example.com")
	suite.a.AssertNoErr(err)

	err = utils.IsURL("https://example.com")
	suite.a.AssertNoErr(err)

	err = utils.IsURL("https://localhost")
	suite.a.AssertNoErr(err)

	err = utils.IsURL("hts://example.com")
	suite.a.AssertErr(err)

	err = utils.IsURL("ko")
	suite.a.AssertErr(err)
}

func (suite utilsTestSuite) TestGetLocales() {
	// Test GetLocales and check for errors
	var notEmbedded embed.FS
	locales, _, err := utils.GetLocales("./locales/", notEmbedded)
	suite.a.AssertNoErr(err)

	// Verify the translation
	expectedLocale := utils.PageLocaleTl{
		Title:         "Shorten URL",
		AltGitHubLogo: "GitHub Logo",
		Source:        "Source",
	}
	suite.a.Assert(locales["en"], expectedLocale)
}

func (suite utilsTestSuite) TestGetLocale() {
	// Test GetLocale with english
	var notEmbedded embed.FS
	locales, supportedLocales, err := utils.GetLocales("./locales/", notEmbedded)
	suite.a.AssertNoErr(err)

	conf := utils.Configuration{
		Locales:          locales,
		SupportedLocales: supportedLocales,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US")

	expectedLocale := utils.PageLocaleTl{
		Title:         "Shorten URL",
		AltGitHubLogo: "GitHub Logo",
		Source:        "Source",
	}

	locale := utils.GetLocale(req, conf)
	suite.a.Assert(locale, expectedLocale)

	// Test GetLocale with unsupported locale
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zz-ZZ")
	locale = utils.GetLocale(req, conf)
	suite.a.Assert(locale, expectedLocale)

	// Test GetLocale with french
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-FR")

	expectedLocale = utils.PageLocaleTl{
		Title:         "Raccourcir URL",
		AltGitHubLogo: "Logo GitHub",
		Source:        "Source",
	}

	locale = utils.GetLocale(req, conf)
	suite.a.Assert(locale, expectedLocale)
}

// Test suite structure.
type utilsTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestUtilsSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := utilsTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestCollectGarbage()
	suite.TestDecodeJSON()
	suite.TestGenStr()
	suite.TestIsURL()
	suite.TestGetLocales()
	suite.TestGetLocale()
}
