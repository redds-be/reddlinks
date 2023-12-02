package main

import "unicode/utf8"

func trimFirstRune(s string) string {
	// Remove the first letter of a string (https://go.dev/play/p/ZOZyRORkK82)
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
