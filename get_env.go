package main

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"strconv"
)

func envCheck(portStr, instanceName, instanceURL, dbURL string, timeBetweenCleanups int) error {
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
	if port == 0 {
		return errors.New("the port can't be null")
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
	if timeBetweenCleanups == 0 {
		return errors.New("the time between database cleanup can't be null")
	}

	// No errors, since everything is fine
	return nil
}

func getEnv(envFile string) (string, string, string, string, int) {
	// Load the env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the port
	portStr := os.Getenv("PORT")

	// Read the instance name
	instanceName := os.Getenv("INSTANCE_NAME")

	// Read the instance URL
	instanceURL := os.Getenv("INSTANCE_URL")

	// Read the database URL
	dbURL := os.Getenv("DB_URL")

	// Read the time between cleanup and convert it to an int
	timeBetweenCleanupsStr := os.Getenv("TIME_BETWEEN_DB_CLEANUP")
	timeBetweenCleanups, err := strconv.Atoi(timeBetweenCleanupsStr)
	if err != nil {
		log.Fatal("the time between database cleanup couldn't be read")
	}

	// Check the port and the database URL
	err = envCheck(portStr, instanceName, instanceURL, dbURL, timeBetweenCleanups)
	if err != nil {
		log.Fatal(err)
	}

	return portStr, instanceName, instanceURL, dbURL, timeBetweenCleanups
}
