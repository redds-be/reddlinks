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

package env_test

import (
	"testing"

	"github.com/redds-be/reddlinks/internal/env"
	"github.com/redds-be/reddlinks/internal/utils"
	"github.com/redds-be/reddlinks/test/helper"
)

func (suite envTestSuite) TestNotEmptyEnv() {
	// Test if the returned env is empty
	suite.a.AssertNotEmpty(env.GetEnv("../.env.test"), env.Env{})
}

func (suite envTestSuite) TestAreValuesFromFiles() {
	ValidEnv := env.Env{
		AddrAndPort:            "127.0.0.1:8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	envToCheck := env.GetEnv("../.env.test")

	// Test if the values of the env corresponds to these valid values
	suite.a.Assert(envToCheck, ValidEnv)
}

func (suite envTestSuite) TestIsErrorForCorrectEnv() {
	envToCheck := env.Env{
		AddrAndPort:            "127.0.0.1:8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	err := envToCheck.EnvCheck()
	suite.a.AssertNoErr(err)
}

func (suite envTestSuite) TestAreErrorsCorrect() { //nolint:funlen
	envToCheck := env.Env{
		AddrAndPort:            "127.0.0.1:8080",
		InstanceName:           "tester",
		InstanceURL:            "http://127.0.0.1:8080/",
		DBType:                 "sqlite",
		DBURL:                  "test.db",
		TimeBetweenCleanups:    1,
		DefaultLength:          6,
		DefaultMaxLength:       255,
		DefaultMaxCustomLength: 255,
		DefaultExpiryTime:      2880,
	}

	// Test if the instance name errors are correct
	envToCheck.InstanceName = ""
	err := envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrEmpty)

	// Reset the instance name
	envToCheck.InstanceName = "tester"

	// Test if the instance URL errors are correct
	envToCheck.InstanceURL = ""
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, utils.ErrEmpty)

	envToCheck.InstanceURL = "htt://example"
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, utils.ErrInvalidURLScheme)

	envToCheck.InstanceURL = "magnet://ls.example.com/"
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, utils.ErrInvalidURLScheme)

	// Reset the instance URL
	envToCheck.InstanceURL = "http://127.0.0.1:8080/"

	// Test if the database type errors are correct
	envToCheck.DBType = ""
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrInvalidOrUnsupported)

	envToCheck.DBType = "mssql"
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrInvalidOrUnsupported)

	// Reset the database type
	envToCheck.DBType = "postgres"

	// Test if the time between cleanups errors are correct
	envToCheck.TimeBetweenCleanups = 0
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	envToCheck.TimeBetweenCleanups = -6
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	// Reset the time between cleanups
	envToCheck.TimeBetweenCleanups = 1

	// Test if the default length errors are correct
	envToCheck.DefaultLength = 0
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultLength = -9
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultLength = 256
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrSuperior)

	// Reset the default length
	envToCheck.DefaultLength = 6

	// Test if the default max length errors are correct, fun fact, 3 of the errors can technically never be encountered
	envToCheck.DefaultMaxLength = 0
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = -9
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = 20
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrSuperior)

	envToCheck.DefaultMaxLength = 8001
	err = envToCheck.EnvCheck()
	suite.a.AssertErr(err)

	// Reset the default max length
	envToCheck.DefaultMaxLength = 255

	// Test if the default max length errors are correct
	envToCheck.DefaultMaxCustomLength = 0
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultMaxCustomLength = -9
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNullOrNegative)

	envToCheck.DefaultMaxCustomLength = 300
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrSuperior)

	// Reset the default max length
	envToCheck.DefaultMaxCustomLength = 255

	// Test if the default expiry time errors are correct
	envToCheck.DefaultExpiryTime = -17
	err = envToCheck.EnvCheck()
	suite.a.AssertErrIs(err, env.ErrNegative)
}

// Test suite structure.
type envTestSuite struct {
	t *testing.T
	a helper.Adapter
}

func TestEnvSuite(t *testing.T) {
	// Enable parallelism
	t.Parallel()

	// Initialize the helper's adapter
	assertHelper := helper.NewAdapter(t)

	// Initialize the test suite
	suite := envTestSuite{t: t, a: assertHelper}

	// Call the tests
	suite.TestNotEmptyEnv()
	suite.TestAreValuesFromFiles()
	suite.TestIsErrorForCorrectEnv()
	suite.TestAreErrorsCorrect()
}
