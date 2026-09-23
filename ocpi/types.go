// Package ocpi implements a lightweight Go client for the LATAM EV Roaming
// Alliance OCPI 2.3.0 Hub.
//
// Today the Hub only implements the Credentials & Registration module
// server-side; every other module described in the OCPI 2.3.0 roadmap
// (Locations, Sessions, CDRs, Tariffs, Tokens, Commands, Hub Client Info,
// Invoice Reconciliation, Charging Profiles) is exposed here only as typed
// stubs that return ErrNotImplemented, so that consumers can already code
// against the final shapes.
package ocpi

import "time"

// OCPI status codes, as defined by the Hub in lib/ocpi/response.ts.
// These follow the OCPI 2.3.0 convention: 1xxx success, 2xxx client errors,
// 3xxx server errors.
const (
	StatusSuccess            = 1000
	StatusClientError        = 2000
	StatusInvalidParameters  = 2001
	StatusNotEnoughInfo      = 2002
	StatusUnknownToken       = 2003
	StatusServerError        = 3000
	StatusUnableToUseAPI     = 3001
	StatusUnsupportedVersion = 3002
)

// RoleType identifies the OCPI party role a set of credentials acts as.
type RoleType string

const (
	RoleCPO  RoleType = "CPO"
	RoleEMSP RoleType = "EMSP"
	RoleHUB  RoleType = "HUB"
)

// Role describes one role advertised by a party during the Credentials
// handshake (POST /credentials body, and the roles returned by the Hub).
// It mirrors the `roles` array documented in
// app/api/ocpi/2.3.0/credentials/route.ts and lib/ocpi/credentials.ts.
type Role struct {
	Role        RoleType `json:"role"`
	PartyID     string   `json:"party_id"`
	CountryCode string   `json:"country_code"`
}

// Envelope is the generic OCPI response envelope used by every endpoint on
// the Hub. See lib/ocpi/response.ts (ocpiSuccess / ocpiError) for the
// server-side implementation this mirrors.
type Envelope[T any] struct {
	Data          T         `json:"data"`
	StatusCode    int       `json:"status_code"`
	StatusMessage string    `json:"status_message"`
	Timestamp     time.Time `json:"timestamp"`
}

// Version describes one supported OCPI version, as returned by
// GET /ocpi/2.3.0/versions.
type Version struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

// VersionsData is the `data` payload of GET /versions.
type VersionsData = []Version

// VersionsResponse is the full envelope returned by GET /versions.
type VersionsResponse = Envelope[VersionsData]

// Endpoint describes one module endpoint exposed by a party, as returned by
// GET /ocpi/2.3.0/details.
type Endpoint struct {
	Identifier string `json:"identifier"`
	Role       string `json:"role"`
	URL        string `json:"url"`
}

// DetailsData is the `data` payload of GET /details.
type DetailsData struct {
	Version   string     `json:"version"`
	Endpoints []Endpoint `json:"endpoints"`
}

// DetailsResponse is the full envelope returned by GET /details.
type DetailsResponse = Envelope[DetailsData]

// Credentials is the `data` payload exchanged during the Credentials &
// Registration handshake (POST/PUT /credentials).
type Credentials struct {
	Token string `json:"token"`
	URL   string `json:"url"`
	Roles []Role `json:"roles"`
}

// CredentialsResponse is the full envelope returned by POST/PUT /credentials.
type CredentialsResponse = Envelope[Credentials]

// credentialsRequestBody is the JSON body sent by the client when
// registering or renewing credentials.
type credentialsRequestBody struct {
	Token string `json:"token"`
	URL   string `json:"url"`
	Roles []Role `json:"roles"`
}
