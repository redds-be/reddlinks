package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *APISuite) TestReadiness() {
	// Test a POST request on a handler that only accepts GET requests
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/status", nil)
	s.Require().NoError(err)
	resp := httptest.NewRecorder()
	handlerReadiness(resp, req)
	s.Equal(http.StatusMethodNotAllowed, resp.Code)
	s.Equal("{\"error\":\"405 Method Not Allowed.\"}", resp.Body.String())

	// Test a GET request on the same handler
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/status", nil)
	s.Require().NoError(err)
	resp = httptest.NewRecorder()
	handlerReadiness(resp, req)
	s.Equal(http.StatusOK, resp.Code)
	s.Equal("{\"status\":\"Alive.\"}", resp.Body.String())
}

func (s *APISuite) TestErr() {
	// Test a POST request on a handler that only accepts GET requests
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/error", nil)
	s.Require().NoError(err)
	resp := httptest.NewRecorder()
	handlerErr(resp, req)
	s.Equal(http.StatusMethodNotAllowed, resp.Code)
	s.Equal("{\"error\":\"405 Method Not Allowed.\"}", resp.Body.String())

	// Test a GET request on the same handler
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/error", nil)
	s.Require().NoError(err)
	resp = httptest.NewRecorder()
	handlerErr(resp, req)
	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 Something went wrong.\"}", resp.Body.String())
}

type APISuite struct {
	suite.Suite
}

func TestAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APISuite))
}
