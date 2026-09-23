package ocpi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPutLocation_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/locations/CL/TST/LOC-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body LocationInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ID != "LOC-1" || body.Coordinates.Latitude != "-33.4" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "CL",
				"party_id":     "TST",
				"id":           "LOC-1",
				"publish":      true,
				"address":      "Av. Test 123",
				"city":         "Santiago",
				"country":      "CHL",
				"coordinates":  map[string]string{"latitude": "-33.4", "longitude": "-70.6"},
				"evses":        []any{},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PutLocation(context.Background(), "TOKEN_B_1", "CL", "TST", "LOC-1", LocationInput{
		ID:          "LOC-1",
		Publish:     true,
		Address:     "Av. Test 123",
		City:        "Santiago",
		Country:     "CHL",
		Coordinates: Coordinates{Latitude: "-33.4", Longitude: "-70.6"},
	})
	if err != nil {
		t.Fatalf("PutLocation returned error: %v", err)
	}
	if resp.Data.ID != "LOC-1" || resp.Data.CountryCode != "CL" || resp.Data.PartyID != "TST" {
		t.Fatalf("unexpected location data: %+v", resp.Data)
	}
}

func TestGetLocation_NotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":           map[string]any{},
			"status_code":    StatusInvalidParameters,
			"status_message": "Location no encontrada.",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	_, err := client.GetLocation(context.Background(), "TOKEN_B_1", "CL", "TST", "DOES-NOT-EXIST")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ocpiErr *OcpiError
	if !AsOcpiError(err, &ocpiErr) {
		t.Fatalf("expected *OcpiError, got %T: %v", err, err)
	}
	if ocpiErr.HTTPStatus != http.StatusNotFound {
		t.Fatalf("expected http status %d, got %d", http.StatusNotFound, ocpiErr.HTTPStatus)
	}
}

func TestGetLocations_Pagination(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locations" || r.URL.RawQuery != "offset=0&limit=10" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"country_code": "CL",
					"party_id":     "TST",
					"id":           "LOC-1",
					"publish":      true,
					"address":      "x",
					"city":         "x",
					"country":      "CHL",
					"coordinates":  map[string]string{"latitude": "0", "longitude": "0"},
					"evses":        []any{},
				},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.GetLocations(context.Background(), "TOKEN_B_1", 0, 10)
	if err != nil {
		t.Fatalf("GetLocations returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "LOC-1" {
		t.Fatalf("unexpected locations data: %+v", resp.Data)
	}
}

func TestPatchLocation_Success(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/locations/CL/TST/LOC-1" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["city"] != "Valparaíso" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"country_code": "CL",
				"party_id":     "TST",
				"id":           "LOC-1",
				"publish":      true,
				"address":      "x",
				"city":         "Valparaíso",
				"country":      "CHL",
				"coordinates":  map[string]string{"latitude": "0", "longitude": "0"},
				"evses":        []any{},
			},
			"status_code":    StatusSuccess,
			"status_message": "Success",
			"timestamp":      "2026-09-23T00:00:00Z",
		})
	})
	defer closeFn()

	resp, err := client.PatchLocation(context.Background(), "TOKEN_B_1", "CL", "TST", "LOC-1", map[string]any{
		"city": "Valparaíso",
	})
	if err != nil {
		t.Fatalf("PatchLocation returned error: %v", err)
	}
	if resp.Data.City != "Valparaíso" {
		t.Fatalf("unexpected city: %q", resp.Data.City)
	}
}
