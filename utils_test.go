package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redds-be/rlinks/database"
	"github.com/stretchr/testify/suite"
)

func (s *UtilsSuite) TestTrimFirstRune() {
	// Test if trimFirstRune() does in fact, trim the first rune
	s.Equal("rimmed", trimFirstRune("trimmed"))
}

func (s *UtilsSuite) TestRandomToken() {
	// Test if randomToken returns something
	s.NotEmpty(randomToken())
	s.NotZero(randomToken())
}

func (s *UtilsSuite) TestCollectGarbage() {
	// Prepare the database needed for garbage collection
	testEnv := getEnv(".env.test")
	testEnv.dbURL = "utils_test.db"
	dataBase, err := database.DBConnect(testEnv.dbType, testEnv.dbURL)
	s.Require().NoError(err)
	err = database.CreateLinksTable(dataBase, testEnv.defaultMaxLength)
	s.Require().NoError(err)
	err = database.CreateLink(
		dataBase,
		uuid.New(),
		time.Now().UTC(),
		time.Now().UTC(),
		"http://example.com",
		"garbage",
		"pass",
	)
	s.Require().NoError(err)

	// Test the execution of collectGarbage()
	conf := &configuration{db: dataBase}
	err = conf.collectGarbage()
	s.Require().NoError(err)
}

func (s *UtilsSuite) TestDecodeJSON() {
	// Set the parameters to encode, decodeJSON() will be expected to return exactly the same values
	paramsToEncode := parameters{
		URL:         "http://example.com",
		Length:      6,
		Path:        "apath",
		ExpireAfter: 5,
		Password:    "pass",
	}

	// Encore de parameters
	enc, err := json.Marshal(paramsToEncode)
	s.Require().NoError(err)

	// Mock request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBuffer(enc))
	s.Require().NoError(err)

	// Test the decodeJSON() function and compare its return value to the expected values
	decodedParams, err := decodeJSON(req)
	s.Require().NoError(err)
	s.Equal(paramsToEncode, decodedParams)
}

type UtilsSuite struct {
	suite.Suite
}

func TestUtilsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UtilsSuite))
}
