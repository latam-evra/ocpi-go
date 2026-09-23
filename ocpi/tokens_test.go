package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPutToken_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/tokens/AR/EMS/TOK-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body TokenInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.UID != "TOK-1" || body.Whitelist != TokenWhitelistAllowed {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "AR", "party_id": "EMS", "uid": "TOK-1",
				"type": "RFID", "contract_id": "C-1", "issuer": "EvraTest",
				"valid": true, "whitelist": "ALLOWED", "last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "TOK-1", TokenInput{
		UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1", Issuer: "EvraTest",
		Valid: true, Whitelist: TokenWhitelistAllowed,
	})
	if err != nil {
		t.Fatalf("PutToken returned error: %v", err)
	}
	if resp.Data.UID != "TOK-1" || !resp.Data.Valid {
		t.Fatalf("unexpected token data: %+v", resp.Data)
	}
}

func TestPatchToken_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/tokens/AR/EMS/TOK-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "AR", "party_id": "EMS", "uid": "TOK-1",
				"type": "RFID", "contract_id": "C-1", "issuer": "EvraTest",
				"valid": false, "whitelist": "ALLOWED", "last_updated": "2026-09-23T00:00:00Z",
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PatchToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "TOK-1", map[string]any{
		"valid": false,
	})
	if err != nil {
		t.Fatalf("PatchToken returned error: %v", err)
	}
	if resp.Data.Valid {
		t.Fatalf("expected valid=false, got %+v", resp.Data)
	}
}

func TestDeleteToken_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/tokens/AR/EMS/TOK-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	if err := client.DeleteToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "TOK-1"); err != nil {
		t.Fatalf("DeleteToken returned error: %v", err)
	}
}

func TestGetToken_NotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Token no encontrado.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.GetToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "DOES-NOT-EXIST")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
}

func TestGetTokens_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tokens" || r.URL.RawQuery != "offset=0&limit=50" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"country_code": "AR", "party_id": "EMS", "uid": "TOK-1",
					"type": "RFID", "contract_id": "C-1", "issuer": "EvraTest",
					"valid": true, "whitelist": "ALLOWED", "last_updated": "2026-09-23T00:00:00Z",
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetTokens(context.Background(), "TOKEN_B_1", 0, 50)
	if err != nil {
		t.Fatalf("GetTokens returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].UID != "TOK-1" {
		t.Fatalf("unexpected tokens data: %+v", resp.Data)
	}
}

func TestAuthorizeToken_Allowed(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tokens/AR/EMS/TOK-1/authorize" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{"allowed": "ALLOWED"},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.AuthorizeToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "TOK-1", nil)
	if err != nil {
		t.Fatalf("AuthorizeToken returned error: %v", err)
	}
	if resp.Data.Allowed != "ALLOWED" {
		t.Fatalf("expected allowed=ALLOWED, got %q", resp.Data.Allowed)
	}
}

func TestAuthorizeToken_NeverReturnsBusinessError(t *testing.T) {
	// Per spec: AuthorizeToken never surfaces a business error — any
	// internal failure (unknown token, disconnected eMSP, timeout)
	// resolves to {"allowed": "BLOCKED"} with a 200/1000 envelope, not an
	// *OcpiError.
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{"allowed": "BLOCKED"},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.AuthorizeToken(context.Background(), "TOKEN_B_1", "AR", "EMS", "UNKNOWN-TOK", nil)
	if err != nil {
		t.Fatalf("AuthorizeToken returned error: %v", err)
	}
	if resp.Data.Allowed != "BLOCKED" {
		t.Fatalf("expected allowed=BLOCKED, got %q", resp.Data.Allowed)
	}
}
