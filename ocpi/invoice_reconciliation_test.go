package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPutInvoiceReconciliation_WithDiscrepancyAmount_ReturnsFXFields(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/invoicereconciliations/CL/TST/REC-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body InvoiceReconciliationInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.CdrID != "CDR-1" || body.DiscrepancyAmount == nil || body.DiscrepancyAmount.ExclVat != 1000 {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code":           "CL",
				"party_id":               "TST",
				"id":                     "REC-1",
				"cdr_id":                 "CDR-1",
				"status":                 "DISPUTED",
				"discrepancy_amount":     map[string]any{"excl_vat": 1000},
				"discrepancy_currency":   "CLP",
				"discrepancy_amount_usd": map[string]any{"excl_vat": 1.05},
				"exchange_rate_used":     952.4,
				"last_updated":           "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutInvoiceReconciliation(context.Background(), "TOKEN_B_1", "CL", "TST", "REC-1", InvoiceReconciliationInput{
		ID:                "REC-1",
		CdrID:             "CDR-1",
		Status:            InvoiceReconciliationDisputed,
		DiscrepancyAmount: &DiscrepancyAmount{ExclVat: 1000},
	})
	if err != nil {
		t.Fatalf("PutInvoiceReconciliation returned error: %v", err)
	}
	if resp.Data.DiscrepancyCurrency != "CLP" {
		t.Fatalf("expected discrepancy_currency CLP, got %q", resp.Data.DiscrepancyCurrency)
	}
	if resp.Data.DiscrepancyAmountUSD == nil || resp.Data.DiscrepancyAmountUSD.ExclVat != 1.05 {
		t.Fatalf("expected discrepancy_amount_usd.excl_vat 1.05, got %+v", resp.Data.DiscrepancyAmountUSD)
	}
	if resp.Data.ExchangeRateUsed == nil || *resp.Data.ExchangeRateUsed != 952.4 {
		t.Fatalf("expected exchange_rate_used 952.4, got %v", resp.Data.ExchangeRateUsed)
	}
}

func TestPutInvoiceReconciliation_WithoutDiscrepancyAmount_FXFieldsAreAbsent(t *testing.T) {
	// Per spec: when the body does NOT include discrepancy_amount, the 3 FX
	// fields (discrepancy_currency, discrepancy_amount_usd,
	// exchange_rate_used) are absent from the response — must decode to
	// their Go zero values (empty string / nil pointer), not be assumed
	// always present.
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "CL",
				"party_id":     "TST",
				"id":           "REC-2",
				"cdr_id":       "CDR-2",
				"status":       "MATCHED",
				"last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutInvoiceReconciliation(context.Background(), "TOKEN_B_1", "CL", "TST", "REC-2", InvoiceReconciliationInput{
		ID:     "REC-2",
		CdrID:  "CDR-2",
		Status: InvoiceReconciliationMatched,
	})
	if err != nil {
		t.Fatalf("PutInvoiceReconciliation returned error: %v", err)
	}
	if resp.Data.DiscrepancyCurrency != "" {
		t.Fatalf("expected discrepancy_currency to be absent, got %q", resp.Data.DiscrepancyCurrency)
	}
	if resp.Data.DiscrepancyAmountUSD != nil {
		t.Fatalf("expected discrepancy_amount_usd to be absent, got %+v", resp.Data.DiscrepancyAmountUSD)
	}
	if resp.Data.ExchangeRateUsed != nil {
		t.Fatalf("expected exchange_rate_used to be absent, got %v", resp.Data.ExchangeRateUsed)
	}
}

func TestPutInvoiceReconciliation_UnsupportedCurrency_PropagatesError(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Moneda del CDR sin proveedor FX soportado.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.PutInvoiceReconciliation(context.Background(), "TOKEN_B_1", "CL", "TST", "REC-3", InvoiceReconciliationInput{
		ID:                "REC-3",
		CdrID:             "CDR-3",
		Status:            InvoiceReconciliationPending,
		DiscrepancyAmount: &DiscrepancyAmount{ExclVat: 10},
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("expected http status %d, got %d", http.StatusBadRequest, ocpiErr.HTTPStatus)
	}
}

func TestDeleteInvoiceReconciliation_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/invoicereconciliations/CL/TST/REC-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	if err := client.DeleteInvoiceReconciliation(context.Background(), "TOKEN_B_1", "CL", "TST", "REC-1"); err != nil {
		t.Fatalf("DeleteInvoiceReconciliation returned error: %v", err)
	}
}

func TestGetInvoiceReconciliations_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoicereconciliations" || r.URL.RawQuery != "offset=0&limit=50" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"country_code": "CL", "party_id": "TST", "id": "REC-1", "cdr_id": "CDR-1",
					"status": "MATCHED", "last_updated": "2026-09-23T00:00:00Z",
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetInvoiceReconciliations(context.Background(), "TOKEN_B_1", 0, 50)
	if err != nil {
		t.Fatalf("GetInvoiceReconciliations returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "REC-1" {
		t.Fatalf("unexpected reconciliations data: %+v", resp.Data)
	}
}
