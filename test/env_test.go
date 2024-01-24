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
	"testing"

	"github.com/redds-be/reddlinks/internal/env"
	"github.com/stretchr/testify/suite"
)

func (s *EnvSuite) TestEmptyOrZeroEnv() {
	// Test if the returned env is empty
	s.NotEmpty(env.GetEnv(".env.test"))
	s.NotZero(env.GetEnv(".env.test"))
}

func (s *EnvSuite) TestAreValuesFromFiles() {
	ValidEnv := env.Env{
		PortStr:                "8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite3",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	envToCheck := env.GetEnv(".env.test")

	// Test if the values of the env corresponds to these valid values
	s.EqualValues(ValidEnv, envToCheck)
}

func (s *EnvSuite) TestIsErrorForCorrectEnv() {
	envToCheck := env.Env{
		PortStr:                "8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite3",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	err := envToCheck.EnvCheck()
	s.Require().NoError(err)
}

func (s *EnvSuite) TestAreErrorsCorrect() { //nolint:funlen
	envToCheck := env.Env{
		PortStr:                "8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite3",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	// Test if the port errors are correct
	envToCheck.PortStr = ""
	err := envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrRead)

	envToCheck.PortStr = "hello"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrRead)

	envToCheck.PortStr = "65536"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	envToCheck.PortStr = "0"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.PortStr = "-3"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	// Reset the port
	envToCheck.PortStr = "8080"

	// Test if the instance name errors are correct
	envToCheck.InstanceName = ""
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrEmpty)

	// Reset the instance name
	envToCheck.InstanceName = "tester"

	// Test if the instance URL errors are correct
	envToCheck.InstanceURL = ""
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "htt://example"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "http://example.com"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "http://ls.example.com"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "https://example.com"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "https://ls.example.com"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	envToCheck.InstanceURL = "magnet://ls.example.com/"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalid)

	// Reset the instance URL
	envToCheck.InstanceURL = "http://127.0.0.1:8080/"

	// Test if the database type errors are correct
	envToCheck.DBType = ""
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalidOrUnsupported)

	envToCheck.DBType = "mssql"
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrInvalidOrUnsupported)

	// Reset the database type
	envToCheck.DBType = "postgres"

	// Test if the database string errors are correct
	envToCheck.DBURL = ""
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrEmpty)

	// Reset the database string
	envToCheck.DBURL = "postgres://user:pass@localhost:5432/db"

	// Test if the time between cleanups errors are correct
	envToCheck.TimeBetweenCleanups = 0
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.TimeBetweenCleanups = -6
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	// Reset the time between cleanups
	envToCheck.TimeBetweenCleanups = 1

	// Test if the default length errors are correct
	envToCheck.DefaultLength = 0
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultLength = -9
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultLength = 256
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	// Reset the default length
	envToCheck.DefaultLength = 6

	// Test if the default max length errors are correct, fun fact, 3 of the errors can technically never be encountered
	envToCheck.DefaultMaxLength = 0
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = -9
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = 20
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = 8001
	err = envToCheck.EnvCheck()
	s.Require().Error(err)

	// Reset the default max length
	envToCheck.DefaultMaxLength = 255

	// Test if the default max length errors are correct
	envToCheck.DefaultMaxCustomLength = 0
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultMaxCustomLength = -9
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultMaxCustomLength = 300
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrSuperior)

	// Reset the default max length
	envToCheck.DefaultMaxCustomLength = 255

	// Test if the default expiry time are correct
	envToCheck.DefaultExpiryTime = 0
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultExpiryTime = -17
	err = envToCheck.EnvCheck()
	s.Require().ErrorIs(err, env.ErrNullOrNegative)
}

type EnvSuite struct {
	suite.Suite
}

func TestEnvSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EnvSuite))
}