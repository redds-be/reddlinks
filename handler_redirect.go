package main

import "net/http"

func handlerRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://go.dev/", http.StatusSeeOther)
}
