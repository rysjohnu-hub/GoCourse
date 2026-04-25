package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"Practice9/task2/storage"
)

func IdempotencyMiddleware(store storage.IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
				return
			}

			ctx := r.Context()

			if cached, err := store.Get(ctx, key); err == nil && cached != nil {
				if cached.Completed {
					fmt.Printf("[MIDDLEWARE] Key %s: Found completed request, returning cached response (status: %d)\n",
						key, cached.StatusCode)
					w.WriteHeader(cached.StatusCode)
					w.Write(cached.Body)
					return
				} else {
					fmt.Printf("[MIDDLEWARE] Key %s: Request still in progress, returning 409 Conflict\n", key)
					http.Error(w, "Duplicate request in progress", http.StatusConflict)
					return
				}
			}

			success, err := store.StartProcessing(ctx, key, 5*time.Minute)
			if err != nil {
				fmt.Printf("[MIDDLEWARE] Key %s: Storage error: %v\n", key, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !success {
				fmt.Printf("[MIDDLEWARE] Key %s: Another request already started, checking status...\n", key)

				time.Sleep(100 * time.Millisecond)

				if cached, err := store.Get(ctx, key); err == nil && cached != nil && cached.Completed {
					fmt.Printf("[MIDDLEWARE] Key %s: Other request completed, returning cached response (status: %d)\n",
						key, cached.StatusCode)
					w.WriteHeader(cached.StatusCode)
					w.Write(cached.Body)
					return
				}

				fmt.Printf("[MIDDLEWARE] Key %s: Still processing or cache error, returning 409 Conflict\n", key)
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
				return
			}

			fmt.Printf("[MIDDLEWARE] Key %s: Processing started\n", key)

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			if err := store.Finish(ctx, key, recorder.Code, recorder.Body.Bytes(), 24*time.Hour); err != nil {
				fmt.Printf("[MIDDLEWARE] Key %s: Error saving result: %v\n", key, err)
			}

			fmt.Printf("[MIDDLEWARE] Key %s: Request completed (status: %d)\n", key, recorder.Code)

			for k, vals := range recorder.Header() {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(recorder.Code)
			w.Write(recorder.Body.Bytes())
		})
	}
}

func ApplyIdempotencyMiddleware(handler http.Handler, store storage.IdempotencyStore) http.Handler {
	return IdempotencyMiddleware(store)(handler)
}
