package retry

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestIsRetryableNetworkTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	<-ctx.Done()

	err := ctx.Err()
	if !IsRetryable(nil, err) {
		t.Error("Should retry on context deadline exceeded")
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"504 Gateway Timeout", http.StatusGatewayTimeout, true},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"401 Unauthorized", http.StatusUnauthorized, false},
		{"403 Forbidden", http.StatusForbidden, false},
		{"404 Not Found", http.StatusNotFound, false},
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			result := IsRetryable(resp, nil)
			if result != tt.expected {
				t.Errorf("StatusCode %d: expected %v, got %v", tt.statusCode, tt.expected, result)
			}
		})
	}
}

func TestCalculateBackoffRange(t *testing.T) {
	cfg := BackoffConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		MaxRetries: 5,
	}

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 0 * time.Millisecond, 100 * time.Millisecond}, // 2^0 = 1 * baseDelay
		{1, 0 * time.Millisecond, 200 * time.Millisecond}, // 2^1 = 2 * baseDelay
		{2, 0 * time.Millisecond, 400 * time.Millisecond}, // 2^2 = 4 * baseDelay
		{3, 0 * time.Millisecond, 800 * time.Millisecond}, // 2^3 = 8 * baseDelay
		{10, 0 * time.Millisecond, 5 * time.Second},       // capped at maxDelay
	}

	for _, tt := range tests {
		t.Run("Backoff calculation", func(t *testing.T) {
			backoff := CalculateBackoff(tt.attempt, cfg)
			if backoff < tt.minExpected || backoff > tt.maxExpected {
				t.Errorf("Attempt %d: expected backoff between %v and %v, got %v",
					tt.attempt, tt.minExpected, tt.maxExpected, backoff)
			}
		})
	}
}

func TestCalculateBackoffJitterRandomness(t *testing.T) {
	cfg := BackoffConfig{
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Second,
		MaxRetries: 5,
	}

	results := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		results[i] = CalculateBackoff(2, cfg)
	}

	varied := false
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			varied = true
			break
		}
	}

	if !varied {
		t.Error("Jitter is not adding randomness - all values are the same")
	}
}
