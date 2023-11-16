package main

import (
	"github.com/redds-be/rlinks/internal/database"
	"net/http"
	"unicode/utf8"
)

func trimFirstRune(s string) string {
	// Remove the first letter of a string
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

type foundHandler func(w http.ResponseWriter, r *http.Request, link database.Link)

func (apiCfg apiConfig) getURL(handler foundHandler) http.HandlerFunc {
	// Get a link from the path and send it to a handler, in this case, the handlerGetLink to then redirect. If the short doesn't exist, tell it to the user
	return func(w http.ResponseWriter, r *http.Request) {
		short := trimFirstRune(r.URL.Path)
		link, err := apiCfg.DB.GetLinkByShort(r.Context(), short)
		if err != nil {
			respondWithError(w, r, 404, "There in not link associated with this path, it probably invalid or expired.")
			return
		}

		handler(w, r, link)
	}
}
