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
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"strconv"
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

func (e env) envCheck() error {
	// Check the port
	port, err := strconv.Atoi(e.portStr)
	if err != nil {
		log.Println(err)
		return errors.New("the port couldn't be read")
	}

	// Check if the instance name isn't null
	if e.instanceName == "" {
		return errors.New("the instance name can't be empty")
	}

	// Check if the instance URL is valid
	instanceURLMatch, err := regexp.MatchString(`^https?://.*\..*/$`, e.instanceURL)
	if err != nil {
		log.Println(err)
		return errors.New("the instance URL could not be checked")
	}
	if e.instanceURL == "" || !instanceURLMatch {
		return errors.New("the instance URL is invalid")
	}

	// Check if the port is valid
	if port <= 0 {
		return errors.New("the port can't be null or negative")
	} else if port > 65535 {
		return errors.New("the port cannot be superior to '65535'")
	}

	// Check if the database type is valid
	dbTypeMatch, err := regexp.MatchString(`^postgres$|^sqlite3$`, e.dbType)
	if err != nil {
		log.Println(err)
		return errors.New("the database type could not be checked")
	}
	if e.dbType == "" || !dbTypeMatch {
		return errors.New("the database type is invalid or unsupported")
	}

	// Check if the database access string is empty or not.
	if e.dbURL == "" {
		return errors.New("the database access string can't be empty")
	}

	// Check the time between cleanup, can be any time really, so only checking if it's 0
	if e.timeBetweenCleanups <= 0 {
		return errors.New("the time between database cleanup can't be null or negative")
	}

	// Check the default short length
	if e.defaultLength <= 0 {
		return errors.New("the default short length can't be null or negative")
	} else if e.defaultLength > e.defaultMaxLength {
		return errors.New("the default short length can't be superior to the default max short length")
	}

	// Check the default max custom short length
	if e.defaultMaxCustomLength <= 0 {
		return errors.New("the default max custom short length can't be null or negative")
	} else if e.defaultMaxCustomLength > e.defaultMaxLength {
		return errors.New("the default max custom short length can't be superior to the default max short length")
	}

	// Check the default max short length
	if e.defaultMaxLength <= 0 {
		return errors.New("the default short length can't be null or negative")
	} else if e.defaultMaxLength <= e.defaultLength {
		return errors.New("the max default short length can't be inferior to the default short length")
	} else if e.defaultMaxLength < e.defaultMaxCustomLength {
		return errors.New("the max default short length can't be inferior to the default max custom short length")
	} else if e.defaultMaxLength > 8000 {
		return errors.New("strangely, some database engines don't support strings over 8000 chars long for fixed-size strings")
	}

	// Check the default expiry time
	if e.defaultExpiryTime <= 0 {
		return errors.New("the default expiry time can't be null or negative")
	}

	// No errors, since everything is fine
	return nil
}

func getEnv(envFile string) env {
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
		log.Println(err)
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

	e := env{
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
	err = e.envCheck()
	if err != nil {
		log.Fatal(err)
	}

	return e
}
