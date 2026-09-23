package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListHubClientInfo_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hubclientinfo" || r.URL.RawQuery != "offset=0&limit=50" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"party_id":     "TST",
					"country_code": "CL",
					"role":         "CPO",
					"status":       "CONNECTED",
					"last_updated": "2026-09-23T00:00:00Z",
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.ListHubClientInfo(context.Background(), "TOKEN_B_1", 0, 50)
	if err != nil {
		t.Fatalf("ListHubClientInfo returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Status != "CONNECTED" {
		t.Fatalf("unexpected hub client info data: %+v", resp.Data)
	}
}

func TestGetHubClientInfo_ByCountryCodePartyId(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hubclientinfo/CL/TST" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"party_id":     "TST",
					"country_code": "CL",
					"role":         "EMSP",
					"status":       "SUSPENDED",
					"last_updated": "2026-09-23T00:00:00Z",
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetHubClientInfo(context.Background(), "TOKEN_B_1", "CL", "TST")
	if err != nil {
		t.Fatalf("GetHubClientInfo returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Role != "EMSP" {
		t.Fatalf("unexpected hub client info data: %+v", resp.Data)
	}
}

func TestListHubClientInfo_UnknownToken(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           []any{},
			"status_code":    StatusUnknownToken,
			"status_message": "TOKEN_B inválido o conexión no activa.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.ListHubClientInfo(context.Background(), "TOKEN_B_INVALID", 0, 50)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.StatusCode != StatusUnknownToken {
		t.Fatalf("expected status_code %d, got %d", StatusUnknownToken, ocpiErr.StatusCode)
	}
}
