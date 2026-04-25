package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LoanRepaymentRequest struct {
	Amount float64 `json:"amount"`
	LoanID string  `json:"loan_id"`
}

type LoanRepaymentResponse struct {
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	TransactionID string    `json:"transaction_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func LoanRepaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoanRepaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("[HANDLER] Processing repayment: amount=%v, loanID=%s\n", req.Amount, req.LoanID)

	fmt.Println("[HANDLER] Simulating payment processing (2 seconds)...")
	time.Sleep(2 * time.Second)
	fmt.Println("[HANDLER] Payment processing completed")

	resp := LoanRepaymentResponse{
		Status:        "paid",
		Amount:        req.Amount,
		TransactionID: "txn-" + uuid.New().String(),
		Timestamp:     time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("[HANDLER] Error encoding response: %v\n", err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
