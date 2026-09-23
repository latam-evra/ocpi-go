package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Locations module (mod_locations), which the Hub
// implements server-side. Field shapes mirror the public JSON emitted by
// toPublicLocation() in lib/ocpi/locations.ts on the Hub.

// Coordinates is a WGS84 latitude/longitude pair, as used in Location.
type Coordinates struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Connector describes one physical connector on an EVSE.
type Connector struct {
	ID                 string   `json:"id"`
	Standard           string   `json:"standard"`
	Format             string   `json:"format"`
	PowerType          string   `json:"power_type"`
	MaxVoltage         int      `json:"max_voltage"`
	MaxAmperage        int      `json:"max_amperage"`
	MaxElectricPower   int      `json:"max_electric_power,omitempty"`
	TariffIDs          []string `json:"tariff_ids,omitempty"`
	TermsAndConditions string   `json:"terms_and_conditions,omitempty"`
	LastUpdated        string   `json:"last_updated,omitempty"`
}

// EVSE describes one charge point within a Location.
type EVSE struct {
	UID         string      `json:"uid"`
	EvseID      string      `json:"evse_id,omitempty"`
	Status      string      `json:"status"`
	Connectors  []Connector `json:"connectors"`
	LastUpdated string      `json:"last_updated,omitempty"`
}

// Location describes a charging site (mod_locations).
type Location struct {
	CountryCode string      `json:"country_code"`
	PartyID     string      `json:"party_id"`
	ID          string      `json:"id"`
	Publish     bool        `json:"publish"`
	Name        string      `json:"name,omitempty"`
	Address     string      `json:"address"`
	City        string      `json:"city"`
	PostalCode  string      `json:"postal_code,omitempty"`
	Country     string      `json:"country"`
	Coordinates Coordinates `json:"coordinates"`
	EVSEs       []EVSE      `json:"evses"`
	LastUpdated string      `json:"last_updated,omitempty"`
}

// LocationsResponse is the full envelope returned by GET /locations.
type LocationsResponse = Envelope[[]Location]

// LocationResponse is the full envelope returned by
// GET/PUT/PATCH /locations/{country_code}/{party_id}/{location_id}.
type LocationResponse = Envelope[Location]

// ConnectorInput is the request shape for one connector within an EvseInput.
type ConnectorInput struct {
	ID                 string   `json:"id"`
	Standard           string   `json:"standard"`
	Format             string   `json:"format"`
	PowerType          string   `json:"power_type"`
	MaxVoltage         int      `json:"max_voltage"`
	MaxAmperage        int      `json:"max_amperage"`
	MaxElectricPower   int      `json:"max_electric_power,omitempty"`
	TariffIDs          []string `json:"tariff_ids,omitempty"`
	TermsAndConditions string   `json:"terms_and_conditions,omitempty"`
}

// EvseInput is the request shape for one EVSE within a LocationInput.
type EvseInput struct {
	UID        string           `json:"uid"`
	EvseID     string           `json:"evse_id,omitempty"`
	Status     string           `json:"status"`
	Connectors []ConnectorInput `json:"connectors"`
}

// LocationInput is the request body for PUT
// /locations/{country_code}/{party_id}/{location_id}.
type LocationInput struct {
	ID          string      `json:"id"`
	Publish     bool        `json:"publish"`
	Name        string      `json:"name,omitempty"`
	Address     string      `json:"address"`
	City        string      `json:"city"`
	PostalCode  string      `json:"postal_code,omitempty"`
	Country     string      `json:"country"`
	Coordinates Coordinates `json:"coordinates"`
	EVSEs       []EvseInput `json:"evses,omitempty"`
}

// GetLocations calls GET /locations and lists the published Locations
// visible to the caller.
func (c *Client) GetLocations(ctx context.Context, tokenB string, offset, limit int) (*LocationsResponse, error) {
	url := c.baseURL + "/locations" + paginationQuery(offset, limit)
	return doRequest[[]Location](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetLocation calls GET /locations/{country_code}/{party_id}/{location_id}
// and fetches a single Location by ID.
func (c *Client) GetLocation(ctx context.Context, tokenB, countryCode, partyID, locationID string) (*LocationResponse, error) {
	url := c.baseURL + "/locations/" + countryCode + "/" + partyID + "/" + locationID
	return doRequest[Location](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PutLocation calls PUT /locations/{country_code}/{party_id}/{location_id}
// to create or fully replace a Location. Requires a CPO role matching
// countryCode/partyID.
func (c *Client) PutLocation(ctx context.Context, tokenB, countryCode, partyID, locationID string, body LocationInput) (*LocationResponse, error) {
	url := c.baseURL + "/locations/" + countryCode + "/" + partyID + "/" + locationID
	return doRequest[Location](ctx, c, http.MethodPut, url, tokenB, body)
}

// PatchLocation calls PATCH /locations/{country_code}/{party_id}/{location_id}
// to partially update a Location. Only non-zero fields of body should be
// set by the caller.
func (c *Client) PatchLocation(ctx context.Context, tokenB, countryCode, partyID, locationID string, body map[string]any) (*LocationResponse, error) {
	url := c.baseURL + "/locations/" + countryCode + "/" + partyID + "/" + locationID
	return doRequest[Location](ctx, c, http.MethodPatch, url, tokenB, body)
}
