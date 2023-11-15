package main

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func envCheck(portStr, dbURL string) error {
	// Check the port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("the port couldn't be read")
	}

	// Check if the port is valid
	if port == 0 {
		return errors.New("the port can't be nil")
	} else if port > 65535 {
		return errors.New("the port cannot be superior to '65535'")
	}

	// Check the database URL, it will only check if it is nil
	if dbURL == "" {
		return errors.New("the database URL can't be nil")
	}

	// No errors, since everything is fine
	return nil
}

func getEnv(envFile string) (string, string) {
	// Load the env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the port
	portStr := os.Getenv("PORT")

	// Read the database URL
	dbURL := os.Getenv("DB_URL")

	// Check the port and the database URL
	err = envCheck(portStr, dbURL)
	if err != nil {
		log.Fatal(err)
	}

	return portStr, dbURL
}
