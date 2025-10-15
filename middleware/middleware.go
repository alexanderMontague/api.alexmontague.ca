package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LogRequest logs the HTTP request with optional user context
func LogRequest(r *http.Request) {
	user := GetAuthUser(r)
	timestamp := time.Now().Format("January-02-2006 | 3:04:5 PM")

	if user != nil {
		fmt.Printf("%s | [%s] | %s | UserID: %d\n", timestamp, r.Method, r.RequestURI, user.UserID)
	} else {
		fmt.Printf("%s | [%s] | %s\n", timestamp, r.Method, r.RequestURI)
	}
}

// LoggingMiddleware : Middleware Function
// Logs the route and time each endpoint was accessed
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for /budget routes - they'll be logged by AuthMiddleware after user context is added
		if !strings.HasPrefix(r.URL.Path, "/budget") {
			LogRequest(r)
		}
		next.ServeHTTP(w, r)
	})
}
