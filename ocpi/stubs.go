package ocpi

import "context"

// This file contains typed request/response shapes for the OCPI 2.3.0
// modules still on the Hub's public roadmap (see
// components/ModuleAccordion.tsx and docs/Roaming_hub_Latam.md in the Hub
// repository). Locations, Tariffs, Hub Client Info, Sessions, CDRs,
// Tokens & Authorisation, Commands and Invoice Reconciliation are all
// implemented server-side now — see locations.go, tariffs.go,
// hub_client_info.go, sessions.go, cdrs.go, tokens.go, commands.go and
// invoice_reconciliation.go. Charging Profiles remains the only module
// still a stub, since it is not on the Hub's roadmap.

// -----------------------------------------------------------------------
// Charging Profiles (Smart Charging)
// -----------------------------------------------------------------------

// ChargingProfile describes a smart-charging power limit schedule
// (mod_charging_profiles).
type ChargingProfile struct {
	StartDateTime    string  `json:"start_date_time"`
	ChargingRateUnit string  `json:"charging_rate_unit"`
	Limit            float64 `json:"limit"`
}

// SetChargingProfile would push a ChargingProfile for a given session to a
// CPO via the Hub. Not implemented by the Hub yet.
func (c *Client) SetChargingProfile(ctx context.Context, sessionID string, profile ChargingProfile) error {
	return notImplementedError("Charging Profiles")
}
