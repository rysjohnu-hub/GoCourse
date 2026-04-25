package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var timeoutErr net.Error
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			return true
		}

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return true
		}

		if errors.Is(err, net.ErrClosed) {
			return true
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}

		return false
	}

	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return true
		case http.StatusInternalServerError:
			return true
		case http.StatusBadGateway:
			return true
		case http.StatusServiceUnavailable:
			return true
		case http.StatusGatewayTimeout:
			return true

		case http.StatusBadRequest:
			return false
		case http.StatusUnauthorized:
			return false
		case http.StatusForbidden:
			return false
		case http.StatusNotFound:
			return false

		default:
			return resp.StatusCode >= 500
		}
	}

	return false
}

type BackoffConfig struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	MaxRetries int
}

func CalculateBackoff(attempt int, cfg BackoffConfig) time.Duration {
	if attempt < 0 {
		return 0
	}

	backoff := cfg.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))

	if backoff > cfg.MaxDelay {
		backoff = cfg.MaxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(backoff)))

	return jitter
}

func LogRetry(attempt int, backoff time.Duration, err error) {
	fmt.Printf("Attempt %d failed: %v, waiting %v before next retry...\n", attempt, err, backoff)
}
