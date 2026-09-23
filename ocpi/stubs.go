package ocpi

import "context"

// This file contains typed request/response shapes for the OCPI 2.3.0
// modules still on the Hub's public roadmap (see
// components/ModuleAccordion.tsx and docs/Roaming_hub_Latam.md in the Hub
// repository). Locations, Tariffs and Hub Client Info are implemented
// server-side — see locations.go, tariffs.go and hub_client_info.go —
// every other method below still returns ErrNotImplemented, but the types
// are already modeled so that consumers can code against the final shapes
// ahead of time.

// -----------------------------------------------------------------------
// Sessions
// -----------------------------------------------------------------------

// Session describes an in-progress or finished charging session
// (mod_sessions).
type Session struct {
	ID            string  `json:"id"`
	StartDateTime string  `json:"start_date_time"`
	KWh           float64 `json:"kwh"`
	Status        string  `json:"status"`
}

// GetActiveSession would fetch a single Session by ID.
// Not implemented by the Hub yet.
func (c *Client) GetActiveSession(ctx context.Context, sessionID string) (*Session, error) {
	return nil, notImplementedError("Sessions")
}

// -----------------------------------------------------------------------
// CDRs (Charge Detail Records)
// -----------------------------------------------------------------------

// CdrToken identifies the token used to authorize the session a CDR bills.
type CdrToken struct {
	CountryCode string `json:"country_code"`
	PartyID     string `json:"party_id"`
	UID         string `json:"uid"`
	Type        string `json:"type"`
	ContractID  string `json:"contract_id"`
}

// CdrLocation is a snapshot of the Location a CDR was generated at.
type CdrLocation struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Address            string      `json:"address"`
	City               string      `json:"city"`
	PostalCode         string      `json:"postal_code"`
	Country            string      `json:"country"`
	Coordinates        Coordinates `json:"coordinates"`
	EVSEID             string      `json:"evse_id"`
	EVSEUID            string      `json:"evse_uid"`
	ConnectorID        string      `json:"connector_id"`
	ConnectorStandard  string      `json:"connector_standard"`
	ConnectorFormat    string      `json:"connector_format"`
	ConnectorPowerType string      `json:"connector_power_type"`
}

// ChargingPeriodDimension is one metered dimension (ENERGY, TIME, ...)
// within a ChargingPeriod.
type ChargingPeriodDimension struct {
	Type   string  `json:"type"`
	Volume float64 `json:"volume"`
}

// ChargingPeriod is one metered slice of a charging session, used in CDRs.
type ChargingPeriod struct {
	StartDateTime string                    `json:"start_date_time"`
	Dimensions    []ChargingPeriodDimension `json:"dimensions"`
}

// Price splits an amount into its tax-excluded and tax-included values, as
// used throughout mod_cdrs.
type Price struct {
	ExclVat float64 `json:"excl_vat"`
	InclVat float64 `json:"incl_vat"`
}

// Cdr is a Charge Detail Record (mod_cdrs), the definitive billing record
// for a finished charging session. Field shapes mirror
// docs/object_ocpi_cdr.md in the Hub repository.
type Cdr struct {
	CountryCode            string           `json:"country_code"`
	PartyID                string           `json:"party_id"`
	ID                     string           `json:"id"`
	StartDateTime          string           `json:"start_date_time"`
	EndDateTime            string           `json:"end_date_time"`
	SessionID              string           `json:"session_id"`
	CdrToken               CdrToken         `json:"cdr_token"`
	AuthMethod             string           `json:"auth_method"`
	AuthorizationReference string           `json:"authorization_reference"`
	CdrLocation            CdrLocation      `json:"cdr_location"`
	Currency               string           `json:"currency"`
	Tariffs                []Tariff         `json:"tariffs"`
	ChargingPeriods        []ChargingPeriod `json:"charging_periods"`
	TotalCost              Price            `json:"total_cost"`
	TotalFixedCost         Price            `json:"total_fixed_cost"`
	TotalEnergy            float64          `json:"total_energy"`
	TotalEnergyCost        Price            `json:"total_energy_cost"`
	TotalTime              float64          `json:"total_time"`
	TotalTimeCost          Price            `json:"total_time_cost"`
	LastUpdated            string           `json:"last_updated"`
}

// GetCdrs would list CDRs visible to the caller.
// Not implemented by the Hub yet.
func (c *Client) GetCdrs(ctx context.Context) ([]Cdr, error) {
	return nil, notImplementedError("CDRs")
}

// SubmitCdr would submit a new CDR to the Hub (CPO -> Hub -> eMSP flow).
// Not implemented by the Hub yet.
func (c *Client) SubmitCdr(ctx context.Context, cdr Cdr) error {
	return notImplementedError("CDRs")
}

// -----------------------------------------------------------------------
// Tokens & Authorisation
// -----------------------------------------------------------------------

// Token identifies an end-user credential (RFID, app, ...) used to
// authorize charging sessions (mod_tokens).
type Token struct {
	UID        string `json:"uid"`
	Type       string `json:"type"`
	ContractID string `json:"contract_id"`
}

// AuthorizeToken would ask the Hub to authorize a Token for a charging
// session. Not implemented by the Hub yet.
func (c *Client) AuthorizeToken(ctx context.Context, tokenUID string) (*Token, error) {
	return nil, notImplementedError("Tokens & Authorisation")
}

// -----------------------------------------------------------------------
// Commands
// -----------------------------------------------------------------------

// CommandTokenRef references the Token a remote command acts on behalf of.
type CommandTokenRef struct {
	UID string `json:"uid"`
}

// StartSessionCommand is the payload for POST /commands/START_SESSION.
type StartSessionCommand struct {
	Token      CommandTokenRef `json:"token"`
	LocationID string          `json:"location_id"`
	EVSEUID    string          `json:"evse_uid"`
}

// StartSession would send a remote START_SESSION command to a CPO via the
// Hub. Not implemented by the Hub yet.
func (c *Client) StartSession(ctx context.Context, cmd StartSessionCommand) error {
	return notImplementedError("Commands")
}

// StopSession would send a remote STOP_SESSION command to a CPO via the
// Hub. Not implemented by the Hub yet.
func (c *Client) StopSession(ctx context.Context, sessionID string) error {
	return notImplementedError("Commands")
}

// UnlockConnector would send a remote UNLOCK_CONNECTOR command to a CPO via
// the Hub. Not implemented by the Hub yet.
func (c *Client) UnlockConnector(ctx context.Context, locationID, evseUID string) error {
	return notImplementedError("Commands")
}

// -----------------------------------------------------------------------
// Invoice Reconciliation (OCPI 2.3.0 Edition 2)
// -----------------------------------------------------------------------

// InvoiceReconciliation matches a CDR against an invoice reference for
// financial reconciliation (mod_invoicereconciliations).
type InvoiceReconciliation struct {
	CdrID            string `json:"cdr_id"`
	InvoiceReference string `json:"invoice_reference"`
	Status           string `json:"status"`
}

// GetInvoiceReconciliations would list invoice reconciliation records.
// Not implemented by the Hub yet.
func (c *Client) GetInvoiceReconciliations(ctx context.Context) ([]InvoiceReconciliation, error) {
	return nil, notImplementedError("Invoice Reconciliation")
}

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
