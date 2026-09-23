package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Tariffs module (mod_tariffs), which the Hub
// implements server-side. Field shapes mirror the public JSON emitted by
// toPublicTariff() in lib/ocpi/tariffs.ts on the Hub.

// PriceComponent is one line item of a TariffElement.
type PriceComponent struct {
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Vat      float64 `json:"vat,omitempty"`
	StepSize int     `json:"step_size"`
}

// TariffElement groups PriceComponents under (optional) restrictions.
type TariffElement struct {
	PriceComponents []PriceComponent `json:"price_components"`
}

// Tariff describes a pricing structure (mod_tariffs).
type Tariff struct {
	CountryCode string          `json:"country_code"`
	PartyID     string          `json:"party_id"`
	ID          string          `json:"id"`
	Currency    string          `json:"currency"`
	Elements    []TariffElement `json:"elements"`
	LastUpdated string          `json:"last_updated,omitempty"`
}

// TariffsResponse is the full envelope returned by GET /tariffs.
type TariffsResponse = Envelope[[]Tariff]

// TariffResponse is the full envelope returned by
// GET/PUT /tariffs/{country_code}/{party_id}/{tariff_id}.
type TariffResponse = Envelope[Tariff]

// TariffInput is the request body for PUT
// /tariffs/{country_code}/{party_id}/{tariff_id}.
type TariffInput struct {
	ID       string          `json:"id"`
	Currency string          `json:"currency"`
	Elements []TariffElement `json:"elements"`
}

// GetTariffs calls GET /tariffs and lists the Tariffs visible to the caller.
func (c *Client) GetTariffs(ctx context.Context, tokenB string, offset, limit int) (*TariffsResponse, error) {
	url := c.baseURL + "/tariffs" + paginationQuery(offset, limit)
	return doRequest[[]Tariff](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetTariff calls GET /tariffs/{country_code}/{party_id}/{tariff_id} and
// fetches a single Tariff by ID.
func (c *Client) GetTariff(ctx context.Context, tokenB, countryCode, partyID, tariffID string) (*TariffResponse, error) {
	url := c.baseURL + "/tariffs/" + countryCode + "/" + partyID + "/" + tariffID
	return doRequest[Tariff](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PutTariff calls PUT /tariffs/{country_code}/{party_id}/{tariff_id} to
// create or replace a Tariff. Requires a CPO role matching
// countryCode/partyID.
func (c *Client) PutTariff(ctx context.Context, tokenB, countryCode, partyID, tariffID string, body TariffInput) (*TariffResponse, error) {
	url := c.baseURL + "/tariffs/" + countryCode + "/" + partyID + "/" + tariffID
	return doRequest[Tariff](ctx, c, http.MethodPut, url, tokenB, body)
}

// DeleteTariff calls DELETE /tariffs/{country_code}/{party_id}/{tariff_id}
// to remove a Tariff.
func (c *Client) DeleteTariff(ctx context.Context, tokenB, countryCode, partyID, tariffID string) error {
	url := c.baseURL + "/tariffs/" + countryCode + "/" + partyID + "/" + tariffID
	_, err := doRequest[map[string]any](ctx, c, http.MethodDelete, url, tokenB, nil)
	return err
}
