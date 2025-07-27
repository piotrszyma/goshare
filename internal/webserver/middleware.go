package webserver

import (
	"log"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides the http.ResponseWriter's WriteHeader method to capture the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs each request and response status
func loggingMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Record the start time
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		handler(wrapped, r)

		// Log the request details
		log.Printf(
			"%s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	}
}
