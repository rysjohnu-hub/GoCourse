package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"Practice9/task1/payment"
	"Practice9/task1/retry"
)

type RequestCounter struct {
	mu    sync.Mutex
	count int
}

func (rc *RequestCounter) Increment() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.count++
	return rc.count
}

func main() {
	fmt.Println("=" + "================================================================")
	fmt.Println("PRACTICE 9 - TASK 1: Resilient HTTP Client with Retry")
	fmt.Println("=" + "================================================================")
	fmt.Println()

	rand.Seed(time.Now().UnixNano())

	counter := &RequestCounter{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := counter.Increment()

		if attempt <= 3 {
			fmt.Printf("[SERVER] Request #%d: Returning 503 Service Unavailable\n", attempt)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
			return
		}

		fmt.Printf("[SERVER] Request #%d: Returning 200 OK with payment response\n", attempt)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := payment.PaymentResponse{
			Status:        "success",
			TransactionID: "txn-12345-67890",
			Amount:        1000.00,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fmt.Println("TEST SERVER SETUP:")
	fmt.Printf("  URL: %s\n", server.URL)
	fmt.Println("  Behavior: 503 (x3) → 200 OK")
	fmt.Println()

	backoffCfg := retry.BackoffConfig{
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		MaxRetries: 5,
	}

	client := payment.NewPaymentClient(backoffCfg)
	fmt.Println("CLIENT CONFIGURATION:")
	fmt.Printf("  Max Retries: %d\n", backoffCfg.MaxRetries)
	fmt.Printf("  Base Delay: %v\n", backoffCfg.BaseDelay)
	fmt.Printf("  Max Delay: %v\n", backoffCfg.MaxDelay)
	fmt.Println()

	paymentReq := payment.PaymentRequest{
		Amount:   1000.00,
		Currency: "USD",
		OrderID:  "order-12345",
	}

	fmt.Println("EXECUTION:")
	fmt.Println("-" + strings.Repeat("-", 79))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()

	resp, err := client.ExecutePaymentWithBody(ctx, server.URL, paymentReq)

	elapsedTime := time.Since(startTime)

	fmt.Println("-" + strings.Repeat("-", 79))
	fmt.Println()

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Println("SUCCESS!")
		fmt.Printf("  Status: %s\n", resp.Status)
		fmt.Printf("  Transaction ID: %s\n", resp.TransactionID)
		fmt.Printf("  Amount: $%.2f\n", resp.Amount)
		fmt.Printf("  Total Attempts: %d\n", counter.count)
		fmt.Printf("  Total Time: %v\n", elapsedTime)
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("DEMONSTRATION COMPLETE")
	fmt.Println("=" + strings.Repeat("=", 79))
}
