package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleTransactions_ValidationErrors(t *testing.T) {
	srv := &Server{DB: nil}

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "invalid json body",
			body:           `{"idempotency_key": "tx-1", invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid JSON request",
		},
		{
			name:           "missing idempotency key",
			body:           `{"description": "no key", "postings": [{"account_id": "a1", "amount": 100}, {"account_id": "a2", "amount": -100}]}`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "idempotency_key is required",
		},
		{
			name:           "empty postings",
			body:           `{"idempotency_key": "tx-1", "postings": []}`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "at least two postings are required",
		},
		{
			name:           "posting with zero amount",
			body:           `{"idempotency_key": "tx-1", "postings": [{"account_id": "a1", "amount": 0}, {"account_id": "a2", "amount": 100}]}`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "each posting needs an account_id and non-zero amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.HandleTransactions(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), tt.expectedMsg) {
				t.Fatalf("expected error body to contain %q, got %q", tt.expectedMsg, rec.Body.String())
			}
		})
	}
}
