package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPutTariff_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/tariffs/CL/TST/TAR-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body TariffInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ID != "TAR-1" || body.Currency != "USD" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "CL",
				"party_id":     "TST",
				"id":           "TAR-1",
				"currency":     "USD",
				"elements": []map[string]any{
					{"price_components": []map[string]any{
						{"type": "ENERGY", "price": 0.35, "step_size": 1},
					}},
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutTariff(context.Background(), "TOKEN_B_1", "CL", "TST", "TAR-1", TariffInput{
		ID:       "TAR-1",
		Currency: "USD",
		Elements: []TariffElement{
			{PriceComponents: []PriceComponent{{Type: "ENERGY", Price: 0.35, StepSize: 1}}},
		},
	})
	if err != nil {
		t.Fatalf("PutTariff returned error: %v", err)
	}
	if resp.Data.ID != "TAR-1" || resp.Data.CountryCode != "CL" || resp.Data.PartyID != "TST" {
		t.Fatalf("unexpected tariff data: %+v", resp.Data)
	}
}

func TestDeleteTariff_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/tariffs/CL/TST/TAR-1" {
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

	if err := client.DeleteTariff(context.Background(), "TOKEN_B_1", "CL", "TST", "TAR-1"); err != nil {
		t.Fatalf("DeleteTariff returned error: %v", err)
	}
}

func TestGetTariff_NotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Tariff no encontrado.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.GetTariff(context.Background(), "TOKEN_B_1", "CL", "TST", "DOES-NOT-EXIST")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.HTTPStatus != http.StatusNotFound {
		t.Fatalf("expected http status %d, got %d", http.StatusNotFound, ocpiErr.HTTPStatus)
	}
}
