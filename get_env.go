package main

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"strconv"
)

func envCheck(portStr, instanceName, instanceURL, dbURL string, timeBetweenCleanups, defaultLength, defaultMaxLength, defaultMaxCustomLength, defaultExpiryTime int) error {
	// Check the port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err)
		return errors.New("the port couldn't be read")
	}

	// Check if the instance name isn't null
	if instanceName == "" {
		return errors.New("the instance name can't be empty")
	}

	// Check if the instance URL is valid
	instanceURLMatch, err := regexp.MatchString(`^https?://.*\..*/$`, instanceURL)
	if err != nil {
		log.Println(err)
		return errors.New("the instance URL could not be checked")
	}
	if instanceURL == "" || !instanceURLMatch {
		return errors.New("the instance URL is invalid")
	}

	// Check if the port is valid
	if port <= 0 {
		return errors.New("the port can't be null or negative")
	} else if port > 65535 {
		return errors.New("the port cannot be superior to '65535'")
	}

	// Check if the database URL is valid.
	// /!\ Will be replaced later by a set of different variables for each of the db information
	dbURLMatch, err := regexp.MatchString(`^postgres://.*:.*@.*:([1-9]|[1-9][0-9]{1,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])/.*$`, dbURL)
	if err != nil {
		log.Println(err)
		return errors.New("the database URL could not be checked")
	}
	if dbURL == "" || !dbURLMatch {
		return errors.New("the database URL is invalid")
	}

	// Check the time between cleanup, can be any time really, so only checking if it's 0
	if timeBetweenCleanups <= 0 {
		return errors.New("the time between database cleanup can't be null or negative")
	}

	// Check the default short length
	if defaultLength <= 0 {
		return errors.New("the default short length can't be null or negative")
	} else if defaultLength > defaultMaxLength {
		return errors.New("the default short length can't be superior to the default max short length")
	}

	// Check the default max custom short length
	if defaultMaxCustomLength <= 0 {
		return errors.New("the default max custom short length can't be null or negative")
	} else if defaultMaxCustomLength > defaultMaxLength {
		return errors.New("the default max custom short length can't be superior to the default max short length")
	}

	// Check the default max short length
	if defaultMaxLength <= 0 {
		return errors.New("the default short length can't be null or negative")
	} else if defaultMaxLength <= defaultLength {
		return errors.New("the max default short length can't be inferior to the default short length")
	} else if defaultMaxLength < defaultMaxCustomLength {
		return errors.New("the max default short length can't be inferior to the default max custom short length")
	} else if defaultMaxLength > 8000 {
		return errors.New("strangely, some database engines don't support strings over 8000 chars long for fixed-size strings")
	}

	// Check the default expiry time
	if defaultExpiryTime <= 0 {
		return errors.New("the default expiry time can't be null or negative")
	}

	// No errors, since everything is fine
	return nil
}

func getEnv(envFile string) (string, int, int, int, int, string, string, string, int) {
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

	// Read the database URL
	dbURL := os.Getenv("RLINKS_DB_URL")

	// Read the time between cleanup and convert it to an int
	timeBetweenCleanupsStr := os.Getenv("RLINKS_TIME_BETWEEN_DB_CLEANUPS")
	timeBetweenCleanups, err := strconv.Atoi(timeBetweenCleanupsStr)
	if err != nil {
		log.Fatal("the time between database cleanups couldn't be read:", err)
	}

	// Check the port and the database URL
	err = envCheck(portStr, instanceName, instanceURL, dbURL, timeBetweenCleanups, defaultLength, defaultMaxLength, defaultMaxCustomLength, defaultExpiryTime)
	if err != nil {
		log.Fatal(err)
	}

	return portStr, defaultLength, defaultMaxLength, defaultMaxCustomLength, defaultExpiryTime, instanceName, instanceURL, dbURL, timeBetweenCleanups
}
