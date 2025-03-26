//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2025 redd
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

// Package utils implements functions and structs that does not need their own package.
package utils

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

var (
	// ErrInvalidURLScheme is returned when the URL scheme is not http or https.
	ErrInvalidURLScheme = errors.New("URL scheme is invalid")

	// ErrInvalidHost is returned when the host is an IP address instead of a domain name.
	ErrInvalidHost = errors.New("URL host is invalid")

	// ErrEmpty is returned when the source URL is an empty string.
	ErrEmpty = errors.New("can't be empty")

	// urlRegex is a precompiled regular expression to quickly match http/https URLs.
	urlRegex = regexp.MustCompile(`^https?://`)

	// bufferPool is a sync.Pool for efficiently reusing bytes.Buffer instances.
	bufferPool = sync.Pool{ //nolint:gochecknoglobals
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// Configuration defines what is going to be sent to the handlers.
//
// DB is a pointer to the database connection,
// InstanceName refers to the name of the reddlinks instance,
// InstanceURL refers to the public URL of the reddlinks instance,
// Version refers to the actual version of the reddlinks instance,
// AddrAndPort refers to the listening port and address of the reddlinks instance,
// DefaultShortLength refers to the default length of generated strings for a short URL,
// DefaultMaxShortLength refers to the maximum length of generated strings for a short URL,
// DefaultMaxCustomLength refers to the maximum length of custom strings for a short URL,
// DefaultExpiryTime refers to the default expiry time of links records,
// ContactEmail refers to an optional admin's contact email,
// Static contains the embedded static filesystem.
type Configuration struct {
	DB                     *sql.DB
	InstanceName           string
	InstanceURL            string
	Version                string
	AddrAndPort            string
	DefaultShortLength     int
	DefaultMaxShortLength  int
	DefaultMaxCustomLength int
	DefaultExpiryTime      int
	ContactEmail           string
	Static                 embed.FS
	LocalesDir             string
	Locales                map[string]PageLocaleTl
	SupportedLocales       map[string]bool
}

// Parameters defines the structure of the JSON payload that will be read from the user.
//
// URL is the URL to shorten,
// Length is the length of the string that will be generated,
// Path refers to the custom string used in the shortened URL,
// ExpireAfter refers the time from now after which the link will expire,
// ExpireDate refers to the exact expiration date for the link,
// Password refers to a password to protect a link from being accessed by anybody.
type Parameters struct {
	URL         string `json:"url"`
	Length      int    `json:"length"`
	Path        string `json:"customPath"`
	ExpireAfter string `json:"expireAfter"`
	ExpireDate  string `json:"expireDate"`
	Password    string `json:"password"`
}

// PageLocaleTl defines the translatable text content for web pages.
//
// This struct contains all text elements that appear in the UI, allowing for
// internationalization. Each field represents a specific piece of text on the page,
// and the values are loaded from language-specific JSON files.
//
// Field names correspond to translation keys, and their string values hold the
// translated content for a specific locale.
type PageLocaleTl struct {
	Title                    string `json:"title"`
	AltGitHubLogo            string `json:"alt_GitHub_logo"`
	Source                   string `json:"source"`
	Version                  string `json:"version"`
	DevelopedBy              string `json:"developed_by"`
	LicensedUnder            string `json:"licensed_under"`
	GetThe                   string `json:"get_the"`
	SourceCode               string `json:"source_code"`
	Error                    string `json:"error"`
	GoBack                   string `json:"go_back"`
	PasswordRequired         string `json:"password_required"`
	AccessLink               string `json:"access_link"`
	DestinationURL           string `json:"destination_url"`
	ShortPath                string `json:"short_path"`
	CreationDate             string `json:"creation_date"`
	ExpirationDate           string `json:"expiration_date"`
	Proceed                  string `json:"proceed"`
	EnterURL                 string `json:"enter_url"`
	CustomPathTitle          string `json:"custom_path_title"`
	CustomPath               string `json:"custom_path"`
	Optional                 string `json:"optional"`
	Example                  string `json:"example"`
	IfNoneGivenPath          string `json:"if_none_given_path"`
	Reserved                 string `json:"reserved"`
	LengthTitle              string `json:"length_title"`
	Length                   string `json:"length"`
	DefaultsToLength         string `json:"defaults_to_length"`
	ExpiryDateTitle          string `json:"expiry_date_title"`
	ExpiryDate               string `json:"expiry_date"`
	DateOfExpiry             string `json:"date_of_expiry"`
	DefaultsToExpiry         string `json:"defaults_to_expiry"`
	PasswordTitle            string `json:"password_title"`
	Password                 string `json:"password"`
	Path                     string `json:"path"`
	WillAskPass              string `json:"will_ask_pass"`
	ShortenURL               string `json:"shorten_url"`
	ShortenedLink            string `json:"shortened_link"`
	LinksTo                  string `json:"links_to"`
	AccessiblePass           string `json:"accessible_pass"`
	RevealPass               string `json:"reveal_pass"`
	WillExpireOn             string `json:"will_expire_on"`
	QRAlt                    string `json:"qr_alt"`
	CopyLink                 string `json:"copy_link"`
	ShortenAnotherURL        string `json:"shorten_another_url"`
	CopiedLink               string `json:"copied_link"`
	PasswordRevealed         string `json:"password_revealed"`
	PrivacyPolicy            string `json:"privacy_policy"`
	PrivIntro                string `json:"priv_intro"`
	PrivDirect               string `json:"priv_direct"`
	PrivDirectStored         string `json:"priv_direct_stored"`
	PrivURL                  string `json:"priv_url"`
	PrivPath                 string `json:"priv_path"`
	PrivLength               string `json:"priv_length"`
	PrivExpiration           string `json:"priv_expiration"`
	PrivCreation             string `json:"priv_creation"`
	PrivPassword             string `json:"priv_password"`
	PrivPassive              string `json:"priv_passive"`
	PrivNotLog               string `json:"priv_not_log"`
	PrivUnenforceableNote    string `json:"priv_unenforceable_note"`
	PrivRemoval              string `json:"priv_removal"`
	PrivToRemove             string `json:"priv_to_remove"`
	PrivUnenforceableRemoval string `json:"priv_unenforceable_removal"`
	PrivContact              string `json:"priv_contact"`
	PrivEmail                string `json:"priv_email"`
	PrivIfEmail              string `json:"priv_if_email"`
	PrivObfuscated           string `json:"priv_obfuscated"`
	PrivWarranty             string `json:"priv_warranty"`
	PrivIssues               string `json:"priv_issues"`
	ErrNotFound              string `json:"err_not_found"`
	ErrPassAccess            string `json:"err_pass_access"`
	ErrWrongPass             string `json:"err_wrong_pass"`
	ErrCompHash              string `json:"err_comp_hash"`
	ErrGetInfo               string `json:"err_get_info"`
	ErrInvalidJSON           string `json:"err_invalid_json"`
	ErrUnableCheckURL        string `json:"err_unable_check_url"`
	ErrInvalidURL            string `json:"err_invalid_url"`
	ErrUnableTellEOW         string `json:"err_unable_tell_eow"`
	ErrParseTime             string `json:"err_parse_time"`
	ErrParseExpiry           string `json:"err_parse_expiry"`
	ErrCheckValidPath        string `json:"err_check_valid_path"`
	ErrAlphaNumeric          string `json:"err_alpha_numeric"`
	ErrRedirectionLoop       string `json:"err_redirection_loop"`
	ErrHashPass              string `json:"err_hash_pass"`
	ErrPathInUse             string `json:"err_path_in_use"`
	ErrNoSpaceLeft           string `json:"err_no_space_left"`
	ErrUnableLoadPage        string `json:"err_unable_load_page"`
	ErrUnableReadForm        string `json:"err_unable_read_form"`
	ErrUnableReadLength      string `json:"err_unable_read_length"`
	ErrReadPass              string `json:"err_read_pass"`
	ErrUnableGen             string `json:"err_unable_gen"`
	InfoLengthChange         string `json:"info_length_change"`
}

// localeResult represents the parsed outcome of reading and processing a single locale file.
// It encapsulates the language code, the parsed locale data, and any potential error.
type localeResult struct {
	lang   string
	locale PageLocaleTl
	err    error
}

// readLocaleFile reads and parses a single locale file from either a custom directory or embedded static files.
//
// This function handles the file reading and JSON parsing for a specific locale file. It supports
// two reading strategies:
// 1. Reading from a custom directory
// 2. Reading from embedded static files
//
// Parameters:
//   - file: The directory entry representing the locale file to be read
//   - customLocalesDir: Path to the custom locales directory (if used)
//   - embeddedStatic: Embedded filesystem containing static locale files
//
// Returns:
//   - A localeResult struct containing:
//   - The language code extracted from the filename
//   - The parsed locale data
//   - Any error encountered during reading or parsing
func readLocaleFile(
	localeFile os.DirEntry,
	customLocalesDir string,
	embeddedStatic embed.FS,
) localeResult {
	// Extract language code by removing .json extension
	lang := strings.TrimSuffix(localeFile.Name(), ".json")

	// Variables to store file reading results
	var JSONLocaleFile []byte
	var locale PageLocaleTl
	var err error

	// Read file from appropriate source
	if customLocalesDir != "" {
		// Read from custom directory
		JSONLocaleFile, err = os.ReadFile(filepath.Join(customLocalesDir, localeFile.Name()))
	} else {
		// Read from embedded static files
		JSONLocaleFile, err = embeddedStatic.ReadFile("static/locales/" + localeFile.Name())
	}

	// Parse JSON if file was read successfully
	if err == nil {
		err = json.Unmarshal(JSONLocaleFile, &locale)
	}

	return localeResult{
		lang:   lang,
		locale: locale,
		err:    err,
	}
}

// processLocaleFiles concurrently reads and processes multiple locale files.
//
// This function manages the parallel processing of locale files using goroutines.
// It provides an efficient way to read and parse multiple locale files simultaneously.
//
// Key features:
// - Uses goroutines for concurrent file processing
// - Pre-allocates maps to reduce memory reallocations
// - Stops processing if any file fails to read or parse
//
// Parameters:
//   - localeFiles: List of directory entries representing locale files
//   - customLocalesDir: Path to the custom locales directory (if used)
//   - embeddedStatic: Embedded filesystem containing static locale files
//
// Returns:
//   - Map of parsed locales (language code to locale data)
//   - Map of supported locale codes
//   - Any error encountered during processing
func processLocaleFiles(
	localeFileList []os.DirEntry,
	customLocalesDir string,
	embeddedStatic embed.FS,
) (map[string]PageLocaleTl, map[string]bool, error) {
	// Initialize maps
	locales := make(map[string]PageLocaleTl, len(localeFileList))
	supportedLocales := make(map[string]bool, len(localeFileList))

	// Define a struct to collect results from concurrent file processing
	type localeResult struct {
		lang   string       // Language code extracted from filename
		locale PageLocaleTl // Parsed locale data
		err    error        // Any error encountered during processing
	}

	// Create a buffered channel to collect results from goroutines
	results := make(chan localeResult, len(localeFileList))

	// WaitGroup to synchronize goroutine completion
	var waitGroup sync.WaitGroup

	// Process each locale file
	for _, localeFile := range localeFileList {
		// Increment WaitGroup before starting a new goroutine
		waitGroup.Add(1)
		go func(localeFile os.DirEntry) {
			defer waitGroup.Done()
			results <- localeResult(readLocaleFile(localeFile, customLocalesDir, embeddedStatic))
		}(localeFile)
	}

	// Close results channel when all goroutines are complete
	go func() {
		waitGroup.Wait()
		close(results)
	}()

	// Collect and process results from all goroutines
	for result := range results {
		// Stop processing and return error if any file fails
		if result.err != nil {
			return nil, nil, result.err
		}

		// Store parsed locale and mark as supported
		locales[result.lang] = result.locale
		supportedLocales[result.lang] = true
	}

	return locales, supportedLocales, nil
}

// GetLocales parses locale JSON files concurrently from either a custom directory or embedded static files.
//
// This is the main entry point for loading localization resources. It provides flexibility
// in sourcing locale files by supporting:
// 1. Reading from a custom directory
// 2. Reading from embedded static files
//
// The function reads all JSON files in the specified source, extracts language codes,
// and parses them into a structured format for internationalization.
//
// Performance characteristics:
// - Concurrent file processing
// - Minimal memory overhead
// - Fail-fast error handling
//
// Parameters:
//   - customLocalesDir: Optional path to a directory containing custom locale JSON files
//     If empty, embedded static files will be used
//   - embeddedStatic: An embed.FS containing embedded static locale files
//
// Returns:
//   - Map of locale codes to their parsed PageLocaleTl structs
//   - Map of supported locale codes
//   - Any error encountered during processing
//
// Example:
//
//	locales, supportedLocales, err := GetLocales("/path/to/locales", embeddedFS)
//	if err != nil {
//	    log.Fatal("Failed to load locales:", err)
//	}
func GetLocales(
	customLocalesDir string,
	embeddedStatic embed.FS,
) (map[string]PageLocaleTl, map[string]bool, error) {
	// Determine the source of locale files
	var localeFileList []os.DirEntry
	var err error

	// Determine whether to use custom locales directory or embedded files
	if customLocalesDir != "" {
		// Read files from custom directory
		localeFileList, err = os.ReadDir(customLocalesDir)
	} else {
		// Read files from embedded static files
		localeFileList, err = embeddedStatic.ReadDir("static/locales")
	}
	if err != nil {
		return nil, nil, err
	}

	// Process locale files concurrently
	return processLocaleFiles(localeFileList, customLocalesDir, embeddedStatic)
}

// GetLocale determines the appropriate PageLocaleTl struct based on the client's language preference.
// It extracts the language from the "Accept-Language" HTTP header, checks if it's supported,
// and returns the corresponding locale from the configuration.
// If the language is not specified or not supported, it defaults to English ("en").
//
// Parameters:
//   - req: The HTTP request containing the "Accept-Language" header
//   - locales: A map of PageLocaleTl struct containing all the locales
//   - supportedLocales: A map indicating which locale is supported
//
// Returns:
//   - PageLocaleTl: The localization struct for the determined language
func GetLocale(req *http.Request, locales map[string]PageLocaleTl, supportedLocales map[string]bool) PageLocaleTl {
	// Get the client's main language
	const localeCodeInt = 2
	lang := req.Header.Get("Accept-Language")
	if len(lang) > localeCodeInt {
		lang = lang[:localeCodeInt]
	}

	// Check if lang is supported, else, default to english
	if _, ok := supportedLocales[lang]; !ok {
		lang = "en"
	}

	// Return the locale according to the chose one
	return locales[lang]
}

// CollectGarbage deletes old expired entries in the database.
//
// This method performs database cleanup by removing links that have expired.
// It uses the database.RemoveExpiredLinks function to delete these outdated entries.
//
// The method operates on the Configuration receiver and attempts to remove
// expired links from the associated database.
//
// Returns:
//   - An error if the link removal process fails, otherwise nil.
//
// Note: The necessity of this method may be subject to review in future iterations.
func (conf Configuration) CollectGarbage() error {
	// Delete expired links
	err := database.RemoveExpiredLinks(conf.DB)
	if err != nil {
		return err
	}

	return nil
}

// DecodeJSON decodes the JSON payload from an HTTP request into a Parameters struct.
//
// It handles various error scenarios, including:
//   - Empty request bodies
//   - Invalid JSON formatting
//   - Oversized request payloads (more than 1MB)
//
// Parameters:
//   - req: A pointer to the http.Request containing the JSON payload
//
// Returns:
//   - A Parameters struct populated with decoded JSON data
//   - An error if decoding fails, with detailed error information
func DecodeJSON(req *http.Request) (Parameters, error) {
	const oneMB = 1_048_576
	// Use http.MaxBytesReader to close the connexion if a requests is more than 1MB (prevents DoS attacks)
	req.Body = http.MaxBytesReader(nil, req.Body, oneMB)

	// Read the entire request body into memory
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return Parameters{}, err
	}

	// Use json.Unmarshal to decode the payload
	var params Parameters
	if err := json.Unmarshal(body, &params); err != nil {
		return Parameters{}, err
	}

	// Close the request body
	err = req.Body.Close()
	if err != nil {
		return Parameters{}, err
	}

	// Return the successfully decoded Parameters struct
	return params, nil
}

// GenStr generates a random string of specified length using the given charset.
//
// Parameters:
//   - length: Desired length of the output string
//   - charset: Set of characters to choose from when generating the string
//
// Returns:
//   - A randomly generated string of the specified length
//   - An error if random generation fails
func GenStr(length int, charset string) (string, error) {
	// Pre-allocate a byte slice
	result := make([]byte, length)

	// Create a big integer representing the maximum random value
	// to ensure uniform distribution across the charset
	maxRand := big.NewInt(int64(len(charset)))

	// Generate each character of the string individually
	for char := range length {
		// Generate a random char
		// Use crypto/rand to make the linter shut up
		index, err := rand.Int(rand.Reader, maxRand)
		if err != nil {
			return "", err
		}

		// Select a character from the charset using the random index
		result[char] = charset[index.Int64()]
	}

	return string(result), nil
}

// IsURL validates whether the provided string is a well-formed HTTP or HTTPS URL.
//
// The function performs multiple checks:
//   - Ensures the URL is not empty
//   - Validates the URL scheme (must be http or https)
//   - Checks that the host part is valid
//
// Parameters:
//   - supposedURL: The URL string to validate
//
// Returns:
//   - nil if the URL is valid
//   - An error describing the specific validation failure
func IsURL(supposedURL string) error {
	// Quick early exit for empty strings
	if supposedURL == "" {
		return ErrEmpty
	}

	// Perform a quick regex pre-check before full parsing
	if !urlRegex.MatchString(supposedURL) {
		return ErrInvalidURLScheme
	}

	// Parse the URL
	parsedURL, err := url.ParseRequestURI(supposedURL)
	if err != nil {
		return err
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURLScheme
	}

	// Validate the host part
	address := net.ParseIP(parsedURL.Host)
	if address != nil {
		return ErrInvalidHost
	}

	return nil
}

// TextToB64QR converts a string content into a base64 encoded QR code image.
//
// The function performs the following steps:
// 1. Retrieves a reusable buffer from the sync.Pool
// 2. Creates a QR code with byte encoding and quartic error correction
// 3. Generates a PNG image of the QR code
// 4. Encodes the image to base64
//
// Parameters:
//   - content: The text to be encoded into the QR code
//
// Returns:
//   - A base64 encoded string representation of the QR code image
//   - An error if QR code generation or encoding fails
func TextToB64QR(content string) (string, error) {
	// Retrieve a buffer from the pool and ensure it's reset
	buf, bufOk := bufferPool.Get().(*bytes.Buffer)
	if !bufOk {
		// Handle the case where the type assertion fails
		buf = &bytes.Buffer{}
	}
	// Clear previous contents
	buf.Reset()

	// Ensure the buffer is returned to the pool after use
	defer bufferPool.Put(buf)

	// Create QR code with specific encoding and error correction settings
	// - EncModeByte: Supports full range of character encodings
	// - ErrorCorrectionQuart: Allows up to 25% of the QR code to be restored if damaged
	qrc, err := qrcode.NewWith(content,
		qrcode.WithEncodingMode(qrcode.EncModeByte),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart),
	)
	if err != nil {
		return "", err
	}

	// Define the width of the QR code (pixel size)
	const qrWidth = 40

	// Create a writer that will generate the QR code image
	// Uses standard PNG encoding with specified width
	writer := emptyCloser{buf}
	image := standard.NewWithWriter(writer, standard.WithQRWidth(qrWidth))

	// Generate and save the QR code image to the buffer
	if err := qrc.Save(image); err != nil {
		return "", err
	}

	// Convert the image buffer to a base64 encoded string
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// It satisfies the io.WriteCloser interface required by some image encoding methods.
type emptyCloser struct {
	io.Writer
}

// Close allows the buffer to be used with writers that expect a Closer.
func (emptyCloser) Close() error { return nil }
