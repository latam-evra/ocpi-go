package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Hub Client Info module (mod_hubclientinfo),
// which the Hub implements server-side. Field shapes mirror
// lib/ocpi/hubClientInfo.ts (toEntry / STATUS_MAP) on the Hub.

// HubClientInfoEntry describes the connection status of one party known to
// the Hub.
type HubClientInfoEntry struct {
	PartyID     string `json:"party_id"`
	CountryCode string `json:"country_code"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	LastUpdated string `json:"last_updated"`
}

// HubClientInfoListResponse is the full envelope returned by
// GET /hubclientinfo.
type HubClientInfoListResponse = Envelope[[]HubClientInfoEntry]

// ListHubClientInfo calls GET /hubclientinfo and lists the connection
// status of every party known to the Hub.
func (c *Client) ListHubClientInfo(ctx context.Context, tokenB string, offset, limit int) (*HubClientInfoListResponse, error) {
	url := c.baseURL + "/hubclientinfo" + paginationQuery(offset, limit)
	return doRequest[[]HubClientInfoEntry](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetHubClientInfo calls GET /hubclientinfo/{country_code}/{party_id} and
// lists the roles/status known to the Hub for that party.
func (c *Client) GetHubClientInfo(ctx context.Context, tokenB, countryCode, partyID string) (*HubClientInfoListResponse, error) {
	url := c.baseURL + "/hubclientinfo/" + countryCode + "/" + partyID
	return doRequest[[]HubClientInfoEntry](ctx, c, http.MethodGet, url, tokenB, nil)
}
