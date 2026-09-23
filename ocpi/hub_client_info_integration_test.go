//go:build integration

package ocpi

import (
	"context"
	"testing"
)

// Mirrors sdks/node/tests/integration/hub-client-info.integration.test.ts
// and sdks/python/tests/integration/test_hub_client_info_integration.py.
// Reuses hubTestURL/createTestRegistration/cleanupTestRegistration from
// integration_test.go.

func TestIntegration_HubClientInfo_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	reg := createTestRegistration(t, "CPO", "CL", "GO3")
	defer cleanupTestRegistration(t, reg.RegistrationID)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	credsResp, err := client.RegisterCredentials(ctx, reg.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "GO3", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials failed: %v", err)
	}
	tokenB := credsResp.Data.Token

	list, err := client.ListHubClientInfo(ctx, tokenB, 0, 100)
	if err != nil {
		t.Fatalf("ListHubClientInfo failed: %v", err)
	}
	found := false
	for _, entry := range list.Data {
		if entry.PartyID == "GO3" && entry.CountryCode == "CL" {
			found = true
			if entry.Role != "CPO" {
				t.Fatalf("expected role CPO, got %q", entry.Role)
			}
			if entry.Status != "CONNECTED" {
				t.Fatalf("expected status CONNECTED, got %q", entry.Status)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected GO3/CL entry in hub client info list, got %+v", list.Data)
	}

	byRole, err := client.GetHubClientInfo(ctx, tokenB, "CL", "GO3")
	if err != nil {
		t.Fatalf("GetHubClientInfo failed: %v", err)
	}
	if len(byRole.Data) != 1 || byRole.Data[0].Status != "CONNECTED" {
		t.Fatalf("unexpected GetHubClientInfo data: %+v", byRole.Data)
	}
}
