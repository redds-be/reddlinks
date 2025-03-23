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
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redds-be/reddlinks/internal/database"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

var (
	ErrInvalidURLScheme = errors.New("URL scheme is invalid")
	errInvalidHost      = errors.New("URL host is invalid")
	ErrEmpty            = errors.New("can't be empty")
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
	SupportedLocales       []string
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
}

// GetLocales parses locale json files and return them as structs
//
// It gets a list of locale files, parses them, add locale as supported in a slice and return a map of PageLocaleTL struct.
func GetLocales(localesDir string) (map[string]PageLocaleTl, []string, error) {
	// Get locales file list
	localeFileList, err := os.ReadDir(localesDir)
	if err != nil {
		return make(map[string]PageLocaleTl), nil, err
	}

	locales := map[string]PageLocaleTl{}
	var supportedLocales []string //nolint:prealloc

	// Get the locale file
	for _, localeFile := range localeFileList {
		lang := strings.TrimSuffix(localeFile.Name(), ".json")

		locales[lang] = PageLocaleTl{}
		locale := PageLocaleTl{}

		jsonLocaleFile, err := os.Open(localesDir + localeFile.Name())
		if err != nil {
			return make(map[string]PageLocaleTl), nil, err
		}

		// Decode locale file
		decoder := json.NewDecoder(jsonLocaleFile)
		err = decoder.Decode(&locale)
		if err != nil {
			return make(map[string]PageLocaleTl), nil, err
		}

		locales[lang] = locale

		supportedLocales = append(supportedLocales, lang)
	}

	return locales, supportedLocales, err
}

// CollectGarbage deletes old expired entries in the database.
//
// It calls [database.RemoveExpiredLinks] which will delete expired links.
// As of now, the necessity of this function is questionable.
func (conf Configuration) CollectGarbage() error {
	// Delete expired links
	err := database.RemoveExpiredLinks(conf.DB)
	if err != nil {
		return err
	}

	return nil
}

// DecodeJSON returns a [utils.Parameters] struct that contains the decoded clients's JSON request.
//
// It creates a decoder using [json.NewDecoder], using this decoder,
// the function decodes the client's JSON and store it in the [utils.Parameters] struct to then be returned.
// As of now, the necessity of keeping the function in utils rather json is questionable.
func DecodeJSON(r *http.Request) (Parameters, error) {
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	return params, err
}

// GenStr returns a string of a set length composed of a specific charset.
//
// It first creates a byte map of a set length, then, for the length of the map,
// select a random char from the charset to be added the map at the actual index of the iteration.
// After all is done, the map is converted into a string while being returned.
func GenStr(length int, charset string) string {
	// Create an empty map for the future string
	randomByteStr := make([]byte, length)

	// For the length of the empty string, append a random character within the charset
	for i := range randomByteStr {
		randomByteStr[i] = charset[rand.New( //nolint:gosec
			rand.NewSource(time.Now().UnixNano())).Intn(len(charset))]
	}

	// Convert and return the generated string
	return string(randomByteStr)
}

// IsURL verifies if an input string is a valid HTTP(s) URL.
func IsURL(source string) error {
	if source == "" {
		return ErrEmpty
	}

	URL, err := url.ParseRequestURI(source)
	if err != nil {
		return err
	}

	if URL.Scheme != "http" && URL.Scheme != "https" {
		return ErrInvalidURLScheme
	}

	address := net.ParseIP(URL.Host)
	if address != nil {
		return errInvalidHost
	}

	return nil
}

// TextToB64QR transforms the source string into a base64 encoded QR.
func TextToB64QR(content string) (string, error) {
	qrc, err := qrcode.NewWith(content,
		qrcode.WithEncodingMode(qrcode.EncModeByte),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart),
	)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	writer := emptyCloser{buf}
	image := standard.NewWithWriter(writer, standard.WithQRWidth(40)) //nolint:mnd
	if err := qrc.Save(image); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

type emptyCloser struct {
	io.Writer
}

func (emptyCloser) Close() error { return nil }
