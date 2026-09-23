package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSetChargingProfile_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chargingprofiles/CL/CM1/SES-1/PUT_CHARGING_PROFILE" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body SetChargingProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ResponseURL != "https://emsp.example.com/callback" || body.ChargingProfile.ChargingRateUnit != "W" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{"result": "ACCEPTED", "timeout": 30},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.SetChargingProfile(context.Background(), "TOKEN_B_1", "CL", "CM1", "SES-1", "https://emsp.example.com/callback", ChargingProfile{
		ChargingRateUnit:      "W",
		ChargingProfilePeriod: []ChargingProfilePeriod{{StartPeriod: 0, Limit: 7400}},
	})
	if err != nil {
		t.Fatalf("SetChargingProfile returned error: %v", err)
	}
	if resp.Data.Result != ChargingProfileAckAccepted || resp.Data.Timeout != 30 {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestGetActiveChargingProfile_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chargingprofiles/CL/CM1/SES-1/GET_ACTIVE_CHARGING_PROFILE" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{"result": "ACCEPTED", "timeout": 30},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetActiveChargingProfile(context.Background(), "TOKEN_B_1", "CL", "CM1", "SES-1", "https://emsp.example.com/callback")
	if err != nil {
		t.Fatalf("GetActiveChargingProfile returned error: %v", err)
	}
	if resp.Data.Result != ChargingProfileAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestDeleteChargingProfile_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chargingprofiles/CL/CM1/SES-1/DELETE_CHARGING_PROFILE" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{"result": "ACCEPTED", "timeout": 30},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.DeleteChargingProfile(context.Background(), "TOKEN_B_1", "CL", "CM1", "SES-1", "https://emsp.example.com/callback")
	if err != nil {
		t.Fatalf("DeleteChargingProfile returned error: %v", err)
	}
	if resp.Data.Result != ChargingProfileAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestGetChargingProfile_UsesCallbackRoute(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/chargingprofiles/callback/CP-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id": "CP-1", "session_id": "SES-1", "action": "PUT_CHARGING_PROFILE",
				"ack_result": "ACCEPTED", "final_result": "ACCEPTED",
				"last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetChargingProfile(context.Background(), "TOKEN_B_1", "CP-1")
	if err != nil {
		t.Fatalf("GetChargingProfile returned error: %v", err)
	}
	if resp.Data.ID != "CP-1" || resp.Data.FinalResult != ChargingProfileFinalAccepted {
		t.Fatalf("unexpected charging profile data: %+v", resp.Data)
	}
}

func TestSetChargingProfile_ErrorPropagates(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "No se encontró la sesión referenciada.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.SetChargingProfile(context.Background(), "TOKEN_B_1", "CL", "CM1", "SES-1", "https://emsp.example.com/callback", ChargingProfile{
		ChargingRateUnit:      "W",
		ChargingProfilePeriod: []ChargingProfilePeriod{{StartPeriod: 0, Limit: 7400}},
	})
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
