package ocpi

import (
	"context"
	"testing"
)

// TestStubsReturnNotImplemented verifies that every roadmap module method
// returns ErrNotImplemented (wrapped with a module-specific message) until
// the Hub actually implements that module server-side.
func TestStubsReturnNotImplemented(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	checks := []struct {
		name string
		err  error
	}{
		{"GetLocations", func() error { _, err := client.GetLocations(ctx); return err }()},
		{"GetLocation", func() error { _, err := client.GetLocation(ctx, "LOC-1"); return err }()},
		{"GetActiveSession", func() error { _, err := client.GetActiveSession(ctx, "SES-1"); return err }()},
		{"GetCdrs", func() error { _, err := client.GetCdrs(ctx); return err }()},
		{"SubmitCdr", client.SubmitCdr(ctx, Cdr{})},
		{"GetTariffs", func() error { _, err := client.GetTariffs(ctx); return err }()},
		{"AuthorizeToken", func() error { _, err := client.AuthorizeToken(ctx, "RFID-1"); return err }()},
		{"StartSession", client.StartSession(ctx, StartSessionCommand{})},
		{"StopSession", client.StopSession(ctx, "SES-1")},
		{"UnlockConnector", client.UnlockConnector(ctx, "LOC-1", "EVSE-1")},
		{"GetHubClientInfo", func() error { _, err := client.GetHubClientInfo(ctx); return err }()},
		{"GetInvoiceReconciliations", func() error { _, err := client.GetInvoiceReconciliations(ctx); return err }()},
		{"SetChargingProfile", client.SetChargingProfile(ctx, "SES-1", ChargingProfile{})},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("%s: expected an error, got nil", tc.name)
			}
			if !IsNotImplemented(tc.err) {
				t.Fatalf("%s: expected ErrNotImplemented, got %v", tc.name, tc.err)
			}
		})
	}
}
