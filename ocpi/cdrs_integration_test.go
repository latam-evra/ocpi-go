//go:build integration

package ocpi

import (
	"context"
	"testing"
)

// Mirrors sdks/node/tests/integration/cdrs.integration.test.ts and
// sdks/python/tests/integration/test_cdrs_integration.py.

func TestIntegration_Cdrs_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	reg := createTestRegistration(t, "CPO", "CL", "CI1")
	defer cleanupTestRegistration(t, reg.RegistrationID)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	credsResp, err := client.RegisterCredentials(ctx, reg.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "CI1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials failed: %v", err)
	}
	tokenB := credsResp.Data.Token

	cdrInput := CdrInput{
		ID:            "CDR-SDK-1",
		StartDateTime: "2026-09-23T09:00:00Z",
		EndDateTime:   "2026-09-23T10:00:00Z",
		CdrToken: CdrToken{
			CountryCode: "AR", PartyID: "EVP", UID: "TOK-SDK-1", Type: TokenTypeRFID, ContractID: "C-SDK-1",
		},
		AuthMethod: AuthMethodWhitelist,
		CdrLocation: CdrLocation{
			ID:          "LOC-SDK-1",
			Address:     "Av. Test 123",
			City:        "Santiago",
			Country:     "CHL",
			Coordinates: Coordinates{Latitude: "-33.4", Longitude: "-70.6"},
			EVSEUID:     "EVSE-1", ConnectorID: "1", ConnectorStandard: "IEC_62196_T2",
			ConnectorFormat: "SOCKET", ConnectorPowerType: "AC_3_PHASE",
		},
		Currency: "USD",
		ChargingPeriods: []ChargingPeriod{
			{StartDateTime: "2026-09-23T09:00:00Z", Dimensions: []ChargingPeriodDimension{{Type: "ENERGY", Volume: 5.5}}},
		},
		TotalCost:   CostAmount{ExclVat: 2.5},
		TotalEnergy: 5.5,
		TotalTime:   1,
	}

	created, err := client.PostCdr(ctx, tokenB, "CL", "CI1", "CDR-SDK-1", cdrInput)
	if err != nil {
		t.Fatalf("PostCdr failed: %v", err)
	}
	if created.Data.ID != "CDR-SDK-1" || created.Data.CountryCode != "CL" || created.Data.TotalEnergy != 5.5 {
		t.Fatalf("unexpected created cdr: %+v", created.Data)
	}

	fetched, err := client.GetCdr(ctx, tokenB, "CL", "CI1", "CDR-SDK-1")
	if err != nil {
		t.Fatalf("GetCdr failed: %v", err)
	}
	if fetched.Data.Currency != "USD" {
		t.Fatalf("expected currency USD, got %q", fetched.Data.Currency)
	}

	list, err := client.GetCdrs(ctx, tokenB, 0, 100)
	if err != nil {
		t.Fatalf("GetCdrs failed: %v", err)
	}
	found := false
	for _, cdr := range list.Data {
		if cdr.ID == "CDR-SDK-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected CDR-SDK-1 in cdrs list, got %+v", list.Data)
	}

	// Immutability: a second POST with the same id fails with 409.
	if _, err := client.PostCdr(ctx, tokenB, "CL", "CI1", "CDR-SDK-1", cdrInput); err == nil {
		t.Fatal("expected a second PostCdr with the same id to fail")
	} else {
		var ocpiErr *OcpiError
		if !AsOcpiError(err, &ocpiErr) {
			t.Fatalf("expected *OcpiError, got %T: %v", err, err)
		}
	}
}
