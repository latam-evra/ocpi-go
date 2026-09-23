package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPutSession_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/sessions/CL/TST/SES-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body SessionInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ID != "SES-1" || body.CdrToken.UID != "TOK-1" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code":    "CL",
				"party_id":        "TST",
				"id":              "SES-1",
				"start_date_time": "2026-09-23T00:00:00Z",
				"kwh":             10.5,
				"cdr_token": map[string]any{
					"country_code": "AR", "party_id": "EMS", "uid": "TOK-1", "type": "RFID", "contract_id": "C-1",
				},
				"auth_method":  "AUTH_REQUEST",
				"location_id":  "LOC-1",
				"evse_uid":     "EVSE-1",
				"connector_id": "1",
				"currency":     "USD",
				"status":       "ACTIVE",
				"last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutSession(context.Background(), "TOKEN_B_1", "CL", "TST", "SES-1", SessionInput{
		ID:            "SES-1",
		StartDateTime: "2026-09-23T00:00:00Z",
		KWh:           10.5,
		CdrToken:      CdrToken{CountryCode: "AR", PartyID: "EMS", UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1"},
		AuthMethod:    AuthMethodAuthRequest,
		LocationID:    "LOC-1",
		EvseUID:       "EVSE-1",
		ConnectorID:   "1",
		Currency:      "USD",
		Status:        SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("PutSession returned error: %v", err)
	}
	if resp.Data.ID != "SES-1" || resp.Data.CountryCode != "CL" || resp.Data.Status != SessionStatusActive {
		t.Fatalf("unexpected session data: %+v", resp.Data)
	}
}

func TestPatchSession_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/sessions/CL/TST/SES-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["status"] != "COMPLETED" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code":    "CL",
				"party_id":        "TST",
				"id":              "SES-1",
				"start_date_time": "2026-09-23T00:00:00Z",
				"kwh":             10.5,
				"cdr_token": map[string]any{
					"country_code": "AR", "party_id": "EMS", "uid": "TOK-1", "type": "RFID", "contract_id": "C-1",
				},
				"auth_method":  "AUTH_REQUEST",
				"location_id":  "LOC-1",
				"evse_uid":     "EVSE-1",
				"connector_id": "1",
				"currency":     "USD",
				"status":       "COMPLETED",
				"last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PatchSession(context.Background(), "TOKEN_B_1", "CL", "TST", "SES-1", map[string]any{
		"status": "COMPLETED",
	})
	if err != nil {
		t.Fatalf("PatchSession returned error: %v", err)
	}
	if resp.Data.Status != SessionStatusCompleted {
		t.Fatalf("unexpected status: %q", resp.Data.Status)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Session no encontrada.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.GetSession(context.Background(), "TOKEN_B_1", "CL", "TST", "DOES-NOT-EXIST")
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

func TestGetSessions_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions" || r.URL.RawQuery != "offset=0&limit=10" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"country_code": "CL", "party_id": "TST", "id": "SES-1",
					"start_date_time": "2026-09-23T00:00:00Z", "kwh": 1.0,
					"cdr_token": map[string]any{
						"country_code": "AR", "party_id": "EMS", "uid": "TOK-1", "type": "RFID", "contract_id": "C-1",
					},
					"auth_method": "AUTH_REQUEST", "location_id": "LOC-1", "evse_uid": "EVSE-1",
					"connector_id": "1", "currency": "USD", "status": "ACTIVE",
					"last_updated": "2026-09-23T00:00:00Z",
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetSessions(context.Background(), "TOKEN_B_1", 0, 10)
	if err != nil {
		t.Fatalf("GetSessions returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "SES-1" {
		t.Fatalf("unexpected sessions data: %+v", resp.Data)
	}
}
