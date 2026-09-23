//go:build integration

package ocpi

import (
	"context"
	"testing"
)

// Mirrors sdks/node/tests/integration/invoice-reconciliation.integration.test.ts
// and sdks/python/tests/integration/test_invoice_reconciliation_integration.py.

func TestIntegration_InvoiceReconciliation_AgainstRealHub(t *testing.T) {
	ctx := context.Background()
	reg := createTestRegistration(t, "CPO", "CL", "IR1")
	defer cleanupTestRegistration(t, reg.RegistrationID)

	client := NewClient(WithBaseURL(hubTestURL() + "/api/ocpi/2.3.0"))

	credsResp, err := client.RegisterCredentials(ctx, reg.RawTokenA, hubTestURL()+"/api/ocpi/2.3.0/versions", []Role{
		{Role: RoleCPO, PartyID: "IR1", CountryCode: "CL"},
	})
	if err != nil {
		t.Fatalf("RegisterCredentials failed: %v", err)
	}
	tokenB := credsResp.Data.Token

	// CDR in CLP (a currency supported by Frankfurter via BCCh) for the FX
	// conversion case below.
	cdrInput := CdrInput{
		ID:            "CDR-IR-1",
		StartDateTime: "2026-09-23T09:00:00Z",
		EndDateTime:   "2026-09-23T10:00:00Z",
		CdrToken: CdrToken{
			CountryCode: "AR", PartyID: "EVP", UID: "TOK-IR-1", Type: TokenTypeRFID, ContractID: "C-IR-1",
		},
		AuthMethod: AuthMethodWhitelist,
		CdrLocation: CdrLocation{
			ID:          "LOC-IR-1",
			Address:     "Av. Test 123",
			City:        "Santiago",
			Country:     "CHL",
			Coordinates: Coordinates{Latitude: "-33.4", Longitude: "-70.6"},
			EVSEUID:     "EVSE-1", ConnectorID: "1", ConnectorStandard: "IEC_62196_T2",
			ConnectorFormat: "SOCKET", ConnectorPowerType: "AC_3_PHASE",
		},
		Currency: "CLP",
		ChargingPeriods: []ChargingPeriod{
			{StartDateTime: "2026-09-23T09:00:00Z", Dimensions: []ChargingPeriodDimension{{Type: "ENERGY", Volume: 10}}},
		},
		TotalCost:   CostAmount{ExclVat: 5000},
		TotalEnergy: 10,
		TotalTime:   1,
	}
	if _, err := client.PostCdr(ctx, tokenB, "CL", "IR1", "CDR-IR-1", cdrInput); err != nil {
		t.Fatalf("PostCdr (fixture) failed: %v", err)
	}

	t.Run("without discrepancy_amount: FX fields absent, full CRUD", func(t *testing.T) {
		created, err := client.PutInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-1", InvoiceReconciliationInput{
			ID: "REC-SDK-1", CdrID: "CDR-IR-1", Status: InvoiceReconciliationPending,
		})
		if err != nil {
			t.Fatalf("PutInvoiceReconciliation failed: %v", err)
		}
		if created.Data.ID != "REC-SDK-1" {
			t.Fatalf("expected id REC-SDK-1, got %q", created.Data.ID)
		}
		if created.Data.DiscrepancyCurrency != "" || created.Data.DiscrepancyAmountUSD != nil || created.Data.ExchangeRateUsed != nil {
			t.Fatalf("expected FX fields absent, got %+v", created.Data)
		}

		fetched, err := client.GetInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-1")
		if err != nil {
			t.Fatalf("GetInvoiceReconciliation failed: %v", err)
		}
		if fetched.Data.Status != InvoiceReconciliationPending {
			t.Fatalf("expected status PENDING, got %q", fetched.Data.Status)
		}

		list, err := client.GetInvoiceReconciliations(ctx, tokenB, 0, 100)
		if err != nil {
			t.Fatalf("GetInvoiceReconciliations failed: %v", err)
		}
		found := false
		for _, rec := range list.Data {
			if rec.ID == "REC-SDK-1" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected REC-SDK-1 in reconciliations list, got %+v", list.Data)
		}

		if err := client.DeleteInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-1"); err != nil {
			t.Fatalf("DeleteInvoiceReconciliation failed: %v", err)
		}
		if _, err := client.GetInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-1"); err == nil {
			t.Fatal("expected GetInvoiceReconciliation to fail after delete")
		}
	})

	t.Run("with discrepancy_amount: real FX conversion to USD (CLP via Frankfurter/BCCh)", func(t *testing.T) {
		created, err := client.PutInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-2", InvoiceReconciliationInput{
			ID: "REC-SDK-2", CdrID: "CDR-IR-1", Status: InvoiceReconciliationDisputed,
			DiscrepancyDescription: "Total cost mismatch",
			DiscrepancyAmount:      &DiscrepancyAmount{ExclVat: 1000},
		})
		if err != nil {
			t.Fatalf("PutInvoiceReconciliation (with FX) failed: %v", err)
		}
		if created.Data.DiscrepancyCurrency != "CLP" {
			t.Fatalf("expected discrepancy_currency CLP, got %q", created.Data.DiscrepancyCurrency)
		}
		if created.Data.DiscrepancyAmountUSD == nil || created.Data.DiscrepancyAmountUSD.ExclVat <= 0 {
			t.Fatalf("expected a positive discrepancy_amount_usd, got %+v", created.Data.DiscrepancyAmountUSD)
		}
		if created.Data.ExchangeRateUsed == nil || *created.Data.ExchangeRateUsed <= 0 {
			t.Fatalf("expected a positive exchange_rate_used, got %v", created.Data.ExchangeRateUsed)
		}

		if err := client.DeleteInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-2"); err != nil {
			t.Fatalf("DeleteInvoiceReconciliation (cleanup) failed: %v", err)
		}
	})

	t.Run("cdr_id that doesn't resolve returns 400", func(t *testing.T) {
		_, err := client.PutInvoiceReconciliation(ctx, tokenB, "CL", "IR1", "REC-SDK-3", InvoiceReconciliationInput{
			ID: "REC-SDK-3", CdrID: "DOES-NOT-EXIST", Status: InvoiceReconciliationPending,
			DiscrepancyAmount: &DiscrepancyAmount{ExclVat: 10},
		})
		if err == nil {
			t.Fatal("expected an error for an unresolved cdr_id")
		}
		var ocpiErr *OcpiError
		if !AsOcpiError(err, &ocpiErr) {
			t.Fatalf("expected *OcpiError, got %T: %v", err, err)
		}
	})
}
