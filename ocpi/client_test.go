package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := NewClient(WithBaseURL(srv.URL))
	return c, srv.Close
}

func TestGetVersions_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/versions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"version": "2.3.0", "url": "http://example.com/details"},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetVersions(context.Background())
	if err != nil {
		t.Fatalf("GetVersions returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Version != "2.3.0" {
		t.Fatalf("unexpected versions data: %+v", resp.Data)
	}
	if resp.StatusCode != StatusSuccess {
		t.Fatalf("expected status_code %d, got %d", StatusSuccess, resp.StatusCode)
	}
}

func TestGetDetails_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/details" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"version": "2.3.0",
				"endpoints": []map[string]string{
					{"identifier": "credentials", "role": "HUB", "url": "http://example.com/credentials"},
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetDetails(context.Background())
	if err != nil {
		t.Fatalf("GetDetails returned error: %v", err)
	}
	if resp.Data.Version != "2.3.0" {
		t.Fatalf("unexpected version: %s", resp.Data.Version)
	}
	if len(resp.Data.Endpoints) != 1 || resp.Data.Endpoints[0].Identifier != "credentials" {
		t.Fatalf("unexpected endpoints: %+v", resp.Data.Endpoints)
	}
}

func TestRegisterCredentials_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/credentials" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Token TOKEN_A_TEST" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}

		var body credentialsRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Token != "TOKEN_A_TEST" || body.URL != "https://csms.example.com/ocpi/versions" {
			t.Fatalf("unexpected request body: %+v", body)
		}
		if len(body.Roles) != 1 || body.Roles[0].Role != RoleCPO {
			t.Fatalf("unexpected roles: %+v", body.Roles)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"token": "TOKEN_B_ISSUED",
				"url":   "http://example.com/versions",
				"roles": []map[string]string{
					{"role": "HUB", "party_id": "LEA", "country_code": "ZZ"},
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	roles := []Role{{Role: RoleCPO, PartyID: "CHG", CountryCode: "CL"}}
	resp, err := client.RegisterCredentials(context.Background(), "TOKEN_A_TEST", "https://csms.example.com/ocpi/versions", roles)
	if err != nil {
		t.Fatalf("RegisterCredentials returned error: %v", err)
	}
	if resp.Data.Token != "TOKEN_B_ISSUED" {
		t.Fatalf("expected issued TOKEN_B, got %q", resp.Data.Token)
	}
}

func TestRegisterCredentials_UnknownToken(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":            map[string]any{},
			"status_code":     StatusUnknownToken,
			"status_message":  "TOKEN_A inválido.",
			"timestamp":       "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	roles := []Role{{Role: RoleCPO, PartyID: "CHG", CountryCode: "CL"}}
	_, err := client.RegisterCredentials(context.Background(), "BAD_TOKEN", "https://csms.example.com/ocpi/versions", roles)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.StatusCode != StatusUnknownToken {
		t.Fatalf("expected status_code %d, got %d", StatusUnknownToken, ocpiErr.StatusCode)
	}
	if ocpiErr.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("expected http status %d, got %d", http.StatusUnauthorized, ocpiErr.HTTPStatus)
	}
}

func TestRenewCredentials_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/credentials" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Token TOKEN_B_OLD" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"token": "TOKEN_B_NEW",
				"url":   "http://example.com/versions",
				"roles": []map[string]string{
					{"role": "HUB", "party_id": "LEA", "country_code": "ZZ"},
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.RenewCredentials(context.Background(), "TOKEN_B_OLD")
	if err != nil {
		t.Fatalf("RenewCredentials returned error: %v", err)
	}
	if resp.Data.Token != "TOKEN_B_NEW" {
		t.Fatalf("expected renewed TOKEN_B, got %q", resp.Data.Token)
	}
}

func TestTerminateCredentials_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/credentials" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":            map[string]any{},
			"status_code":     StatusSuccess,
			"status_message":  "Success",
			"timestamp":       "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	if err := client.TerminateCredentials(context.Background(), "TOKEN_B_OLD"); err != nil {
		t.Fatalf("TerminateCredentials returned error: %v", err)
	}
}

func TestTerminateCredentials_UnknownToken(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":            map[string]any{},
			"status_code":     StatusUnknownToken,
			"status_message":  "TOKEN_B inválido.",
			"timestamp":       "2026-09-22T00:00:00Z",
		})
	})
	defer closeFn()

	err := client.TerminateCredentials(context.Background(), "BAD_TOKEN")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.StatusCode != StatusUnknownToken {
		t.Fatalf("expected status_code %d, got %d", StatusUnknownToken, ocpiErr.StatusCode)
	}
}
