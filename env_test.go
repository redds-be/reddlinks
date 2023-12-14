//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2023 redd
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

package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *EnvSuite) TestEmptyOrZeroEnv() {
	// Test if the returned env is empty
	s.NotEmpty(getEnv(".env.test"))
	s.NotZero(getEnv(".env.test"))
}

func (s *EnvSuite) TestAreValuesFromFiles() {
	ValidEnv := env{
		portStr:                "8080",
		instanceName:           "tester",
		instanceURL:            "http://127.0.0.1:8080/",
		dbType:                 "sqlite3",
		dbURL:                  "test.db",
		timeBetweenCleanups:    1,
		defaultLength:          6,
		defaultMaxLength:       255,
		defaultMaxCustomLength: 255,
		defaultExpiryTime:      2880,
	}

	envToCheck := getEnv(".env.test")

	// Test if the values of the env corresponds to these valid values
	s.EqualValues(ValidEnv, envToCheck)
}

func (s *EnvSuite) TestIsErrorForCorrectEnv() {
	envToCheck := env{
		portStr:                "8080",
		instanceName:           "tester",
		instanceURL:            "http://127.0.0.1:8080/",
		dbType:                 "sqlite3",
		dbURL:                  "test.db",
		timeBetweenCleanups:    1,
		defaultLength:          6,
		defaultMaxLength:       255,
		defaultMaxCustomLength: 255,
		defaultExpiryTime:      2880,
	}

	err := envToCheck.envCheck()
	s.Require().NoError(err)
}

func (s *EnvSuite) TestAreErrorsCorrect() { //nolint:funlen
	envToCheck := env{
		portStr:                "8080",
		instanceName:           "tester",
		instanceURL:            "http://127.0.0.1:8080/",
		dbType:                 "sqlite3",
		dbURL:                  "test.db",
		timeBetweenCleanups:    1,
		defaultLength:          6,
		defaultMaxLength:       255,
		defaultMaxCustomLength: 255,
		defaultExpiryTime:      2880,
	}

	// Test if the port errors are correct
	envToCheck.portStr = ""
	err := envToCheck.envCheck()
	s.Require().ErrorIs(err, errRead)

	envToCheck.portStr = "hello"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errRead)

	envToCheck.portStr = "65536"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	envToCheck.portStr = "0"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.portStr = "-3"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	// Reset the port
	envToCheck.portStr = "8080"

	// Test if the instance name errors are correct
	envToCheck.instanceName = ""
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errEmpty)

	// Reset the instance name
	envToCheck.instanceName = "tester"

	// Test if the instance URL errors are correct
	envToCheck.instanceURL = ""
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "htt://example"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "http://example.com"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "http://ls.example.com"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "https://example.com"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "https://ls.example.com"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	envToCheck.instanceURL = "magnet://ls.example.com/"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalid)

	// Reset the instance URL
	envToCheck.instanceURL = "http://127.0.0.1:8080/"

	// Test if the database type errors are correct
	envToCheck.dbType = ""
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalidOrUnsupported)

	envToCheck.dbType = "mssql"
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errInvalidOrUnsupported)

	// Reset the database type
	envToCheck.dbType = "postgres"

	// Test if the database string errors are correct
	envToCheck.dbURL = ""
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errEmpty)

	// Reset the database string
	envToCheck.dbURL = "postgres://user:pass@localhost:5432/db"

	// Test if the time between cleanups errors are correct
	envToCheck.timeBetweenCleanups = 0
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.timeBetweenCleanups = -6
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	// Reset the time between cleanups
	envToCheck.timeBetweenCleanups = 1

	// Test if the default length errors are correct
	envToCheck.defaultLength = 0
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.defaultLength = -9
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.defaultLength = 256
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	// Reset the default length
	envToCheck.defaultLength = 6

	// Test if the default max length errors are correct, fun fact, 3 of the errors can technically never be encountered
	envToCheck.defaultMaxLength = 0
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	envToCheck.defaultMaxLength = -9
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	envToCheck.defaultMaxLength = 20
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	envToCheck.defaultMaxLength = 8001
	err = envToCheck.envCheck()
	s.Require().Error(err)

	// Reset the default max length
	envToCheck.defaultMaxLength = 255

	// Test if the default max length errors are correct
	envToCheck.defaultMaxCustomLength = 0
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.defaultMaxCustomLength = -9
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.defaultMaxCustomLength = 300
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errSuperior)

	// Reset the default max length
	envToCheck.defaultMaxCustomLength = 255

	// Test if the default expiry time are correct
	envToCheck.defaultExpiryTime = 0
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)

	envToCheck.defaultExpiryTime = -17
	err = envToCheck.envCheck()
	s.Require().ErrorIs(err, errNullOrNegative)
}

type EnvSuite struct {
	suite.Suite
}

func TestEnvSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EnvSuite))
}
