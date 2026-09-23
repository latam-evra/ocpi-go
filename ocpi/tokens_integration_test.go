//go:build integration

package ocpi

import (
	"context"
	"testing"
)

// Mirrors sdks/node/tests/integration/tokens.integration.test.ts and
// sdks/python/tests/integration/test_tokens_integration.py.

func TestIntegration_Tokens_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	emsp := createTestRegistration(t, "EMSP", "AR", "TI1")
	defer cleanupTestRegistration(t, emsp.RegistrationID)
	emspCreds, err := client.RegisterCredentials(ctx, emsp.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleEMSP, PartyID: "TI1", CountryCode: "AR"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (EMSP) failed: %v", err)
	}
	emspTokenB := emspCreds.Data.Token

	cpo := createTestRegistration(t, "CPO", "CL", "TI2")
	defer cleanupTestRegistration(t, cpo.RegistrationID)
	cpoCreds, err := client.RegisterCredentials(ctx, cpo.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "TI2", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials (CPO) failed: %v", err)
	}
	cpoTokenB := cpoCreds.Data.Token

	created, err := client.PutToken(ctx, emspTokenB, "AR", "TI1", "TOK-SDK-1", TokenInput{
		UID: "TOK-SDK-1", Type: TokenTypeRFID, ContractID: "C-SDK-1",
		Issuer: "LATAM EVP", Valid: true, Whitelist: TokenWhitelistAlways,
	})
	if err != nil {
		t.Fatalf("PutToken failed: %v", err)
	}
	if created.Data.UID != "TOK-SDK-1" || created.Data.CountryCode != "AR" {
		t.Fatalf("unexpected created token: %+v", created.Data)
	}

	fetched, err := client.GetToken(ctx, emspTokenB, "AR", "TI1", "TOK-SDK-1")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if fetched.Data.Issuer != "LATAM EVP" {
		t.Fatalf("expected issuer LATAM EVP, got %q", fetched.Data.Issuer)
	}

	patched, err := client.PatchToken(ctx, emspTokenB, "AR", "TI1", "TOK-SDK-1", map[string]any{"valid": false})
	if err != nil {
		t.Fatalf("PatchToken failed: %v", err)
	}
	if patched.Data.Valid {
		t.Fatalf("expected valid=false after patch, got %+v", patched.Data)
	}

	list, err := client.GetTokens(ctx, emspTokenB, 0, 100)
	if err != nil {
		t.Fatalf("GetTokens failed: %v", err)
	}
	found := false
	for _, tok := range list.Data {
		if tok.UID == "TOK-SDK-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected TOK-SDK-1 in tokens list, got %+v", list.Data)
	}

	if err := client.DeleteToken(ctx, emspTokenB, "AR", "TI1", "TOK-SDK-1"); err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}
	if _, err := client.GetToken(ctx, emspTokenB, "AR", "TI1", "TOK-SDK-1"); err == nil {
		t.Fatal("expected GetToken to fail after delete")
	}

	// authorizeToken() as CPO always resolves to {allowed}, never throws for
	// a business error — BLOCKED for a nonexistent token.
	authResp, err := client.AuthorizeToken(ctx, cpoTokenB, "AR", "ZZZ", "DOES-NOT-EXIST-SDK", nil)
	if err != nil {
		t.Fatalf("AuthorizeToken should never return a business error, got: %v", err)
	}
	if authResp.Data.Allowed != "BLOCKED" {
		t.Fatalf("expected allowed=BLOCKED, got %q", authResp.Data.Allowed)
	}
}
