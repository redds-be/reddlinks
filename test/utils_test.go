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
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redds-be/reddlinks/internal/database"
	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/stretchr/testify/suite"
)

func (s *UtilsSuite) TestCollectGarbage() {
	// Prepare the database needed for garbage collection
	testEnv := env.GetEnv(".env.test")
	testEnv.DBURL = "utils_test.db"
	dataBase, err := database.DBConnect(testEnv.DBType, testEnv.DBURL)
	s.Require().NoError(err)
	err = database.CreateLinksTable(dataBase, testEnv.DefaultMaxLength)
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
	conf := &utils.Configuration{DB: dataBase}
	err = conf.CollectGarbage()
	s.Require().NoError(err)
}

func (s *UtilsSuite) TestDecodeJSON() {
	// Set the parameters to encode, decodeJSON() will be expected to return exactly the same values
	paramsToEncode := utils.Parameters{
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
	decodedParams, err := utils.DecodeJSON(req)
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
