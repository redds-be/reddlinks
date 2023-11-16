package main

import "net/http"

func handlerRedirect(w http.ResponseWriter, r *http.Request, url string) {
	// Redirect the client to a URL
	http.Redirect(w, r, url, http.StatusSeeOther)
}
