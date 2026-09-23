package ocpi

import (
	"errors"
	"fmt"
)

// ErrNotImplemented is returned by every stub method for OCPI modules that
// the Hub does not yet implement server-side (everything except
// Credentials & Registration). See docs/Roaming_hub_Latam.md in the Hub
// repository for the module roadmap.
var ErrNotImplemented = errors.New("ocpi-go: module not implemented by the Hub yet")

// OcpiError represents an OCPI-level error: the HTTP call succeeded (usually
// with a 200 OK, per the Hub's convention in lib/ocpi/response.ts) but the
// envelope's status_code indicates a client (2xxx) or server (3xxx) error.
type OcpiError struct {
	// StatusCode is the OCPI status_code from the envelope (e.g. 2003 for
	// UNKNOWN_TOKEN). See the Status* constants in types.go.
	StatusCode int
	// StatusMessage is the human-readable status_message from the envelope.
	StatusMessage string
	// HTTPStatus is the actual HTTP status code of the response, when known.
	HTTPStatus int
}

func (e *OcpiError) Error() string {
	return fmt.Sprintf("ocpi: request failed with status_code=%d (%s) [http %d]",
		e.StatusCode, e.StatusMessage, e.HTTPStatus)
}

// IsNotImplemented reports whether err is (or wraps) ErrNotImplemented.
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented)
}

// AsOcpiError is a thin convenience wrapper around errors.As for
// extracting an *OcpiError from an error chain, e.g.:
//
//	var ocpiErr *ocpi.OcpiError
//	if ocpi.AsOcpiError(err, &ocpiErr) {
//	    fmt.Println(ocpiErr.StatusCode)
//	}
func AsOcpiError(err error, target **OcpiError) bool {
	return errors.As(err, target)
}

// notImplementedError builds a clear, module-specific ErrNotImplemented
// wrapper.
func notImplementedError(module string) error {
	return fmt.Errorf("ocpi-go: el módulo %s aún no está disponible en el Hub — ver roadmap en docs/Roaming_hub_Latam.md: %w", module, ErrNotImplemented)
}
