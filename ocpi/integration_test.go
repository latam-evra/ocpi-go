//go:build integration

package ocpi

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration suite exercising Credentials, Locations and Tariffs against a
// real, locally running Hub — no mocks. Mirrors
// sdks/node/tests/integration/locations-tariffs.integration.test.ts and
// sdks/python/tests/integration/test_locations_tariffs_integration.py.
//
// Requires OCPI_ALLOW_LOOPBACK=true on the Hub instance under test (its
// Credentials handshake simulates a CSMS running on localhost, which the
// Hub's SSRF protection blocks by default). Never point this at the
// production pm2 process — see sdks/node/README.md for the full rationale.
// Run with:
//
//	OCPI_HUB_TEST_URL=http://localhost:3948 go test -tags=integration ./...

func hubTestURL() string {
	if url := os.Getenv("OCPI_HUB_TEST_URL"); url != "" {
		return url
	}
	return "http://localhost:3947"
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// sdks/go/ocpi -> repo root is two levels up.
	return filepath.Join(wd, "..", "..", "..")
}

type testRegistration struct {
	RegistrationID string `json:"registrationId"`
	RawTokenA      string `json:"rawTokenA"`
}

func createTestRegistration(t *testing.T, role, countryCode, partyID string) testRegistration {
	t.Helper()
	cmd := exec.Command("npx", "tsx", "scripts/create-test-registration.ts",
		"--role", role, "--country-code", countryCode, "--party-id", partyID)
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("create-test-registration.ts failed: %v\nstderr: %s", err, exitErr.Stderr)
		}
		t.Fatalf("create-test-registration.ts failed: %v", err)
	}

	var reg testRegistration
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &reg); err != nil {
		t.Fatalf("failed to parse create-test-registration.ts output: %v\noutput: %s", err, out)
	}
	return reg
}

func cleanupTestRegistration(t *testing.T, registrationID string) {
	t.Helper()
	cmd := exec.Command("npx", "tsx", "scripts/create-test-registration.ts", "--cleanup", registrationID)
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("cleanup of registration %s failed: %v\noutput: %s", registrationID, err, out)
	}
}

func TestIntegration_LocationsAndTariffs_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	reg := createTestRegistration(t, "CPO", "CL", "GO1")
	defer cleanupTestRegistration(t, reg.RegistrationID)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	credsResp, err := client.RegisterCredentials(ctx, reg.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "GO1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials failed: %v", err)
	}
	tokenB := credsResp.Data.Token

	created, err := client.PutLocation(ctx, tokenB, "CL", "GO1", "LOC-GO-1", LocationInput{
		ID:          "LOC-GO-1",
		Publish:     true,
		Address:     "Av. Go 789",
		City:        "Santiago",
		Country:     "CHL",
		Coordinates: Coordinates{Latitude: "-33.4", Longitude: "-70.6"},
	})
	if err != nil {
		t.Fatalf("PutLocation failed: %v", err)
	}
	if created.Data.ID != "LOC-GO-1" {
		t.Fatalf("expected location id LOC-GO-1, got %q", created.Data.ID)
	}

	fetched, err := client.GetLocation(ctx, tokenB, "CL", "GO1", "LOC-GO-1")
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}
	if fetched.Data.Address != "Av. Go 789" {
		t.Fatalf("expected address %q, got %q", "Av. Go 789", fetched.Data.Address)
	}

	patched, err := client.PatchLocation(ctx, tokenB, "CL", "GO1", "LOC-GO-1", map[string]any{
		"city": "Valparaíso",
	})
	if err != nil {
		t.Fatalf("PatchLocation failed: %v", err)
	}
	if patched.Data.City != "Valparaíso" {
		t.Fatalf("expected city %q, got %q", "Valparaíso", patched.Data.City)
	}

	page, err := client.GetLocations(ctx, tokenB, 0, 100)
	if err != nil {
		t.Fatalf("GetLocations failed: %v", err)
	}
	found := false
	for _, loc := range page.Data {
		if loc.ID == "LOC-GO-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected LOC-GO-1 in locations list, got %+v", page.Data)
	}

	_, err = client.PutTariff(ctx, tokenB, "CL", "GO1", "TAR-GO-1", TariffInput{
		ID:       "TAR-GO-1",
		Currency: "USD",
		Elements: []TariffElement{
			{PriceComponents: []PriceComponent{{Type: "ENERGY", Price: 0.4, StepSize: 1}}},
		},
	})
	if err != nil {
		t.Fatalf("PutTariff failed: %v", err)
	}

	tariff, err := client.GetTariff(ctx, tokenB, "CL", "GO1", "TAR-GO-1")
	if err != nil {
		t.Fatalf("GetTariff failed: %v", err)
	}
	if tariff.Data.Currency != "USD" {
		t.Fatalf("expected currency USD, got %q", tariff.Data.Currency)
	}

	if err := client.DeleteTariff(ctx, tokenB, "CL", "GO1", "TAR-GO-1"); err != nil {
		t.Fatalf("DeleteTariff failed: %v", err)
	}

	if _, err := client.GetTariff(ctx, tokenB, "CL", "GO1", "TAR-GO-1"); err == nil {
		t.Fatal("expected GetTariff to fail after delete, got nil error")
	} else {
		var ocpiErr *OcpiError
		if !AsOcpiError(err, &ocpiErr) {
			t.Fatalf("expected *OcpiError after delete, got %T: %v", err, err)
		}
	}
}
