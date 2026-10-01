package ledger

import (
	"context"
	"strings"
	"testing"
)

func TestRecordTransaction_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		req         TransactionRequest
		expectedErr string
	}{
		{
			name: "less than two postings",
			req: TransactionRequest{
				IdempotencyKey: "tx-1",
				Description:    "single posting",
				Postings: []Posting{
					{AccountID: "acc-1", Amount: 100},
				},
			},
			expectedErr: "at least two postings",
		},
		{
			name: "zero amount in posting",
			req: TransactionRequest{
				IdempotencyKey: "tx-2",
				Description:    "zero amount",
				Postings: []Posting{
					{AccountID: "acc-1", Amount: 0},
					{AccountID: "acc-2", Amount: 0},
				},
			},
			expectedErr: "cannot be zero",
		},
		{
			name: "empty account ID",
			req: TransactionRequest{
				IdempotencyKey: "tx-3",
				Description:    "empty account id",
				Postings: []Posting{
					{AccountID: "", Amount: 100},
					{AccountID: "acc-2", Amount: -100},
				},
			},
			expectedErr: "cannot be empty",
		},
		{
			name: "unbalanced postings",
			req: TransactionRequest{
				IdempotencyKey: "tx-4",
				Description:    "unbalanced postings",
				Postings: []Posting{
					{AccountID: "acc-1", Amount: 100},
					{AccountID: "acc-2", Amount: -80},
				},
			},
			expectedErr: "unbalanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RecordTransaction(ctx, nil, &tt.req)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}
