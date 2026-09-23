package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func cdrTestBody() CdrInput {
	return CdrInput{
		ID:            "CDR-1",
		StartDateTime: "2026-09-23T00:00:00Z",
		EndDateTime:   "2026-09-23T01:00:00Z",
		CdrToken:      CdrToken{CountryCode: "AR", PartyID: "EMS", UID: "TOK-1", Type: TokenTypeRFID, ContractID: "C-1"},
		AuthMethod:    AuthMethodAuthRequest,
		CdrLocation: CdrLocation{
			ID: "LOC-1", Address: "x", City: "x", Country: "CHL",
			Coordinates: Coordinates{Latitude: "0", Longitude: "0"},
			EVSEUID:     "EVSE-1", ConnectorID: "1", ConnectorStandard: "IEC_62196_T2",
			ConnectorFormat: "SOCKET", ConnectorPowerType: "AC_3_PHASE",
		},
		Currency: "USD",
		ChargingPeriods: []ChargingPeriod{
			{StartDateTime: "2026-09-23T00:00:00Z", Dimensions: []ChargingPeriodDimension{{Type: "ENERGY", Volume: 10}}},
		},
		TotalCost:   CostAmount{ExclVat: 5},
		TotalEnergy: 10,
		TotalTime:   1,
	}
}

func cdrTestResponseData() map[string]any {
	return map[string]any{
		"country_code":    "CL",
		"party_id":        "TST",
		"id":              "CDR-1",
		"start_date_time": "2026-09-23T00:00:00Z",
		"end_date_time":   "2026-09-23T01:00:00Z",
		"cdr_token": map[string]any{
			"country_code": "AR", "party_id": "EMS", "uid": "TOK-1", "type": "RFID", "contract_id": "C-1",
		},
		"auth_method": "AUTH_REQUEST",
		"cdr_location": map[string]any{
			"id": "LOC-1", "address": "x", "city": "x", "country": "CHL",
			"coordinates": map[string]string{"latitude": "0", "longitude": "0"},
			"evse_uid":    "EVSE-1", "connector_id": "1", "connector_standard": "IEC_62196_T2",
			"connector_format": "SOCKET", "connector_power_type": "AC_3_PHASE",
		},
		"currency": "USD",
		"charging_periods": []map[string]any{
			{"start_date_time": "2026-09-23T00:00:00Z", "dimensions": []map[string]any{{"type": "ENERGY", "volume": 10}}},
		},
		"total_cost":   map[string]any{"excl_vat": 5},
		"total_energy": 10,
		"total_time":   1,
		"last_updated": "2026-09-23T00:00:00Z",
	}
}

func TestPostCdr_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/cdrs/CL/TST/CDR-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body CdrInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ID != "CDR-1" || body.TotalEnergy != 10 {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           cdrTestResponseData(),
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PostCdr(context.Background(), "TOKEN_B_1", "CL", "TST", "CDR-1", cdrTestBody())
	if err != nil {
		t.Fatalf("PostCdr returned error: %v", err)
	}
	if resp.Data.ID != "CDR-1" || resp.Data.CountryCode != "CL" {
		t.Fatalf("unexpected cdr data: %+v", resp.Data)
	}
}

func TestPostCdr_DuplicateReturnsConflict(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Este CDR ya fue emitido y es inmutable.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.PostCdr(context.Background(), "TOKEN_B_1", "CL", "TST", "CDR-1", cdrTestBody())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.HTTPStatus != http.StatusConflict {
		t.Fatalf("expected http status %d, got %d", http.StatusConflict, ocpiErr.HTTPStatus)
	}
}

func TestGetCdr_NotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "CDR no encontrado.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.GetCdr(context.Background(), "TOKEN_B_1", "CL", "TST", "DOES-NOT-EXIST")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
}

func TestGetCdrs_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cdrs" || r.URL.RawQuery != "offset=0&limit=10" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           []map[string]any{cdrTestResponseData()},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetCdrs(context.Background(), "TOKEN_B_1", 0, 10)
	if err != nil {
		t.Fatalf("GetCdrs returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "CDR-1" {
		t.Fatalf("unexpected cdrs data: %+v", resp.Data)
	}
}
