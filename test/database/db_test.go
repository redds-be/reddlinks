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

package database_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite dbTestSuite) TestDB() { //nolint:funlen
	// Testing the creation of the database
	testEnv := env.GetEnv("../.env.test")
	testEnv.DBURL = "db_test.db"
	dataBase, err := database.DBConnect(testEnv.DBType, testEnv.DBURL)
	suite.a.AssertNoErr(err)

	// Testing the creation of the database with a random driver
	_, err = database.DBConnect("legitdriver", testEnv.DBURL)
	suite.a.AssertErr(err)

	// Testing the creation of the links table
	err = database.CreateLinksTable(dataBase, testEnv.DefaultMaxLength)
	suite.a.AssertNoErr(err)

	// Testing the creation of a link entry
	err = database.CreateLink(
		dataBase,
		uuid.New(),
		time.Now().UTC(),
		time.Now().UTC(),
		"http://example.com",
		"custom",
		"pass",
	)
	suite.a.AssertNoErr(err)

	// Testing the creation of a link entry that will cause an error
	err = database.CreateLink(
		dataBase,
		uuid.New(),
		time.Now().UTC(),
		time.Now().UTC(),
		"http://example.com",
		"custom",
		"pass",
	)
	suite.a.AssertErr(err)

	// Testing the creation of an expired link
	err = database.CreateLink(
		dataBase,
		uuid.New(),
		time.Now().UTC(),
		time.Now().Add(time.Duration(-1)*time.Hour),
		"http://example.com",
		"willExpire",
		"pass",
	)
	suite.a.AssertErr(err)

	// Testing the query to get an url by its short
	URL, err := database.GetURLByShort(dataBase, "custom")
	suite.a.AssertNoErr(err)
	suite.a.Assert(URL, "http://example.com")

	// Testing the query to get an url by its short that will cause an error
	_, err = database.GetURLByShort(dataBase, "doesnotexist")
	suite.a.AssertErr(err)

	// Testing the query to get a hash by its short
	pass, err := database.GetHashByShort(dataBase, "custom")
	suite.a.AssertNoErr(err)
	suite.a.Assert(pass, "pass")

	// Testing the query to get a hash by its short that will cause an error
	_, err = database.GetHashByShort(dataBase, "doesnotexist")
	suite.a.AssertErr(err)

	// Testing the removal of expired entries
	err = database.RemoveExpiredLinks(dataBase)
	suite.a.AssertNoErr(err)
}

// Test suite structure.
type dbTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestDBSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := dbTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestDB()
}
