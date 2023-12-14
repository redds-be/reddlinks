package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
	"github.com/stretchr/testify/suite"
)

func (s *DBSuite) TestDB() {
	// Testing the creation of the database
	testEnv := getEnv(".env.test")
	testEnv.dbURL = "db_test.db"
	dataBase, err := database.DBConnect(testEnv.dbType, testEnv.dbURL)
	s.Require().NoError(err)

	// Testing the creation of the links table
	err = database.CreateLinksTable(dataBase, testEnv.defaultMaxLength)
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	// Testing the query to get an url by its short
	URL, err := database.GetURLByShort(dataBase, "custom")
	s.Require().NoError(err)
	s.Equal("http://example.com", URL)

	// Testing the query to get a hash by its short
	pass, err := database.GetHashByShort(dataBase, "custom")
	s.Require().NoError(err)
	s.Equal("pass", pass)

	// Testing the query to get all the entries
	links, err := database.GetLinks(dataBase)
	s.Require().NoError(err)
	s.NotEmpty(links)

	// Testing the removal of an entry
	err = database.RemoveLink(dataBase, "custom")
	s.Require().NoError(err)
}

type DBSuite struct {
	suite.Suite
}

func TestDBSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(DBSuite))
}
