package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
)

var secretKey string

func init() {
	// Generate a random secret key at startup
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Failed to generate random secret key:", err)
	}
	secretKey = hex.EncodeToString(bytes)
}

// validateKey checks if the request has a valid key parameter
func validateKey(r *http.Request) bool {
	keys, ok := r.URL.Query()["key"]

	log.Printf("request with keys = %s", keys)

	if !ok || len(keys) == 0 {
		return false
	}
	return keys[0] == secretKey
}

// validateKeyCookie checks if the request has a valid key cookie
func validateKeyCookie(r *http.Request) bool {
	cookie, err := r.Cookie("key")
	if err != nil {
		return false
	}

	log.Printf("request with key cookie = %s", cookie.Value)

	return cookie.Value == secretKey
}

// requireKey is middleware that checks for a valid key cookie
func requireKey(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateKeyCookie(r) {
			http.Error(w, "Unauthorized: invalid or missing key cookie", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
