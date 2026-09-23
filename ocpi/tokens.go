package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Tokens & Authorisation module (mod_tokens),
// which the Hub implements server-side. Field shapes mirror the public
// JSON emitted by toPublicToken() in lib/ocpi/tokens.ts on the Hub.
// TokenType is shared with sessions.go/cdrs.go (CdrToken).

// TokenWhitelist controls how a CPO should treat a Token when offline.
type TokenWhitelist string

const (
	TokenWhitelistNever          TokenWhitelist = "NEVER"
	TokenWhitelistAllowed        TokenWhitelist = "ALLOWED"
	TokenWhitelistAllowedOffline TokenWhitelist = "ALLOWED_OFFLINE"
	TokenWhitelistAlways         TokenWhitelist = "ALWAYS"
)

// TokenInput is the request body for PUT/PATCH
// /tokens/{country_code}/{party_id}/{token_uid}.
type TokenInput struct {
	UID                string         `json:"uid"`
	Type               TokenType      `json:"type"`
	ContractID         string         `json:"contract_id"`
	VisualNumber       string         `json:"visual_number,omitempty"`
	Issuer             string         `json:"issuer"`
	GroupID            string         `json:"group_id,omitempty"`
	Valid              bool           `json:"valid"`
	Whitelist          TokenWhitelist `json:"whitelist"`
	Language           string         `json:"language,omitempty"`
	DefaultProfileType string         `json:"default_profile_type,omitempty"`
	EnergyContract     map[string]any `json:"energy_contract,omitempty"`
}

// Token identifies an end-user credential (RFID, app, ...) used to
// authorize charging sessions (mod_tokens), as returned by the Hub.
type Token struct {
	CountryCode        string         `json:"country_code"`
	PartyID            string         `json:"party_id"`
	UID                string         `json:"uid"`
	Type               TokenType      `json:"type"`
	ContractID         string         `json:"contract_id"`
	VisualNumber       string         `json:"visual_number,omitempty"`
	Issuer             string         `json:"issuer"`
	GroupID            string         `json:"group_id,omitempty"`
	Valid              bool           `json:"valid"`
	Whitelist          TokenWhitelist `json:"whitelist"`
	Language           string         `json:"language,omitempty"`
	DefaultProfileType string         `json:"default_profile_type,omitempty"`
	EnergyContract     map[string]any `json:"energy_contract,omitempty"`
	LastUpdated        string         `json:"last_updated"`
}

// AuthorizeResult is the response of POST
// /tokens/{country_code}/{party_id}/{token_uid}/authorize. Allowed is kept
// as a plain string rather than a closed enum — observed values are
// "ALLOWED"/"BLOCKED", but the Hub forwards whatever the remote eMSP
// returns verbatim.
type AuthorizeResult struct {
	Allowed string `json:"allowed"`
}

// TokensResponse is the full envelope returned by GET /tokens.
type TokensResponse = Envelope[[]Token]

// TokenResponse is the full envelope returned by
// GET/PUT/PATCH /tokens/{country_code}/{party_id}/{token_uid}.
type TokenResponse = Envelope[Token]

// AuthorizeResultResponse is the full envelope returned by
// POST /tokens/{country_code}/{party_id}/{token_uid}/authorize.
type AuthorizeResultResponse = Envelope[AuthorizeResult]

// GetTokens calls GET /tokens and lists the Tokens visible to the caller.
func (c *Client) GetTokens(ctx context.Context, tokenB string, offset, limit int) (*TokensResponse, error) {
	url := c.baseURL + "/tokens" + paginationQuery(offset, limit)
	return doRequest[[]Token](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetToken calls GET /tokens/{country_code}/{party_id}/{token_uid} and
// fetches a single Token by UID.
func (c *Client) GetToken(ctx context.Context, tokenB, countryCode, partyID, uid string) (*TokenResponse, error) {
	url := c.baseURL + "/tokens/" + countryCode + "/" + partyID + "/" + uid
	return doRequest[Token](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PutToken calls PUT /tokens/{country_code}/{party_id}/{token_uid} to
// create or fully replace a Token. Requires an EMSP role matching
// countryCode/partyID.
func (c *Client) PutToken(ctx context.Context, tokenB, countryCode, partyID, uid string, body TokenInput) (*TokenResponse, error) {
	url := c.baseURL + "/tokens/" + countryCode + "/" + partyID + "/" + uid
	return doRequest[Token](ctx, c, http.MethodPut, url, tokenB, body)
}

// PatchToken calls PATCH /tokens/{country_code}/{party_id}/{token_uid} to
// partially update a Token. Returns 404 if it doesn't exist yet.
func (c *Client) PatchToken(ctx context.Context, tokenB, countryCode, partyID, uid string, body map[string]any) (*TokenResponse, error) {
	url := c.baseURL + "/tokens/" + countryCode + "/" + partyID + "/" + uid
	return doRequest[Token](ctx, c, http.MethodPatch, url, tokenB, body)
}

// DeleteToken calls DELETE /tokens/{country_code}/{party_id}/{token_uid} to
// remove a Token.
func (c *Client) DeleteToken(ctx context.Context, tokenB, countryCode, partyID, uid string) error {
	url := c.baseURL + "/tokens/" + countryCode + "/" + partyID + "/" + uid
	_, err := doRequest[map[string]any](ctx, c, http.MethodDelete, url, tokenB, nil)
	return err
}

// AuthorizeToken calls POST
// /tokens/{country_code}/{party_id}/{token_uid}/authorize — a special,
// non-CRUD endpoint. The caller must hold a CPO role (not necessarily
// matching countryCode/partyID of the token being authorized).
// locationReferences is optional (pass nil for an empty body {}); it is
// forwarded to the remote eMSP as-is. This call never surfaces a business
// error: any internal failure (unknown token, disconnected eMSP, 6s
// timeout) resolves to AuthorizeResult{Allowed: "BLOCKED"} from the Hub,
// not an *OcpiError.
func (c *Client) AuthorizeToken(ctx context.Context, tokenB, countryCode, partyID, uid string, locationReferences any) (*AuthorizeResultResponse, error) {
	url := c.baseURL + "/tokens/" + countryCode + "/" + partyID + "/" + uid + "/authorize"
	body := locationReferences
	if body == nil {
		body = map[string]any{}
	}
	return doRequest[AuthorizeResult](ctx, c, http.MethodPost, url, tokenB, body)
}
