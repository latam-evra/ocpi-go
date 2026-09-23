package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestStartSession_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commands/START_SESSION" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body StartSessionCommand
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ResponseURL != "https://emsp.example.com/callback" || body.LocationID != "LOC-1" {
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

	resp, err := client.StartSession(context.Background(), "TOKEN_B_1", StartSessionCommand{
		ResponseURL: "https://emsp.example.com/callback",
		CountryCode: "CL",
		PartyID:     "CM1",
		Token:       CommandTokenRef{UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1"},
		LocationID:  "LOC-1",
	})
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	if resp.Data.Result != CommandAckAccepted || resp.Data.Timeout != 30 {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestReserveNow_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commands/RESERVE_NOW" {
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

	resp, err := client.ReserveNow(context.Background(), "TOKEN_B_1", ReserveNowCommand{
		ResponseURL:   "https://emsp.example.com/callback",
		CountryCode:   "CL",
		PartyID:       "CM1",
		Token:         CommandTokenRef{UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1"},
		ExpiryDate:    "2026-09-24T00:00:00Z",
		ReservationID: "RES-1",
		LocationID:    "LOC-1",
	})
	if err != nil {
		t.Fatalf("ReserveNow returned error: %v", err)
	}
	if resp.Data.Result != CommandAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestStopSession_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commands/STOP_SESSION" {
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

	resp, err := client.StopSession(context.Background(), "TOKEN_B_1", StopSessionCommand{
		ResponseURL: "https://emsp.example.com/callback",
		CountryCode: "CL",
		PartyID:     "CM1",
		SessionID:   "SES-1",
	})
	if err != nil {
		t.Fatalf("StopSession returned error: %v", err)
	}
	if resp.Data.Result != CommandAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestUnlockConnector_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commands/UNLOCK_CONNECTOR" {
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

	resp, err := client.UnlockConnector(context.Background(), "TOKEN_B_1", UnlockConnectorCommand{
		ResponseURL: "https://emsp.example.com/callback",
		CountryCode: "CL",
		PartyID:     "CM1",
		SessionID:   "SES-1",
	})
	if err != nil {
		t.Fatalf("UnlockConnector returned error: %v", err)
	}
	if resp.Data.Result != CommandAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestCancelReservation_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commands/CANCEL_RESERVATION" {
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

	resp, err := client.CancelReservation(context.Background(), "TOKEN_B_1", CancelReservationCommand{
		ResponseURL: "https://emsp.example.com/callback",
		CountryCode: "CL",
		PartyID:     "CM1",
		SessionID:   "SES-1",
	})
	if err != nil {
		t.Fatalf("CancelReservation returned error: %v", err)
	}
	if resp.Data.Result != CommandAckAccepted {
		t.Fatalf("unexpected ack: %+v", resp.Data)
	}
}

func TestGetCommand_UsesCallbackRoute(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/commands/callback/CMD-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id": "CMD-1", "type": "START_SESSION", "payload": map[string]any{},
				"ack_result": "ACCEPTED", "final_result": "ACCEPTED",
				"last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetCommand(context.Background(), "TOKEN_B_1", "CMD-1")
	if err != nil {
		t.Fatalf("GetCommand returned error: %v", err)
	}
	if resp.Data.ID != "CMD-1" || resp.Data.FinalResult != CommandFinalAccepted {
		t.Fatalf("unexpected command data: %+v", resp.Data)
	}
}

func TestStartSession_ErrorPropagates(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "El CPO destino no está conectado.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.StartSession(context.Background(), "TOKEN_B_1", StartSessionCommand{
		ResponseURL: "https://emsp.example.com/callback",
		CountryCode: "CL",
		PartyID:     "CM1",
		Token:       CommandTokenRef{UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1"},
		LocationID:  "LOC-1",
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.HTTPStatus != http.StatusUnprocessableEntity {
		t.Fatalf("expected http status %d, got %d", http.StatusUnprocessableEntity, ocpiErr.HTTPStatus)
	}
}
