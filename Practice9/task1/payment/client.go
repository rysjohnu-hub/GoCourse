package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"Practice9/task1/retry"
)

type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	OrderID  string  `json:"order_id"`
}

type PaymentResponse struct {
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
}

type PaymentClient struct {
	httpClient *http.Client
	backoffCfg retry.BackoffConfig
}

func NewPaymentClient(backoffCfg retry.BackoffConfig) *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		backoffCfg: backoffCfg,
	}
}

func (pc *PaymentClient) ExecutePayment(ctx context.Context, url string, req PaymentRequest) (*PaymentResponse, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt < pc.backoffCfg.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		body, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := pc.httpClient.Do(httpReq)

		if !retry.IsRetryable(resp, err) {
			if resp != nil {
				defer resp.Body.Close()
			}
			return nil, fmt.Errorf("non-retriable error: %w", err)
		}

		lastErr = err
		lastResp = resp

		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			var paymentResp PaymentResponse
			if err := json.Unmarshal(respBody, &paymentResp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			fmt.Printf("Attempt %d: Success! Status: %s, TransactionID: %s\n",
				attempt+1, paymentResp.Status, paymentResp.TransactionID)
			return &paymentResp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == pc.backoffCfg.MaxRetries-1 {
			break
		}

		backoff := retry.CalculateBackoff(attempt, pc.backoffCfg)
		fmt.Printf("Attempt %d failed (status: %v), waiting %v before next retry...\n",
			attempt+1, lastResp.StatusCode, backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("failed after %d retries: last error: %w", pc.backoffCfg.MaxRetries, lastErr)
}

func (pc *PaymentClient) ExecutePaymentWithBody(ctx context.Context, url string, req PaymentRequest) (*PaymentResponse, error) {
	var lastErr error
	var lastResp *http.Response

	rand.Seed(time.Now().UnixNano())

	for attempt := 0; attempt < pc.backoffCfg.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		bodyBytes, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := pc.httpClient.Do(httpReq)

		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			var paymentResp PaymentResponse
			if err := json.Unmarshal(respBody, &paymentResp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			fmt.Printf("Attempt %d: Success! Status: %s, TransactionID: %s\n",
				attempt+1, paymentResp.Status, paymentResp.TransactionID)
			return &paymentResp, nil
		}

		if !retry.IsRetryable(resp, err) {
			if resp != nil {
				defer resp.Body.Close()
			}
			if err != nil {
				return nil, fmt.Errorf("non-retriable error: %w", err)
			}
			return nil, fmt.Errorf("non-retriable HTTP status: %d", resp.StatusCode)
		}

		lastErr = err
		lastResp = resp

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == pc.backoffCfg.MaxRetries-1 {
			break
		}

		backoff := retry.CalculateBackoff(attempt, pc.backoffCfg)
		if lastResp != nil {
			fmt.Printf("Attempt %d failed (status: %v), waiting %v before next retry...\n",
				attempt+1, lastResp.StatusCode, backoff)
		} else if lastErr != nil {
			fmt.Printf("Attempt %d failed (error: %v), waiting %v before next retry...\n",
				attempt+1, lastErr, backoff)
		}

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("failed after %d retries: last error: %w", pc.backoffCfg.MaxRetries, lastErr)
}
