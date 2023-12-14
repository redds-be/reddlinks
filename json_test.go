package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *JSONSuite) TestRespondWithError() {
	resp := httptest.NewRecorder()
	respondWithError(resp, http.StatusBadRequest, "An error.")
	s.Equal(http.StatusBadRequest, resp.Code)
	s.Equal("{\"error\":\"400 An error.\"}", resp.Body.String())
}

func (s *JSONSuite) TestRespondWithJSON() {
	// Testing a JSON response
	type msg struct {
		Msg string `json:"msg"`
	}
	resp := httptest.NewRecorder()
	respondWithJSON(resp, http.StatusOK, msg{Msg: "OK"})
	s.Equal(http.StatusOK, resp.Code)
	s.Equal("{\"msg\":\"OK\"}", resp.Body.String())
}

type JSONSuite struct {
	suite.Suite
}

func TestJSONSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(JSONSuite))
}
