package ocpi

import (
	"context"
	"testing"
)

// TestStubsReturnNotImplemented verifies that the one remaining roadmap
// module (Charging Profiles) still returns ErrNotImplemented, wrapped with
// a module-specific message.
func TestStubsReturnNotImplemented(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	err := client.SetChargingProfile(ctx, "SES-1", ChargingProfile{})
	if err == nil {
		t.Fatal("SetChargingProfile: expected an error, got nil")
	}
	if !IsNotImplemented(err) {
		t.Fatalf("SetChargingProfile: expected ErrNotImplemented, got %v", err)
	}
}
