// Package env provides functionality for loading and validating environment
// configuration from both .env files and system environment variables.
package env

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/redds-be/reddlinks/internal/utils"
)

// Env defines the configuration settings for the application.
type Env struct {
	AddrAndPort            string // Address and port the server listens on (format: "host:port")
	InstanceName           string // Name of this instance
	InstanceURL            string // Base URL where this instance is accessible
	DBType                 string // Database type ("postgres" or "sqlite")
	DBUser                 string // Database username
	DBPass                 string // Database password
	DBHost                 string // Database host address
	DBPort                 string // Database port
	DBName                 string // Database name
	DBURL                  string // Full database connection string (optional, if not provided will be built from other DB fields)
	ContactEmail           string // Admin contact email address
	TimeBetweenCleanups    int    // Time between garbage collection runs (in minutes)
	DefaultLength          int    // Default length for generated short URLs
	DefaultMaxLength       int    // Maximum allowed length for any short URL
	DefaultMaxCustomLength int    // Maximum allowed length for custom short URLs
	DefaultExpiryTime      int    // Default time until links expire (in minutes, 0 for no expiry)
}

// EnvCheck performs comprehensive validation of the environment configuration.
// It ensures that all required fields are present and that their values meet
// the application's requirements in terms of format, range, and consistency.
//
// This method delegates specific validation tasks to specialized helper methods
// that focus on validating related groups of configuration settings.
//
// Returns an error if any validation check fails with a detailed message about
// which validation failed and why. Returns nil if all validations pass.
func (env Env) EnvCheck() error {
	// Validate instance settings
	if err := env.validateInstanceConfig(); err != nil {
		return err
	}

	// Validate database settings
	if err := env.validateDatabaseConfig(); err != nil {
		return err
	}

	if err := env.validateLengthConstraints(); err != nil {
		return err
	}

	return nil
}

// validateInstanceConfig checks the validity of instance name and URL parameters.
// It ensures that:
// - The instance name is not empty
// - The instance URL is properly formatted as a valid URL
//
// Returns an error if any validation fails, nil otherwise.
func (env Env) validateInstanceConfig() error {
	// Check if the instance name isn't null
	if env.InstanceName == "" {
		return fmt.Errorf("the instance name %w", ErrEmpty)
	}

	// Check if the instance URL is valid
	if err := utils.IsURL(env.InstanceURL); err != nil {
		return err
	}

	return nil
}

// validateDatabaseConfig checks the validity of database connection parameters and other database related configs.
// It ensures that:
// - The database type is one of the supported types (postgres or sqlite)
// - The database type is not empty
// - The time between cleanups is positive
//
// Returns an error if any validation fails, nil otherwise.
func (env Env) validateDatabaseConfig() error {
	// Check if the database type is valid
	if env.DBType == "" || !regexp.MustCompile(`^postgres$|^sqlite$`).MatchString(env.DBType) {
		return fmt.Errorf("the database type %w", ErrInvalidOrUnsupported)
	}

	// Check the time between cleanups
	if env.TimeBetweenCleanups <= 0 {
		return fmt.Errorf("the time between database cleanups %w", ErrNullOrNegative)
	}

	return nil
}

// validateLengthConstraints checks the consistency and validity of URL length parameters.
// It ensures that:
// - DefaultLength is positive
// - DefaultMaxCustomLength is positive
// - DefaultMaxLength is positive
// - DefaultLength does not exceed DefaultMaxLength
// - DefaultMaxCustomLength does not exceed DefaultMaxLength
// - DefaultMaxLength is not less than DefaultLength
// - DefaultMaxLength is not less than DefaultMaxCustomLength
// - DefaultMaxLength does not exceed the maximum string length supported by databases
// - DefaultExpiryTime is positive
//
// Returns an error if any validation fails, nil otherwise.
func (env Env) validateLengthConstraints() error {
	// Set max string size of string in db to avoid having a magic number
	const maxStringLength = 8000

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
	switch {
	case env.DefaultMaxLength <= 0:
		return fmt.Errorf("the default short length %w", ErrNullOrNegative)
	case env.DefaultMaxLength < env.DefaultLength:
		return fmt.Errorf("the max default short length %w the default short length", ErrInferior)
	case env.DefaultMaxLength < env.DefaultMaxCustomLength:
		return fmt.Errorf(
			"the max default short length %w the default max custom short length",
			ErrInferior,
		)
	case env.DefaultMaxLength > maxStringLength:
		return fmt.Errorf( //nolint:goerr113
			"strangely, some database engines don't support strings over %d chars long"+
				" for fixed-sized strings",
			maxStringLength,
		)
	}

	// Check the default expiry time
	if env.DefaultExpiryTime < 0 {
		return fmt.Errorf("the default expiry time %w", ErrNegative)
	}

	return nil
}

