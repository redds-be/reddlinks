package main

import "net/http"

func handlerErr(w http.ResponseWriter, _ *http.Request) {
	respondWithError(w, 400, "Something went wrong.")
}
