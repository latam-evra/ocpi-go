//go:build integration

package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Mirrors sdks/node/tests/integration/chargingProfiles.integration.test.ts:
// a real HTTP server simulating the target CSMS/CPO, exposing /versions,
// /details (advertising a "chargingprofiles" endpoint) and
// GET|PUT|DELETE /chargingprofiles/{session_id}, responding with a valid
// ACK.

type mockChargingProfileCpoServer struct {
	server     *httptest.Server
	mu         sync.Mutex
	lastMethod string
}

func (m *mockChargingProfileCpoServer) LastMethod() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMethod
}

func startMockChargingProfileCpoServer(t *testing.T) *mockChargingProfileCpoServer {
	t.Helper()
	m := &mockChargingProfileCpoServer{}

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
				"endpoints": []map[string]string{{"identifier": "chargingprofiles", "role": "CPO", "url": m.server.URL + "/cpo/chargingprofiles"}},
			},
			"status_code": StatusSuccess,
		})
	})
	mux.HandleFunc("/cpo/chargingprofiles/", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.lastMethod = r.Method
		m.mu.Unlock()
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

func TestIntegration_ChargingProfiles_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	mock := startMockChargingProfileCpoServer(t)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	cpo := createTestRegistration(t, "CPO", "CL", "CQ1")
	defer cleanupTestRegistration(t, cpo.RegistrationID)
	cpoCreds, err := client.RegisterCredentials(ctx, cpo.RawTokenA, mock.server.URL+"/cpo/versions", []Role{
		{Role: RoleCPO, PartyID: "CQ1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (CPO) failed: %v", err)
	}
	cpoTokenB := cpoCreds.Data.Token

	if _, err := client.PutSession(ctx, cpoTokenB, "CL", "CQ1", "SES-CP-1", SessionInput{
		ID:            "SES-CP-1",
		StartDateTime: "2026-09-23T09:00:00Z",
		KWh:           0,
		CdrToken: CdrToken{
			CountryCode: "AR", PartyID: "EMQ", UID: "TOK-CP-1", Type: TokenTypeRFID, ContractID: "C-CP-1",
		},
		AuthMethod:  AuthMethodAuthRequest,
		LocationID:  "LOC-CP-1",
		EvseUID:     "EVSE-1",
		ConnectorID: "1",
		Currency:    "USD",
		Status:      SessionStatusActive,
	}); err != nil {
		t.Fatalf("PutSession (fixture) failed: %v", err)
	}

	emsp := createTestRegistration(t, "EMSP", "AR", "EMQ")
	defer cleanupTestRegistration(t, emsp.RegistrationID)
	emspCreds, err := client.RegisterCredentials(ctx, emsp.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleEMSP, PartyID: "EMQ", CountryCode: "AR"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (EMSP) failed: %v", err)
	}
	emspTokenB := emspCreds.Data.Token

	t.Run("SetChargingProfile resolves the target CPO via session_id, forwards a PUT and returns the ACK", func(t *testing.T) {
		ack, err := client.SetChargingProfile(ctx, emspTokenB, "CL", "CQ1", "SES-CP-1", "https://emsp.example.com/callback", ChargingProfile{
			ChargingRateUnit:      "W",
			ChargingProfilePeriod: []ChargingProfilePeriod{{StartPeriod: 0, Limit: 7400}},
		})
		if err != nil {
			t.Fatalf("SetChargingProfile failed: %v", err)
		}
		if ack.Data.Result != ChargingProfileAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
		if mock.LastMethod() != http.MethodPut {
			t.Fatalf("expected PUT, got %s", mock.LastMethod())
		}
	})

	t.Run("GetActiveChargingProfile forwards a GET and returns the ACK", func(t *testing.T) {
		ack, err := client.GetActiveChargingProfile(ctx, emspTokenB, "CL", "CQ1", "SES-CP-1", "https://emsp.example.com/callback")
		if err != nil {
			t.Fatalf("GetActiveChargingProfile failed: %v", err)
		}
		if ack.Data.Result != ChargingProfileAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
		if mock.LastMethod() != http.MethodGet {
			t.Fatalf("expected GET, got %s", mock.LastMethod())
		}
	})

	t.Run("DeleteChargingProfile forwards a DELETE and returns the ACK", func(t *testing.T) {
		ack, err := client.DeleteChargingProfile(ctx, emspTokenB, "CL", "CQ1", "SES-CP-1", "https://emsp.example.com/callback")
		if err != nil {
			t.Fatalf("DeleteChargingProfile failed: %v", err)
		}
		if ack.Data.Result != ChargingProfileAckAccepted {
			t.Fatalf("expected ACCEPTED, got %+v", ack.Data)
		}
		if mock.LastMethod() != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", mock.LastMethod())
		}
	})
}