// GetEnv loads and validates the application's environment configuration.
// It first attempts to load variables from a specified .env file if it exists,
// then falls back to system environment variables. It applies default values
// where appropriate, performs validation on all values, and returns a fully
// populated and validated Env struct.
//
// The .env file is only loaded if it exists at the specified path. If no file
// is found, the function will look for environment variables exported in the
// system environment.
//
// If required environment variables are missing or validation fails, the
// function will terminate the program with a fatal error.
//
// Parameters:
//   - envFile: Path to an optional .env file containing environment variables.
//
// Returns:
//   - A fully populated and validated Env struct with all configuration values.
func GetEnv(envFile string) Env {
	// Set some default numbers as const to not have magic numbers
	const defaultCleanupTime = 1
	const defaultShortLength = 3
	const defaultMaxLength = 12
	const defaultCustomShortLength = 12
	const defaultExpiryTime = 2880

	loadEnvFile(envFile)

	env := Env{
		// Server settings with defaults
		AddrAndPort:  getEnvWithDefault("REDDLINKS_LISTEN_ADDR", "0.0.0.0:8080"),
		InstanceName: getEnvWithDefault("REDDLINKS_INSTANCE_NAME", "reddlinks"),
		InstanceURL:  getRequiredEnv("REDDLINKS_INSTANCE_URL"),

		// Database settings
		DBType: getRequiredEnv("REDDLINKS_DB_TYPE"),
		DBURL:  os.Getenv("REDDLINKS_DB_STRING"),
	}

	// Only require these if no direct DB string is provided
	if env.DBURL == "" {
		env.DBUser = getRequiredEnv("REDDLINKS_DB_USERNAME")
		env.DBPass = getRequiredEnv("REDDLINKS_DB_PASSWORD")
		env.DBHost = getRequiredEnv("REDDLINKS_DB_HOST")
		env.DBPort = getRequiredEnv("REDDLINKS_DB_PORT")
		env.DBName = getRequiredEnv("REDDLINKS_DB_NAME")
	}

	// Add trailing slash to instance URL if missing
	if !strings.HasSuffix(env.InstanceURL, "/") {
		env.InstanceURL += "/"
	}

	// Load numeric values
	env.TimeBetweenCleanups = getEnvAsIntWithDefault("REDDLINKS_TIME_BETWEEN_DB_CLEANUPS", defaultCleanupTime)
	env.DefaultLength = getEnvAsIntWithDefault("REDDLINKS_DEF_SHORT_LENGTH", defaultShortLength)
	env.DefaultMaxLength = getEnvAsIntWithDefault("REDDLINKS_MAX_SHORT_LENGTH", defaultMaxLength)
	env.DefaultMaxCustomLength = getEnvAsIntWithDefault("REDDLINKS_MAX_CUSTOM_SHORT_LENGTH", defaultCustomShortLength)
	env.DefaultExpiryTime = getEnvAsIntWithDefault("REDDLINKS_DEF_EXPIRY_TIME", defaultExpiryTime)

	// Optional values
	env.ContactEmail = os.Getenv("REDDLINKS_CONTACT_EMAIL")

	// Validate the configuration
	if err := env.EnvCheck(); err != nil {
		log.Fatal(err)
	}

	return env
}

// loadEnvFile attempts to load environment variables from a specified .env file.
// If the file exists at the given path, it will be loaded using godotenv.Load().
// If the file does not exist, this function does nothing and returns silently.
// If the file exists but cannot be loaded, the function will terminate the program
// with a fatal error.
//
// Parameters:
//   - envFile: Path to the .env file to load.
func loadEnvFile(envFile string) {
	if _, err := os.Stat(envFile); !errors.Is(err, os.ErrNotExist) {
		if err := godotenv.Load(envFile); err != nil {
			log.Fatal(err)
		}
	}
}

// getEnvWithDefault retrieves the value of an environment variable with a fallback
// default value. If the environment variable is not set or is empty, the function
// returns the provided default value.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//   - defaultValue: The value to return if the environment variable is not set or empty.
//
// Returns:
//   - The value of the environment variable, or the default value if not set.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// getRequiredEnv retrieves the value of a required environment variable.
// If the environment variable is not set or is empty, the function will
// terminate the program with a fatal error.
//
// Parameters:
//   - key: The name of the required environment variable to retrieve.
//
// Returns:
//   - The value of the environment variable.
//
// Fatal error if:
//   - The environment variable is not set or is empty.
func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("reddlinks could not find a value for %s env variable", key)
	}

	return value
}

// getEnvAsIntWithDefault retrieves an environment variable as an integer with
// a fallback default value. If the environment variable is not set or is empty,
// the function returns the provided default value. If the variable is set but
// cannot be converted to an integer, the function will terminate the program
// with a fatal error.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//   - defaultValue: The integer value to return if the environment variable is not set or empty.
//
// Returns:
//   - The integer value of the environment variable, or the default value if not set.
//
// Fatal error if:
//   - The environment variable is set but cannot be converted to an integer.
func getEnvAsIntWithDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("the value for %s couldn't be converted to an integer: %v", key, err)
	}

	return value
}
