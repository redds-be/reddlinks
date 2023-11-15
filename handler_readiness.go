package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, _ *http.Request) {
	respondWithJSON(w, 200, struct{}{})
}
