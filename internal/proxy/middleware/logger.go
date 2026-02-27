package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// Logger is a middleware that logs every request/response
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		res := newResponseWriter(w)

		next.ServeHTTP(res, r)

		log.Printf("[%s] %s %s %s %d %v",
			start.Format(time.RFC3339),
			r.Method,
			r.Host,
			r.URL.Path,
			res.Status(),
			time.Since(start),
		)
	})
}

func prettyJSON(b []byte) string {
	replacer := strings.NewReplacer("\"", "'", "\n", "")
	return replacer.Replace(string(b))
}
