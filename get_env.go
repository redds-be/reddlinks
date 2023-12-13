//    rlinks, a simple link shortener written in Go.
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
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/joho/godotenv"
)

type env struct {
	// Define a structure for the env variables
	portStr                string
	instanceName           string
	instanceURL            string
	dbType                 string
	dbURL                  string
	timeBetweenCleanups    int
	defaultLength          int
	defaultMaxLength       int
	defaultMaxCustomLength int
	defaultExpiryTime      int
}

func (env env) envCheck() error { //nolint:funlen,cyclop
	// Check the port
	port, err := strconv.Atoi(env.portStr)
	if err != nil {
		return fmt.Errorf("the port %w", errRead)
	}

	// Check if the instance name isn't null
	if env.instanceName == "" {
		return fmt.Errorf("the instance name %w", errEmpty)
	}

	// Check if the instance URL is valid
	instanceURLMatch, err := regexp.MatchString(`^https?://.*\..*/$`, env.instanceURL)
	if err != nil {
		return fmt.Errorf("the instance URL %w", errNotChecked)
	}
	if env.instanceURL == "" || !instanceURLMatch {
		return fmt.Errorf("the instance URL %w", errInvalid)
	}

	// Check if the port is valid
	const maxPort = 65535
	if port <= 0 {
		return fmt.Errorf("the port %w", errNullOrNegative)
	} else if port > maxPort {
		return fmt.Errorf("the port %w '%d'", errSuperior, maxPort)
	}

	// Check if the database type is valid
	dbTypeMatch, err := regexp.MatchString(`^postgres$|^sqlite3$`, env.dbType)
	if err != nil {
		return fmt.Errorf("the database type %w: %w", errNotChecked, err)
	}
	if env.dbType == "" || !dbTypeMatch {
		return fmt.Errorf("the database type %w", errInvalidOrUnsupported)
	}

	// Check if the database access string is empty or not.
	if env.dbURL == "" {
		return fmt.Errorf("the database access string %w", errEmpty)
	}

	// Check the time between cleanups, can be any time really, so only checking if it's 0
	if env.timeBetweenCleanups <= 0 {
		return fmt.Errorf("the time between database cleanups %w", errNullOrNegative)
	}

	// Check the default short length
	if env.defaultLength <= 0 {
		return fmt.Errorf("the default short length %w", errNullOrNegative)
	} else if env.defaultLength > env.defaultMaxLength {
		return fmt.Errorf("the default short length %w the default max short length", errSuperior)
	}

	// Check the default max custom short length
	if env.defaultMaxCustomLength <= 0 {
		return fmt.Errorf("the default max custom short length %w", errNullOrNegative)
	} else if env.defaultMaxCustomLength > env.defaultMaxLength {
		return fmt.Errorf("the default max custom short %w the default max short length", errSuperior)
	}

	// Check the default max short length
	const maxString = 8000
	switch {
	case env.defaultMaxLength <= 0:
		return fmt.Errorf("the default short length %w", errNullOrNegative)
	case env.defaultMaxLength <= env.defaultLength:
		return fmt.Errorf("the max default short length %w the default short length", errInferior)
	case env.defaultMaxLength < env.defaultMaxCustomLength:
		return fmt.Errorf("the max default short length %w the default max custom short length", errInferior)
	case env.defaultMaxLength > maxString:
		return fmt.Errorf( //nolint:goerr113
			"strangely, some database engines don't support strings over %d chars long"+
				" for fixed-sized strings",
			maxString,
		)
	}

	// Check the default expiry time
	if env.defaultExpiryTime <= 0 {
		return fmt.Errorf("the default expiry time %w", errNullOrNegative)
	}

	// No errors, since everything is fine
	return nil
}

func getEnv(envFile string) env { //nolint:funlen
	// Load the env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the port
	portStr := os.Getenv("RLINKS_PORT")

	// Read the default short length
	defaultLengthStr := os.Getenv("RLINKS_DEF_SHORT_LENGTH")
	defaultLength, err := strconv.Atoi(defaultLengthStr)
	if err != nil {
		log.Fatal("the default length couldn't be read:", err)
	}

	// Read the default max short length
	defaultMaxLengthStr := os.Getenv("RLINKS_MAX_SHORT_LENGTH")
	defaultMaxLength, err := strconv.Atoi(defaultMaxLengthStr)
	if err != nil {
		log.Fatal("the default max length couldn't be read:", err)
	}

	// Read the default max custom short length
	defaultMaxCustomLengthStr := os.Getenv("RLINKS_MAX_CUSTOM_SHORT_LENGTH")
	defaultMaxCustomLength, err := strconv.Atoi(defaultMaxCustomLengthStr)
	if err != nil {
		log.Fatal("the default max custom short length couldn't be read:", err)
	}

	// Read the default expiry time
	defaultExpiryTimeStr := os.Getenv("RLINKS_DEF_EXPIRY_TIME")
	defaultExpiryTime, err := strconv.Atoi(defaultExpiryTimeStr)
	if err != nil {
		log.Fatal("the default expiry time couldn't be read:", err)
	}

	// Read the instance name
	instanceName := os.Getenv("RLINKS_INSTANCE_NAME")

	// Read the instance URL
	instanceURL := os.Getenv("RLINKS_INSTANCE_URL")

	// Read the database type
	dbType := os.Getenv("RLINKS_DB_TYPE")

	// Read the database URL
	dbURL := os.Getenv("RLINKS_DB_STRING")

	// Read the time between cleanup and convert it to an int
	timeBetweenCleanupsStr := os.Getenv("RLINKS_TIME_BETWEEN_DB_CLEANUPS")
	timeBetweenCleanups, err := strconv.Atoi(timeBetweenCleanupsStr)
	if err != nil {
		log.Fatal("the time between database cleanups couldn't be read:", err)
	}

	env := env{
		portStr:                portStr,
		instanceName:           instanceName,
		instanceURL:            instanceURL,
		dbType:                 dbType,
		dbURL:                  dbURL,
		timeBetweenCleanups:    timeBetweenCleanups,
		defaultLength:          defaultLength,
		defaultMaxLength:       defaultMaxLength,
		defaultMaxCustomLength: defaultMaxCustomLength,
		defaultExpiryTime:      defaultExpiryTime,
	}

	// Check the port and the database URL
	err = env.envCheck()
	if err != nil {
		log.Fatal(err)
	}

	return env
}
