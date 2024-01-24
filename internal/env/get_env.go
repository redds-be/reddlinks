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

package env

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/joho/godotenv"
)

// Env defines a structure for the env variables.
type Env struct {
	PortStr                string
	InstanceName           string
	InstanceURL            string
	DBType                 string
	DBURL                  string
	TimeBetweenCleanups    int
	DefaultLength          int
	DefaultMaxLength       int
	DefaultMaxCustomLength int
	DefaultExpiryTime      int
}

// EnvCheck checks the values of the Env struct.
func (env Env) EnvCheck() error { //nolint:funlen,cyclop
	// Check the port
	port, err := strconv.Atoi(env.PortStr)
	if err != nil {
		return fmt.Errorf("the port %w", ErrRead)
	}

	// Check if the instance name isn't null
	if env.InstanceName == "" {
		return fmt.Errorf("the instance name %w", ErrEmpty)
	}

	// Check if the instance URL is valid
	instanceURLMatch, err := regexp.MatchString(`^https?://.*\..*/$`, env.InstanceURL)
	if err != nil {
		return fmt.Errorf("the instance URL %w", ErrNotChecked)
	}
	if env.InstanceURL == "" || !instanceURLMatch {
		return fmt.Errorf("the instance URL %w", ErrInvalid)
	}

	// Check if the port is valid
	const maxPort = 65535
	if port <= 0 {
		return fmt.Errorf("the port %w", ErrNullOrNegative)
	} else if port > maxPort {
		return fmt.Errorf("the port %w '%d'", ErrSuperior, maxPort)
	}

	// Check if the database type is valid
	dbTypeMatch, err := regexp.MatchString(`^postgres$|^sqlite3$`, env.DBType)
	if err != nil {
		return fmt.Errorf("the database type %w: %w", ErrNotChecked, err)
	}
	if env.DBType == "" || !dbTypeMatch {
		return fmt.Errorf("the database type %w", ErrInvalidOrUnsupported)
	}

	// Check if the database access string is empty or not.
	if env.DBURL == "" {
		return fmt.Errorf("the database access string %w", ErrEmpty)
	}

	// Check the time between cleanups, can be any time really, so only checking if it's 0
	if env.TimeBetweenCleanups <= 0 {
		return fmt.Errorf("the time between database cleanups %w", ErrNullOrNegative)
	}

	// Check the default short length
	if env.DefaultLength <= 0 {
		return fmt.Errorf("the default short length %w", ErrNullOrNegative)
	} else if env.DefaultLength > env.DefaultMaxLength {
		return fmt.Errorf("the default short length %w the default max short length", ErrSuperior)
	}

	// Check the default max custom short length
	if env.DefaultMaxCustomLength <= 0 {
		return fmt.Errorf("the default max custom short length %w", ErrNullOrNegative)
	} else if env.DefaultMaxCustomLength > env.DefaultMaxLength {
		return fmt.Errorf("the default max custom short %w the default max short length", ErrSuperior)
	}

	// Check the default max short length
	const maxString = 8000
	switch {
	case env.DefaultMaxLength <= 0:
		return fmt.Errorf("the default short length %w", ErrNullOrNegative)
	case env.DefaultMaxLength < env.DefaultLength:
		return fmt.Errorf("the max default short length %w the default short length", ErrInferior)
	case env.DefaultMaxLength < env.DefaultMaxCustomLength:
		return fmt.Errorf("the max default short length %w the default max custom short length", ErrInferior)
	case env.DefaultMaxLength > maxString:
		return fmt.Errorf( //nolint:goerr113
			"strangely, some database engines don't support strings over %d chars long"+
				" for fixed-sized strings",
			maxString,
		)
	}

	// Check the default expiry time
	if env.DefaultExpiryTime <= 0 {
		return fmt.Errorf("the default expiry time %w", ErrNullOrNegative)
	}

	// No errors, since everything is fine
	return nil
}

// GetEnv gets the env variables from given .env.
func GetEnv(envFile string) Env { //nolint:funlen
	// Load the env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the port
	portStr := os.Getenv("REDDLINKS_PORT")

	// Read the default short length
	defaultLengthStr := os.Getenv("REDDLINKS_DEF_SHORT_LENGTH")
	defaultLength, err := strconv.Atoi(defaultLengthStr)
	if err != nil {
		log.Fatal("the default length couldn't be read:", err)
	}

	// Read the default max short length
	defaultMaxLengthStr := os.Getenv("REDDLINKS_MAX_SHORT_LENGTH")
	defaultMaxLength, err := strconv.Atoi(defaultMaxLengthStr)
	if err != nil {
		log.Fatal("the default max length couldn't be read:", err)
	}

	// Read the default max custom short length
	defaultMaxCustomLengthStr := os.Getenv("REDDLINKS_MAX_CUSTOM_SHORT_LENGTH")
	defaultMaxCustomLength, err := strconv.Atoi(defaultMaxCustomLengthStr)
	if err != nil {
		log.Fatal("the default max custom short length couldn't be read:", err)
	}

	// Read the default expiry time
	defaultExpiryTimeStr := os.Getenv("REDDLINKS_DEF_EXPIRY_TIME")
	defaultExpiryTime, err := strconv.Atoi(defaultExpiryTimeStr)
	if err != nil {
		log.Fatal("the default expiry time couldn't be read:", err)
	}

	// Read the instance name
	instanceName := os.Getenv("REDDLINKS_INSTANCE_NAME")

	// Read the instance URL
	instanceURL := os.Getenv("REDDLINKS_INSTANCE_URL")

	// Read the database type
	dbType := os.Getenv("REDDLINKS_DB_TYPE")

	// Read the database URL
	dbURL := os.Getenv("REDDLINKS_DB_STRING")

	// Read the time between cleanup and convert it to an int
	timeBetweenCleanupsStr := os.Getenv("REDDLINKS_TIME_BETWEEN_DB_CLEANUPS")
	timeBetweenCleanups, err := strconv.Atoi(timeBetweenCleanupsStr)
	if err != nil {
		log.Fatal("the time between database cleanups couldn't be read:", err)
	}

	env := Env{
		PortStr:                portStr,
		InstanceName:           instanceName,
		InstanceURL:            instanceURL,
		DBType:                 dbType,
		DBURL:                  dbURL,
		TimeBetweenCleanups:    timeBetweenCleanups,
		DefaultLength:          defaultLength,
		DefaultMaxLength:       defaultMaxLength,
		DefaultMaxCustomLength: defaultMaxCustomLength,
		DefaultExpiryTime:      defaultExpiryTime,
	}

	// Check the port and the database URL
	err = env.EnvCheck()
	if err != nil {
		log.Fatal(err)
	}

	return env
}
