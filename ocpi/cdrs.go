package ocpi

import (
	"context"
	"net/http"
)

// This file implements the CDRs module (mod_cdrs), which the Hub
// implements server-side. Field shapes mirror the public JSON emitted by
// toPublicCdr() in lib/ocpi/cdrs.ts on the Hub. CdrToken and
// ChargingPeriodDimension/ChargingPeriod are shared with sessions.go.

// CdrLocation is a snapshot of the Location a CDR was generated at.
type CdrLocation struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name,omitempty"`
	Address            string      `json:"address"`
	City               string      `json:"city"`
	PostalCode         string      `json:"postal_code,omitempty"`
	Country            string      `json:"country"`
	Coordinates        Coordinates `json:"coordinates"`
	EVSEID             string      `json:"evse_id,omitempty"`
	EVSEUID            string      `json:"evse_uid"`
	ConnectorID        string      `json:"connector_id"`
	ConnectorStandard  string      `json:"connector_standard"`
	ConnectorFormat    string      `json:"connector_format"`
	ConnectorPowerType string      `json:"connector_power_type"`
}

// CdrTariff is a Tariff snapshot embedded in a CDR — its own shape,
// distinct from Tariff in tariffs.go, but reuses PriceComponent and
// TariffElement (defined in tariffs.go) since those sub-shapes match.
type CdrTariff struct {
	CountryCode string          `json:"country_code"`
	PartyID     string          `json:"party_id"`
	ID          string          `json:"id"`
	Currency    string          `json:"currency"`
	Elements    []TariffElement `json:"elements"`
}

// CdrInput is the request body for POST
// /cdrs/{country_code}/{party_id}/{cdr_id}.
type CdrInput struct {
	ID                     string           `json:"id"`
	StartDateTime          string           `json:"start_date_time"`
	EndDateTime            string           `json:"end_date_time"`
	SessionID              string           `json:"session_id,omitempty"`
	CdrToken               CdrToken         `json:"cdr_token"`
	AuthMethod             AuthMethod       `json:"auth_method"`
	AuthorizationReference string           `json:"authorization_reference,omitempty"`
	CdrLocation            CdrLocation      `json:"cdr_location"`
	MeterID                string           `json:"meter_id,omitempty"`
	Currency               string           `json:"currency"`
	Tariffs                []CdrTariff      `json:"tariffs,omitempty"`
	ChargingPeriods        []ChargingPeriod `json:"charging_periods"`
	SignedData             map[string]any   `json:"signed_data,omitempty"`
	TotalCost              CostAmount       `json:"total_cost"`
	TotalFixedCost         *CostAmount      `json:"total_fixed_cost,omitempty"`
	TotalEnergy            float64          `json:"total_energy"`
	TotalEnergyCost        *CostAmount      `json:"total_energy_cost,omitempty"`
	TotalTime              float64          `json:"total_time"`
	TotalTimeCost          *CostAmount      `json:"total_time_cost,omitempty"`
	TotalParkingTime       *float64         `json:"total_parking_time,omitempty"`
	TotalParkingCost       *CostAmount      `json:"total_parking_cost,omitempty"`
	TotalReservationCost   *CostAmount      `json:"total_reservation_cost,omitempty"`
	Remark                 string           `json:"remark,omitempty"`
	InvoiceReferenceID     string           `json:"invoice_reference_id,omitempty"`
	Credit                 *bool            `json:"credit,omitempty"`
	CreditReferenceID      string           `json:"credit_reference_id,omitempty"`
}

// Cdr is a Charge Detail Record (mod_cdrs), the definitive billing record
// for a finished charging session. CDRs are immutable: there is no
// PUT/PATCH/DELETE, only POST (create) and GET.
type Cdr struct {
	CountryCode            string           `json:"country_code"`
	PartyID                string           `json:"party_id"`
	ID                     string           `json:"id"`
	StartDateTime          string           `json:"start_date_time"`
	EndDateTime            string           `json:"end_date_time"`
	SessionID              string           `json:"session_id,omitempty"`
	CdrToken               CdrToken         `json:"cdr_token"`
	AuthMethod             AuthMethod       `json:"auth_method"`
	AuthorizationReference string           `json:"authorization_reference,omitempty"`
	CdrLocation            CdrLocation      `json:"cdr_location"`
	MeterID                string           `json:"meter_id,omitempty"`
	Currency               string           `json:"currency"`
	Tariffs                []CdrTariff      `json:"tariffs,omitempty"`
	ChargingPeriods        []ChargingPeriod `json:"charging_periods"`
	SignedData             map[string]any   `json:"signed_data,omitempty"`
	TotalCost              CostAmount       `json:"total_cost"`
	TotalFixedCost         *CostAmount      `json:"total_fixed_cost,omitempty"`
	TotalEnergy            float64          `json:"total_energy"`
	TotalEnergyCost        *CostAmount      `json:"total_energy_cost,omitempty"`
	TotalTime              float64          `json:"total_time"`
	TotalTimeCost          *CostAmount      `json:"total_time_cost,omitempty"`
	TotalParkingTime       *float64         `json:"total_parking_time,omitempty"`
	TotalParkingCost       *CostAmount      `json:"total_parking_cost,omitempty"`
	TotalReservationCost   *CostAmount      `json:"total_reservation_cost,omitempty"`
	Remark                 string           `json:"remark,omitempty"`
	InvoiceReferenceID     string           `json:"invoice_reference_id,omitempty"`
	Credit                 *bool            `json:"credit,omitempty"`
	CreditReferenceID      string           `json:"credit_reference_id,omitempty"`
	LastUpdated            string           `json:"last_updated"`
}

// CdrsResponse is the full envelope returned by GET /cdrs.
type CdrsResponse = Envelope[[]Cdr]

// CdrResponse is the full envelope returned by
// GET/POST /cdrs/{country_code}/{party_id}/{cdr_id}.
type CdrResponse = Envelope[Cdr]

// GetCdrs calls GET /cdrs and lists the CDRs visible to the caller.
func (c *Client) GetCdrs(ctx context.Context, tokenB string, offset, limit int) (*CdrsResponse, error) {
	url := c.baseURL + "/cdrs" + paginationQuery(offset, limit)
	return doRequest[[]Cdr](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetCdr calls GET /cdrs/{country_code}/{party_id}/{cdr_id} and fetches a
// single Cdr by ID. Returns 404 if it doesn't exist.
func (c *Client) GetCdr(ctx context.Context, tokenB, countryCode, partyID, cdrID string) (*CdrResponse, error) {
	url := c.baseURL + "/cdrs/" + countryCode + "/" + partyID + "/" + cdrID
	return doRequest[Cdr](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PostCdr calls POST /cdrs/{country_code}/{party_id}/{cdr_id} to create a
// new, immutable Cdr. Requires a CPO role matching countryCode/partyID. A
// second POST with the same id returns a 409 *OcpiError — CDRs have no
// PUT/PATCH/DELETE by OCPI design.
func (c *Client) PostCdr(ctx context.Context, tokenB, countryCode, partyID, cdrID string, body CdrInput) (*CdrResponse, error) {
	url := c.baseURL + "/cdrs/" + countryCode + "/" + partyID + "/" + cdrID
	return doRequest[Cdr](ctx, c, http.MethodPost, url, tokenB, body)
}
