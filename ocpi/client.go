package ocpi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// DefaultBaseURL points at the LATAM EV Roaming Alliance Hub's OCPI 2.3.0
// root. Override it via WithBaseURL when talking to a staging environment
// or a local instance of the Hub.
const DefaultBaseURL = "https://latam-evra.org/api/ocpi/2.3.0"

// Client is a lightweight OCPI 2.3.0 client for the LATAM EV Roaming
// Alliance Hub. It uses only the standard library for HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the Hub's OCPI 2.3.0 base URL (default:
// DefaultBaseURL). Useful for pointing at a local/staging Hub, e.g.
// "http://localhost:3000/api/ocpi/2.3.0".
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient overrides the underlying *http.Client (default:
// &http.Client{Timeout: 30 * time.Second}).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient builds a new Hub client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// doRequest performs an HTTP request against the Hub, decodes the OCPI
// envelope into env, and translates an OCPI-level error (status_code !=
// StatusSuccess) into an *OcpiError.
func doRequest[T any](ctx context.Context, c *Client, method, url, token string, body any) (*Envelope[T], error) {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("ocpi-go: encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("ocpi-go: building request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ocpi-go: performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ocpi-go: reading response body: %w", err)
	}

	if resp.StatusCode >= 500 && len(bytes.TrimSpace(respBody)) == 0 {
		return nil, &OcpiError{
			StatusCode:    StatusServerError,
			StatusMessage: resp.Status,
			HTTPStatus:    resp.StatusCode,
		}
	}

	var env Envelope[T]
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("ocpi-go: decoding response body (http %d): %w", resp.StatusCode, err)
	}

	if env.StatusCode != StatusSuccess {
		return &env, &OcpiError{
			StatusCode:    env.StatusCode,
			StatusMessage: env.StatusMessage,
			HTTPStatus:    resp.StatusCode,
		}
	}

	return &env, nil
}

// paginationQuery builds a "?offset=N&limit=N" query string, shared by every
// list endpoint (GetLocations, GetTariffs, ...).
func paginationQuery(offset, limit int) string {
	return "?offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(limit)
}

// GetVersions calls GET /versions and returns the list of OCPI versions the
// Hub advertises.
func (c *Client) GetVersions(ctx context.Context) (*VersionsResponse, error) {
	url := c.baseURL + "/versions"
	return doRequest[VersionsData](ctx, c, http.MethodGet, url, "", nil)
}

// GetDetails calls GET /details and returns the module endpoints exposed by
// the Hub for OCPI 2.3.0.
func (c *Client) GetDetails(ctx context.Context) (*DetailsResponse, error) {
	url := c.baseURL + "/details"
	return doRequest[DetailsData](ctx, c, http.MethodGet, url, "", nil)
}

// RegisterCredentials performs the initial Credentials & Registration
// handshake: POST /credentials with `Authorization: Token <tokenA>`. On
// success the Hub returns a new TOKEN_B (Credentials.Token) that must be
// used for all subsequent authenticated calls, including RenewCredentials
// and TerminateCredentials.
func (c *Client) RegisterCredentials(ctx context.Context, tokenA string, csmsURL string, roles []Role) (*CredentialsResponse, error) {
	endpoint := c.baseURL + "/credentials"
	body := credentialsRequestBody{
		Token: tokenA,
		URL:   csmsURL,
		Roles: roles,
	}
	return doRequest[Credentials](ctx, c, http.MethodPost, endpoint, tokenA, body)
}

// RenewCredentials calls PUT /credentials with `Authorization: Token
// <tokenB>` to rotate the current TOKEN_B for a new one.
func (c *Client) RenewCredentials(ctx context.Context, tokenB string) (*CredentialsResponse, error) {
	endpoint := c.baseURL + "/credentials"
	return doRequest[Credentials](ctx, c, http.MethodPut, endpoint, tokenB, nil)
}

// TerminateCredentials calls DELETE /credentials with `Authorization: Token
// <tokenB>` to terminate the connection between the CSMS and the Hub.
func (c *Client) TerminateCredentials(ctx context.Context, tokenB string) error {
	endpoint := c.baseURL + "/credentials"
	_, err := doRequest[map[string]any](ctx, c, http.MethodDelete, endpoint, tokenB, nil)
	return err
}
