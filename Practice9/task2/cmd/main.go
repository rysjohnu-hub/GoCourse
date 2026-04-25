package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"Practice9/task2/handler"
	"Practice9/task2/middleware"
	"Practice9/task2/storage"
)

type TestResult struct {
	RequestNum int
	StatusCode int
	Body       []byte
	Error      error
	Duration   time.Duration
	Timestamp  time.Time
}

func main() {
	fmt.Println("=" + "================================================================")
	fmt.Println("PRACTICE 9 - TASK 2: Loan Repayment with Idempotency")
	fmt.Println("=" + "================================================================")
	fmt.Println()

	store := storage.NewMemoryStore()
	defer store.Close()

	fmt.Println("STORAGE SETUP:")
	fmt.Println("  Type: In-Memory Store")
	fmt.Println("  Purpose: Track idempotent requests")
	fmt.Println()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthCheckHandler)
	mux.HandleFunc("/repay", handler.LoanRepaymentHandler)

	wrappedMux := middleware.ApplyIdempotencyMiddleware(mux, store)

	server := &http.Server{
		Addr:    ":8080",
		Handler: wrappedMux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	fmt.Println("SERVER STARTED:")
	fmt.Println("  Listening on: http://localhost:8080")
	fmt.Println()

	fmt.Println("-" + "================================================================")
	fmt.Println("SCENARIO 1: Sequential Requests (Same Idempotency Key)")
	fmt.Println("-" + "================================================================")
	fmt.Println()

	idempotencyKey := "repay-" + fmt.Sprintf("%d", time.Now().Unix())

	fmt.Printf("Request 1: Sending initial payment request with key '%s'\n", idempotencyKey)
	resp1, err1 := sendRepaymentRequest(idempotencyKey, 1000.0)
	if err1 != nil {
		fmt.Printf("  ERROR: %v\n", err1)
	} else {
		fmt.Printf("  Status: %d\n", resp1.StatusCode)
		fmt.Printf("  Response: %s\n", string(resp1.Body))
	}
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("Request 2: Sending duplicate payment request with same key\n")
	resp2, err2 := sendRepaymentRequest(idempotencyKey, 1000.0)
	if err2 != nil {
		fmt.Printf("  ERROR: %v\n", err2)
	} else {
		fmt.Printf("  Status: %d (Expected: 200 - Cached Result)\n", resp2.StatusCode)
		fmt.Printf("  Response: %s\n", string(resp2.Body))
	}
	fmt.Println()

	fmt.Println("-" + "================================================================")
	fmt.Println("SCENARIO 2: Concurrent Requests (Same Idempotency Key - Race Condition)")
	fmt.Println("-" + "================================================================")
	fmt.Println()

	concurrentKey := "repay-concurrent-" + fmt.Sprintf("%d", time.Now().Unix())
	numConcurrent := 10

	fmt.Printf("Sending %d simultaneous requests with key '%s'\n", numConcurrent, concurrentKey)
	fmt.Println("Expected behavior:")
	fmt.Println("  - 1 request processes normally (200 OK)")
	fmt.Println("  - 9 requests receive 409 Conflict (duplicate in progress)")
	fmt.Println()

	results := make([]TestResult, numConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()

	for i := 1; i <= numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			reqStartTime := time.Now()
			fmt.Printf("[%d] Sending request at %s\n", idx, reqStartTime.Format("15:04:05.000"))

			resp, err := sendRepaymentRequest(concurrentKey, 1000.0)

			duration := time.Since(reqStartTime)

			mu.Lock()
			results[idx-1] = TestResult{
				RequestNum: idx,
				StatusCode: resp.StatusCode,
				Body:       resp.Body,
				Error:      err,
				Duration:   duration,
				Timestamp:  reqStartTime,
			}
			mu.Unlock()

			if err != nil {
				fmt.Printf("[%d] ERROR: %v\n", idx, err)
			} else {
				fmt.Printf("[%d] Response status: %d (took %v)\n", idx, resp.StatusCode, duration)
			}
		}(i)

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	fmt.Println()
	fmt.Println("RESULTS SUMMARY:")
	fmt.Println("-" + "================================================================")

	statusCounts := make(map[int]int)
	var uniqueTransactions []string

	for _, result := range results {
		if result.Error == nil {
			statusCounts[result.StatusCode]++

			var resp handler.LoanRepaymentResponse
			if err := json.Unmarshal(result.Body, &resp); err == nil {
				uniqueTransactions = append(uniqueTransactions, resp.TransactionID)
			}
		}
	}

	fmt.Printf("Total requests: %d\n", numConcurrent)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Println()

	fmt.Println("Status Code Distribution:")
	for status, count := range statusCounts {
		percentage := float64(count) / float64(numConcurrent) * 100
		fmt.Printf("  HTTP %d: %d requests (%.1f%%)\n", status, count, percentage)
	}
	fmt.Println()

	fmt.Println("Transaction ID Analysis:")
	fmt.Printf("  Total successful responses: %d\n", statusCounts[200]+statusCounts[409])
	fmt.Printf("  Unique transaction IDs: %d\n", len(uniqueTransactions))

	if len(uniqueTransactions) == 1 {
		fmt.Printf("  ✓ All responses share the same transaction ID: %s\n", uniqueTransactions[0])
		fmt.Println("  ✓ IDEMPOTENCY VERIFIED: Only one operation was executed")
	} else if len(uniqueTransactions) > 1 {
		fmt.Println("  ✗ IDEMPOTENCY FAILED: Multiple different transaction IDs detected")
		for i, txn := range uniqueTransactions {
			fmt.Printf("    %d. %s\n", i+1, txn)
		}
	}

	fmt.Println()
	fmt.Println("=" + "================================================================")
	fmt.Println("DEMONSTRATION COMPLETE")
	fmt.Println("=" + "================================================================")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func sendRepaymentRequest(idempotencyKey string, amount float64) (*ResponseWithBody, error) {
	reqBody := handler.LoanRepaymentRequest{
		Amount: amount,
		LoanID: "loan-12345",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/repay",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &ResponseWithBody{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

type ResponseWithBody struct {
	StatusCode int
	Body       []byte
}
