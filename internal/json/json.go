// Package json provides utilities for handling JSON HTTP responses.
package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// bufferPool creates a sync.Pool for reusing byte buffers during JSON marshaling.
var bufferPool = sync.Pool{ //nolint:gochecknoglobals
	New: func() interface{} {
		// The initial buffer capacity can be tuned based on your typical response size
		return new(bytes.Buffer)
	},
}

// ErrResponse defines a standardized structure for error messages.
type ErrResponse struct {
	Error string `json:"error"`
}

// InfoResponse defines the structure for URL shortener information responses.
type InfoResponse struct {
	DstURL    string `json:"dstUrl"`    // The destination URL that the short URL redirects to
	Short     string `json:"short"`     // The shortened URL identifier
	CreatedAt string `json:"createdAt"` // Timestamp when the shortened URL was created
	ExpiresAt string `json:"expiresAt"` // Timestamp when the shortened URL will expire
}

// RespondWithError sends a standardized error response to the client.
// It formats the error message with the HTTP status code and sends it as JSON.
//
// Parameters:
//   - writer: The HTTP response writer to send the response through
//   - code: The HTTP status code to return
//   - msg: The error message to include in the response
func RespondWithError(writer http.ResponseWriter, code int, msg string) {
	RespondWithJSON(writer, code, ErrResponse{Error: fmt.Sprintf("%d %s", code, msg)})
}

// RespondWithJSON marshals the provided payload into JSON and sends it to the client.
//
// Parameters:
//   - writer: The HTTP response writer to send the response through
//   - code: The HTTP status code to return
//   - payload: The data structure to marshal and send as JSON
func RespondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	// Set content type header
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get a buffer from the pool and ensure it's reset
	buf, bufOk := bufferPool.Get().(*bytes.Buffer)
	if !bufOk {
		// Handle the case where the type assertion fails
		buf = &bytes.Buffer{}
	}
	buf.Reset()

	// Ensure the buffer is returned to the pool when we're done
	defer bufferPool.Put(buf)

	// Encode directly into the buffer
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Set status code and write the response
	writer.WriteHeader(code)
	if _, err := writer.Write(buf.Bytes()); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}
