package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// LoggingMiddleware : Middleware Function
// Logs the route and time each endpoint was accessed
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s | [%s] | %s\n", time.Now().Format("January-02-2006 | 3:04:5 PM"), r.Method, r.RequestURI)

		next.ServeHTTP(w, r)
	})
}
