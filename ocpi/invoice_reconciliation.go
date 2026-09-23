package ocpi

import (
	"context"
	"net/http"
)

// This file implements the Invoice Reconciliation module
// (mod_invoicereconciliations, OCPI 2.3.0 Edition 2), which the Hub
// implements server-side. Field shapes mirror the public JSON emitted by
// toPublicInvoiceReconciliation() in lib/ocpi/invoiceReconciliation.ts on
// the Hub. Unlike CDRs (POST-only) this module is PUT-only (upsert), with
// no POST — the two shouldn't be confused despite looking similar.

// InvoiceReconciliationStatus is the reconciliation status of a record.
type InvoiceReconciliationStatus string

const (
	InvoiceReconciliationMatched  InvoiceReconciliationStatus = "MATCHED"
	InvoiceReconciliationDisputed InvoiceReconciliationStatus = "DISPUTED"
	InvoiceReconciliationPending  InvoiceReconciliationStatus = "PENDING"
	InvoiceReconciliationResolved InvoiceReconciliationStatus = "RESOLVED"
)

// DiscrepancyAmount is the disputed amount in the CDR's original currency.
type DiscrepancyAmount struct {
	ExclVat float64  `json:"excl_vat"`
	InclVat *float64 `json:"incl_vat,omitempty"`
}

// InvoiceReconciliationInput is the request body for PUT
// /invoicereconciliations/{country_code}/{party_id}/{reconciliation_id}.
type InvoiceReconciliationInput struct {
	ID                     string                      `json:"id"`
	CdrID                  string                      `json:"cdr_id"`
	InvoiceReferenceID     string                      `json:"invoice_reference_id,omitempty"`
	Status                 InvoiceReconciliationStatus `json:"status"`
	DiscrepancyDescription string                      `json:"discrepancy_description,omitempty"`
	DiscrepancyAmount      *DiscrepancyAmount          `json:"discrepancy_amount,omitempty"`
}

// InvoiceReconciliation matches a CDR against an invoice reference for
// financial reconciliation, as returned by the Hub. If the input included
// DiscrepancyAmount, the Hub resolves the CDR's currency, converts it to
// USD via Frankfurter, and populates DiscrepancyCurrency /
// DiscrepancyAmountUSD / ExchangeRateUsed server-side — the SDK only types
// the result, it does not replicate the FX calculation. When the input did
// NOT include DiscrepancyAmount, those 3 fields are absent (nil pointers /
// zero values here), never assume they're always present.
type InvoiceReconciliation struct {
	CountryCode            string                      `json:"country_code"`
	PartyID                string                      `json:"party_id"`
	ID                     string                      `json:"id"`
	CdrID                  string                      `json:"cdr_id"`
	InvoiceReferenceID     string                      `json:"invoice_reference_id,omitempty"`
	Status                 InvoiceReconciliationStatus `json:"status"`
	DiscrepancyDescription string                      `json:"discrepancy_description,omitempty"`
	DiscrepancyAmount      *DiscrepancyAmount          `json:"discrepancy_amount,omitempty"`
	DiscrepancyCurrency    string                      `json:"discrepancy_currency,omitempty"`
	DiscrepancyAmountUSD   *DiscrepancyAmount          `json:"discrepancy_amount_usd,omitempty"`
	ExchangeRateUsed       *float64                    `json:"exchange_rate_used,omitempty"`
	LastUpdated            string                      `json:"last_updated"`
}

// InvoiceReconciliationsResponse is the full envelope returned by
// GET /invoicereconciliations.
type InvoiceReconciliationsResponse = Envelope[[]InvoiceReconciliation]

// InvoiceReconciliationResponse is the full envelope returned by
// GET/PUT /invoicereconciliations/{country_code}/{party_id}/{reconciliation_id}.
type InvoiceReconciliationResponse = Envelope[InvoiceReconciliation]

// GetInvoiceReconciliations calls GET /invoicereconciliations and lists the
// records visible to the caller.
func (c *Client) GetInvoiceReconciliations(ctx context.Context, tokenB string, offset, limit int) (*InvoiceReconciliationsResponse, error) {
	url := c.baseURL + "/invoicereconciliations" + paginationQuery(offset, limit)
	return doRequest[[]InvoiceReconciliation](ctx, c, http.MethodGet, url, tokenB, nil)
}

// GetInvoiceReconciliation calls GET
// /invoicereconciliations/{country_code}/{party_id}/{reconciliation_id}
// and fetches a single record by ID.
func (c *Client) GetInvoiceReconciliation(ctx context.Context, tokenB, countryCode, partyID, reconciliationID string) (*InvoiceReconciliationResponse, error) {
	url := c.baseURL + "/invoicereconciliations/" + countryCode + "/" + partyID + "/" + reconciliationID
	return doRequest[InvoiceReconciliation](ctx, c, http.MethodGet, url, tokenB, nil)
}

// PutInvoiceReconciliation calls PUT
// /invoicereconciliations/{country_code}/{party_id}/{reconciliation_id} to
// create or replace a record (upsert). Requires a CPO or EMSP role
// matching countryCode/partyID (unlike CDRs/Tokens, which require a single
// fixed role). If body.CdrID doesn't resolve to an existing Cdr, or the
// Cdr's currency has no FX provider (supported: ARS/CLP/MXN/BRL), the Hub
// returns a 400 *OcpiError, which this method lets propagate unmodified.
func (c *Client) PutInvoiceReconciliation(ctx context.Context, tokenB, countryCode, partyID, reconciliationID string, body InvoiceReconciliationInput) (*InvoiceReconciliationResponse, error) {
	url := c.baseURL + "/invoicereconciliations/" + countryCode + "/" + partyID + "/" + reconciliationID
	return doRequest[InvoiceReconciliation](ctx, c, http.MethodPut, url, tokenB, body)
}

// DeleteInvoiceReconciliation calls DELETE
// /invoicereconciliations/{country_code}/{party_id}/{reconciliation_id} to
// remove a record.
func (c *Client) DeleteInvoiceReconciliation(ctx context.Context, tokenB, countryCode, partyID, reconciliationID string) error {
	url := c.baseURL + "/invoicereconciliations/" + countryCode + "/" + partyID + "/" + reconciliationID
	_, err := doRequest[map[string]any](ctx, c, http.MethodDelete, url, tokenB, nil)
	return err
}
