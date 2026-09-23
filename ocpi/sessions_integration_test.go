//go:build integration

package ocpi

import (
	"context"
	"testing"
)

// Mirrors sdks/node/tests/integration/sessions.integration.test.ts and
// sdks/python/tests/integration/test_sessions_integration.py. Reuses
// hubTestURL/createTestRegistration/cleanupTestRegistration from
// integration_test.go.

func TestIntegration_Sessions_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	reg := createTestRegistration(t, "CPO", "CL", "SI1")
	defer cleanupTestRegistration(t, reg.RegistrationID)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	credsResp, err := client.RegisterCredentials(ctx, reg.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "SI1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials failed: %v", err)
	}
	tokenB := credsResp.Data.Token

	created, err := client.PutSession(ctx, tokenB, "CL", "SI1", "SES-SDK-1", SessionInput{
		ID:            "SES-SDK-1",
		StartDateTime: "2026-09-23T09:00:00Z",
		KWh:           5.5,
		CdrToken: CdrToken{
			CountryCode: "AR", PartyID: "EVP", UID: "TOK-SDK-1", Type: TokenTypeRFID, ContractID: "C-SDK-1",
		},
		AuthMethod:  AuthMethodWhitelist,
		LocationID:  "LOC-SDK-1",
		EvseUID:     "EVSE-1",
		ConnectorID: "1",
		Currency:    "USD",
		Status:      SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("PutSession failed: %v", err)
	}
	if created.Data.ID != "SES-SDK-1" || created.Data.CountryCode != "CL" || created.Data.Status != SessionStatusActive {
		t.Fatalf("unexpected created session: %+v", created.Data)
	}

	fetched, err := client.GetSession(ctx, tokenB, "CL", "SI1", "SES-SDK-1")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if fetched.Data.KWh != 5.5 {
		t.Fatalf("expected kwh 5.5, got %v", fetched.Data.KWh)
	}

	patched, err := client.PatchSession(ctx, tokenB, "CL", "SI1", "SES-SDK-1", map[string]any{
		"kwh":    9.2,
		"status": "COMPLETED",
	})
	if err != nil {
		t.Fatalf("PatchSession failed: %v", err)
	}
	if patched.Data.KWh != 9.2 || patched.Data.Status != SessionStatusCompleted {
		t.Fatalf("unexpected patched session: %+v", patched.Data)
	}

	list, err := client.GetSessions(ctx, tokenB, 0, 100)
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}
	found := false
	for _, s := range list.Data {
		if s.ID == "SES-SDK-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected SES-SDK-1 in sessions list, got %+v", list.Data)
	}

	if _, err := client.PatchSession(ctx, tokenB, "CL", "SI1", "DOES-NOT-EXIST", map[string]any{"kwh": 1}); err == nil {
		t.Fatal("expected PatchSession on a non-existent session to fail with 404")
	} else {
		var ocpiErr *OcpiError
		if !AsOcpiError(err, &ocpiErr) {
			t.Fatalf("expected *OcpiError, got %T: %v", err, err)
		}
	}
}
