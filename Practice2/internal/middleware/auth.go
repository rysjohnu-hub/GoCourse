package middleware

import (
	"log"
	"net/http"
	"time"
)

const API_KEY = "secret12345"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		key := r.Header.Get("X-API-KEY")
		if key != API_KEY {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}

		next.ServeHTTP(w, r)
	}
}