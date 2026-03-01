package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("[%s] %s %s - Started",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
		)

		wrapped := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s - Completed with status %d - Duration: %v",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
		)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey == "" {
			log.Printf("[%s] Unauthorized request: %s %s - Missing X-API-KEY header",
				time.Now().Format("2006-01-02 15:04:05"),
				r.Method,
				r.RequestURI,
			)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized: missing X-API-KEY header"}`, http.StatusUnauthorized)
			return
		}

		if apiKey != "secret123" {
			log.Printf("[%s] Unauthorized request: %s %s - Invalid X-API-KEY",
				time.Now().Format("2006-01-02 15:04:05"),
				r.Method,
				r.RequestURI,
			)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized: invalid X-API-KEY"}`, http.StatusUnauthorized)
			return
		}

		log.Printf("[%s] Authorized request: %s %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
		)

		next.ServeHTTP(w, r)
	})
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}