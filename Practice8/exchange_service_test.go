package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/convert", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("from"))
		assert.Equal(t, "EUR", r.URL.Query().Get("to"))

		response := RateResponse{
			Base:   "USD",
			Target: "EUR",
			Rate:   0.92,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")

	assert.NoError(t, err)
	assert.Equal(t, 0.92, rate)
}

func TestGetRateAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid currency pair"}`))
		response := RateResponse{
			ErrorMsg: "invalid currency pair",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("INVALID", "PAIR")

	assert.Error(t, err)
	assert.Equal(t, float64(0), rate)
	assert.Equal(t, "api error: invalid currency pair", err.Error())
}

func TestGetRateMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")

	assert.Error(t, err)
	assert.Equal(t, float64(0), rate)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRateNetworkError(t *testing.T) {
	service := NewExchangeService("http://invalid-domain-that-does-not-exist.local")

	rate, err := service.GetRate("USD", "EUR")

	assert.Error(t, err)
	assert.Equal(t, float64(0), rate)
	assert.Contains(t, err.Error(), "network error")
}

func TestGetRateServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")

	assert.Error(t, err)
	assert.Equal(t, float64(0), rate)
	assert.Contains(t, err.Error(), "unexpected status: 500")
}

func TestGetRateEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(""))
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")

	assert.Error(t, err)
	assert.Equal(t, float64(0), rate)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRateTableDrivenIntegration(t *testing.T) {
	tests := []struct {
		name          string
		from, to      string
		responseBody  interface{}
		statusCode    int
		expectedRate  float64
		expectedError bool
	}{
		{
			name:          "successful USD to EUR",
			from:          "USD",
			to:            "EUR",
			responseBody:  RateResponse{Base: "USD", Target: "EUR", Rate: 0.92},
			statusCode:    http.StatusOK,
			expectedRate:  0.92,
			expectedError: false,
		},
		{
			name:          "successful GBP to JPY",
			from:          "GBP",
			to:            "JPY",
			responseBody:  RateResponse{Base: "GBP", Target: "JPY", Rate: 151.50},
			statusCode:    http.StatusOK,
			expectedRate:  151.50,
			expectedError: false,
		},
		{
			name:          "invalid currency pair",
			from:          "INVALID",
			to:            "PAIR",
			responseBody:  RateResponse{ErrorMsg: "invalid currency pair"},
			statusCode:    http.StatusBadRequest,
			expectedRate:  0,
			expectedError: true,
		},
		{
			name:          "unsupported currency",
			from:          "USD",
			to:            "XYZ",
			responseBody:  RateResponse{ErrorMsg: "unsupported currency"},
			statusCode:    http.StatusNotFound,
			expectedRate:  0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			service := NewExchangeService(server.URL)
			rate, err := service.GetRate(tt.from, tt.to)

			if tt.expectedError {
				assert.Error(t, err, fmt.Sprintf("Expected error for %s -> %s", tt.from, tt.to))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRate, rate)
			}
		})
	}
}
