package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Sessions module (mod_sessions), which the Hub
// implements server-side. Field shapes mirror the public JSON emitted by
// toPublicSession() in lib/ocpi/sessions.ts on the Hub.

// TokenType is the OCPI token type used by CdrToken/Token/CommandTokenRef.
type TokenType string

const (
	TokenTypeAdHocUser TokenType = "AD_HOC_USER"
	TokenTypeAppUser   TokenType = "APP_USER"
	TokenTypeOther     TokenType = "OTHER"
	TokenTypeRFID      TokenType = "RFID"
)

// SessionStatus is the lifecycle status of a Session.
type SessionStatus string

const (
	SessionStatusActive      SessionStatus = "ACTIVE"
	SessionStatusCompleted   SessionStatus = "COMPLETED"
	SessionStatusInvalid     SessionStatus = "INVALID"
	SessionStatusPending     SessionStatus = "PENDING"
	SessionStatusReservation SessionStatus = "RESERVATION"
)

// AuthMethod is how a Session/Cdr was authorized.
type AuthMethod string

const (
	AuthMethodAuthRequest AuthMethod = "AUTH_REQUEST"
	AuthMethodCommand     AuthMethod = "COMMAND"
	AuthMethodWhitelist   AuthMethod = "WHITELIST"
)

// CdrToken identifies the token used to authorize a Session/Cdr. Shared by
// sessions.go and cdrs.go (defined here since sessions.go is read first
// alphabetically).
type CdrToken struct {
	CountryCode string    `json:"country_code"`
	PartyID     string    `json:"party_id"`
	UID         string    `json:"uid"`
	Type        TokenType `json:"type"`
	ContractID  string    `json:"contract_id"`
}

// ChargingPeriodDimension is one metered dimension (ENERGY, TIME, ...)
// within a ChargingPeriod. Shared by sessions.go and cdrs.go.
type ChargingPeriodDimension struct {
	Type   string  `json:"type"`
	Volume float64 `json:"volume"`
}

// ChargingPeriod is one metered slice of a charging session, used in
// Sessions and CDRs.
type ChargingPeriod struct {
	StartDateTime string                    `json:"start_date_time"`
	Dimensions    []ChargingPeriodDimension `json:"dimensions"`
	TariffID      string                    `json:"tariff_id,omitempty"`
}

// CostAmount splits an amount into its tax-excluded and tax-included
// values, as used throughout mod_cdrs/mod_invoicereconciliations.
type CostAmount struct {
	ExclVat float64  `json:"excl_vat"`
	InclVat *float64 `json:"incl_vat,omitempty"`
}

// SessionInput is the request body for PUT/PATCH
// /sessions/{country_code}/{party_id}/{session_id}.
type SessionInput struct {
	ID                     string           `json:"id"`
	StartDateTime          string           `json:"start_date_time"`
	EndDateTime            string           `json:"end_date_time,omitempty"`
	KWh                    float64          `json:"kwh"`
	CdrToken               CdrToken         `json:"cdr_token"`
	AuthMethod             AuthMethod       `json:"auth_method"`
	AuthorizationReference string           `json:"authorization_reference,omitempty"`
	LocationID             string           `json:"location_id"`
	EvseUID                string           `json:"evse_uid"`
	ConnectorID            string           `json:"connector_id"`
	MeterID                string           `json:"meter_id,omitempty"`
	Currency               string           `json:"currency"`
	ChargingPeriods        []ChargingPeriod `json:"charging_periods,omitempty"`
	TotalCost              *CostAmount      `json:"total_cost,omitempty"`
	Status                 SessionStatus    `json:"status"`
}

// Session describes an in-progress or finished charging session
// (mod_sessions), as returned by the Hub.
type Session struct {
	CountryCode            string           `json:"country_code"`
	PartyID                string           `json:"party_id"`
	ID                     string           `json:"id"`
	StartDateTime          string           `json:"start_date_time"`
	EndDateTime            string           `json:"end_date_time,omitempty"`
	KWh                    float64          `json:"kwh"`
	CdrToken               CdrToken         `json:"cdr_token"`
	AuthMethod             AuthMethod       `json:"auth_method"`
	AuthorizationReference string           `json:"authorization_reference,omitempty"`
	LocationID             string           `json:"location_id"`
	EvseUID                string           `json:"evse_uid"`
	ConnectorID            string           `json:"connector_id"`
	MeterID                string           `json:"meter_id,omitempty"`
	Currency               string           `json:"currency"`
	ChargingPeriods        []ChargingPeriod `json:"charging_periods,omitempty"`
	TotalCost              *CostAmount      `json:"total_cost,omitempty"`
	Status                 SessionStatus    `json:"status"`
	LastUpdated            string           `json:"last_updated"`
}

// SessionsResponse is the full envelope returned by GET /sessions.
type SessionsResponse = Envelope[[]Session]

// SessionResponse is the full envelope returned by
// GET/PUT/PATCH /sessions/{country_code}/{party_id}/{session_id}.
type SessionResponse = Envelope[Session]

// GetSessions calls GET /sessions and lists the Sessions visible to the
// caller (any CONNECTED connection).
func (c *Client) GetSessions(ctx context.Context, tokenB string, offset, limit int) (*SessionsResponse, error) {
	url := c.baseURL + "/sessions" + paginationQuery(offset, limit)
	return doRequest[[]Session](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetSession calls GET /sessions/{country_code}/{party_id}/{session_id} and
// fetches a single Session by ID.
func (c *Client) GetSession(ctx context.Context, tokenB, countryCode, partyID, sessionID string) (*SessionResponse, error) {
	url := c.baseURL + "/sessions/" + countryCode + "/" + partyID + "/" + sessionID
	return doRequest[Session](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PutSession calls PUT /sessions/{country_code}/{party_id}/{session_id} to
// create or fully replace a Session. Requires a CPO role matching
// countryCode/partyID.
func (c *Client) PutSession(ctx context.Context, tokenB, countryCode, partyID, sessionID string, body SessionInput) (*SessionResponse, error) {
	url := c.baseURL + "/sessions/" + countryCode + "/" + partyID + "/" + sessionID
	return doRequest[Session](ctx, c, http.MethodPut, url, tokenB, body)
}

// PatchSession calls PATCH /sessions/{country_code}/{party_id}/{session_id}
// to partially update a Session. Only non-zero fields of body should be set
// by the caller. Returns 404 if the Session doesn't exist yet.
func (c *Client) PatchSession(ctx context.Context, tokenB, countryCode, partyID, sessionID string, body map[string]any) (*SessionResponse, error) {
	url := c.baseURL + "/sessions/" + countryCode + "/" + partyID + "/" + sessionID
	return doRequest[Session](ctx, c, http.MethodPatch, url, tokenB, body)
}
