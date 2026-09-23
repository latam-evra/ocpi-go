//go:build integration

package ocpi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
)

// Mirrors sdks/node/tests/integration/commands.integration.test.ts and
// sdks/python/tests/integration/test_commands_integration.py: a real HTTP
// server simulating the target CSMS/CPO, exposing /versions, /details
// (advertising a "commands" endpoint) and /commands/{TYPE}, responding
// with a valid ACK.

var commandIDPattern = regexp.MustCompile(`/commands/callback/([^/]+)$`)

type mockCpoServer struct {
	server *httptest.Server
	mu     sync.Mutex
	lastID string
}

func (m *mockCpoServer) LastCommandID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastID
}

func startMockCpoServer(t *testing.T) *mockCpoServer {
	t.Helper()
	m := &mockCpoServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/cpo/versions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":        []map[string]string{{"version": "2.3.0", "url": m.server.URL + "/cpo/details"}},
			"status_code": StatusSuccess,
		})
	})
	mux.HandleFunc("/cpo/details", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"version":   "2.3.0",
				"endpoints": []map[string]string{{"identifier": "commands", "role": "CPO", "url": m.server.URL + "/cpo/commands"}},
			},
			"status_code": StatusSuccess,
		})
	})
	mux.HandleFunc("/cpo/commands/", func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body struct {
			ResponseURL string `json:"response_url"`
		}
		if err := json.Unmarshal(raw, &body); err == nil && body.ResponseURL != "" {
			if match := commandIDPattern.FindStringSubmatch(body.ResponseURL); match != nil {
				m.mu.Lock()
				m.lastID = match[1]
				m.mu.Unlock()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":        map[string]any{"result": "ACCEPTED", "timeout": 30},
			"status_code": StatusSuccess,
		})
	})

	m.server = httptest.NewServer(mux)
	t.Cleanup(m.server.Close)
	return m
}

func TestIntegration_Commands_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	mock := startMockCpoServer(t)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	cpo := createTestRegistration(t, "CPO", "CL", "CM1")
	defer cleanupTestRegistration(t, cpo.RegistrationID)
	cpoCreds, err := client.RegisterCredentials(ctx, cpo.RawTokenA, mock.server.URL+"/cpo/versions", []Role{
		{Role: RoleCPO, PartyID: "CM1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (CPO) failed: %v", err)
	}
	cpoTokenB := cpoCreds.Data.Token

	// Publish the location and session the commands will reference, using
	// the SDK client itself (CPO role).
	if _, err := client.PutLocation(ctx, cpoTokenB, "CL", "CM1", "LOC-CMD-1", LocationInput{
		ID: "LOC-CMD-1", Publish: true, Address: "x", City: "x", Country: "CHL",
		Coordinates: Coordinates{Latitude: "0", Longitude: "0"},
	}); err != nil {
		t.Fatalf("PutLocation (fixture) failed: %v", err)
	}
	if _, err := client.PutSession(ctx, cpoTokenB, "CL", "CM1", "SES-CMD-1", SessionInput{
		ID:            "SES-CMD-1",
		StartDateTime: "2026-09-23T09:00:00Z",
		KWh:           0,
		CdrToken: CdrToken{
			CountryCode: "AR", PartyID: "EMS", UID: "TOK-CMD-1", Type: TokenTypeRFID, ContractID: "C-CMD-1",
		},
		AuthMethod:  AuthMethodAuthRequest,
		LocationID:  "LOC-CMD-1",
		EvseUID:     "EVSE-1",
		ConnectorID: "1",
		Currency:    "USD",
		Status:      SessionStatusActive,
	}); err != nil {
		t.Fatalf("PutSession (fixture) failed: %v", err)
	}

	emsp := createTestRegistration(t, "EMSP", "AR", "EMS")
	defer cleanupTestRegistration(t, emsp.RegistrationID)
	emspCreds, err := client.RegisterCredentials(ctx, emsp.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleEMSP, PartyID: "EMS", CountryCode: "AR"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (EMSP) failed: %v", err)
	}
	emspTokenB := emspCreds.Data.Token

	t.Run("StartSession resolves the target CPO via location_id and returns the mock CPO's ACK", func(t *testing.T) {
		ack, err := client.StartSession(ctx, emspTokenB, StartSessionCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1",
			Token:      CommandTokenRef{UID: "TOK-CMD-1", Type: TokenTypeRFID, ContractID: "C-CMD-1"},
			LocationID: "LOC-CMD-1",
		})
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
	})

	t.Run("StopSession resolves the target CPO via session_id and returns the ACK", func(t *testing.T) {
		ack, err := client.StopSession(ctx, emspTokenB, StopSessionCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1", SessionID: "SES-CMD-1",
		})
		if err != nil {
			t.Fatalf("StopSession failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
	})

	t.Run("UnlockConnector resolves the target CPO via session_id and returns the ACK", func(t *testing.T) {
		ack, err := client.UnlockConnector(ctx, emspTokenB, UnlockConnectorCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1", SessionID: "SES-CMD-1",
		})
		if err != nil {
			t.Fatalf("UnlockConnector failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
	})

	t.Run("CancelReservation resolves the target CPO via session_id and returns the ACK", func(t *testing.T) {
		ack, err := client.CancelReservation(ctx, emspTokenB, CancelReservationCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1", SessionID: "SES-CMD-1",
		})
		if err != nil {
			t.Fatalf("CancelReservation failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
	})

	t.Run("ReserveNow resolves the target CPO via location_id and returns the ACK", func(t *testing.T) {
		ack, err := client.ReserveNow(ctx, emspTokenB, ReserveNowCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1",
			Token:         CommandTokenRef{UID: "TOK-CMD-1", Type: TokenTypeRFID, ContractID: "C-CMD-1"},
			ExpiryDate:    "2026-09-24T00:00:00Z",
			ReservationID: "RES-CMD-1",
			LocationID:    "LOC-CMD-1",
		})
		if err != nil {
			t.Fatalf("ReserveNow failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
	})

	t.Run("GetCommand fetches the command created by StartSession via GET /commands/callback/{id}", func(t *testing.T) {
		ack, err := client.StartSession(ctx, emspTokenB, StartSessionCommand{
			ResponseURL: "https://emsp.example.com/callback",
			CountryCode: "CL", PartyID: "CM1",
			Token:      CommandTokenRef{UID: "TOK-CMD-1", Type: TokenTypeRFID, ContractID: "C-CMD-1"},
			LocationID: "LOC-CMD-1",
		})
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
		if ack.Data.Result != CommandAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}

		// The ACK returned by StartSession is the target CPO's, and does not
		// carry the command id — the mock CPO captures it by reading the
		// response_url the Hub rewrites internally before forwarding the
		// POST.
		commandID := mock.LastCommandID()
		if commandID == "" {
			t.Fatal("expected the mock CPO to have captured a command id")
		}

		command, err := client.GetCommand(ctx, emspTokenB, commandID)
		if err != nil {
			t.Fatalf("GetCommand failed: %v", err)
		}
		if command.Data.ID != commandID {
			t.Fatalf("expected command id %q, got %q", commandID, command.Data.ID)
		}
		if command.Data.Type != "START_SESSION" {
			t.Fatalf("expected type START_SESSION, got %q", command.Data.Type)
		}
		if command.Data.AckResult != CommandAckAccepted {
			t.Fatalf("expected ack_result ACCEPTED, got %q", command.Data.AckResult)
		}
	})
}
