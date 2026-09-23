# ocpi-go

Cliente Go ligero para el Hub de roaming OCPI 2.3.0 de **LATAM EV Roaming
Alliance**. Usa exclusivamente la librería estándar de Go (`net/http`,
`encoding/json`) — sin dependencias externas.

> **Estado actual del Hub**: el Hub implementa realmente todos los
> módulos del roadmap OCPI 2.3.0, y este SDK expone un cliente completo
> y funcional para cada uno de ellos.

## Instalación

```bash
go get github.com/latam-evra/ocpi-go
```

### Nota: módulo aún no publicado/taggeado

Este módulo todavía no fue publicado con un tag en el repositorio público de
GitHub (`github.com/latam-evra/ocpi-go`), por lo que `go get` puede fallar
hasta que exista al menos un tag `v0.1.0`. Mientras tanto, para usarlo
localmente desde otro proyecto Go, cloná o copiá el directorio `sdks/go/` y
agregá una directiva `replace` en el `go.mod` del consumidor:

```
require github.com/latam-evra/ocpi-go v0.2.0

replace github.com/latam-evra/ocpi-go => ../ruta/a/latam-evra.org/sdks/go
```

(ajustá la ruta relativa/absoluta según dónde hayas clonado este repo).

## Uso: handshake de Credentials completo

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/latam-evra/ocpi-go/ocpi"
)

func main() {
	ctx := context.Background()

	// Apuntá al Hub de producción (default) o a un entorno local/staging:
	client := ocpi.NewClient(
	// ocpi.WithBaseURL("http://localhost:3000/api/ocpi/2.3.0"),
	)

	// 1. Descubrí las versiones OCPI que soporta el Hub.
	versions, err := client.GetVersions(ctx)
	if err != nil {
		log.Fatalf("GetVersions: %v", err)
	}
	fmt.Printf("Versiones soportadas: %+v\n", versions.Data)

	// 2. Registrá tus credenciales (TOKEN_A provisto por el Hub de antemano,
	//    y la URL de tu propio endpoint /versions para que el Hub te
	//    contacte y valide OCPI 2.3.0).
	roles := []ocpi.Role{
		{Role: ocpi.RoleCPO, PartyID: "CHG", CountryCode: "CL"},
	}
	creds, err := client.RegisterCredentials(
		ctx,
		"TOKEN_A_TEMPORAL",
		"https://tu-csms.example.com/ocpi/versions",
		roles,
	)
	if err != nil {
		var ocpiErr *ocpi.OcpiError
		if ocpi.AsOcpiError(err, &ocpiErr) {
			log.Fatalf("Handshake rechazado: status_code=%d message=%q",
				ocpiErr.StatusCode, ocpiErr.StatusMessage)
		}
		log.Fatalf("RegisterCredentials: %v", err)
	}

	tokenB := creds.Data.Token
	fmt.Printf("Handshake OK. TOKEN_B recibido: %s\n", tokenB)

	// 3. Guardá el TOKEN_B de forma segura. Más adelante podés renovarlo...
	renewed, err := client.RenewCredentials(ctx, tokenB)
	if err != nil {
		log.Fatalf("RenewCredentials: %v", err)
	}
	tokenB = renewed.Data.Token

	// 4. ...o terminar la conexión.
	if err := client.TerminateCredentials(ctx, tokenB); err != nil {
		log.Fatalf("TerminateCredentials: %v", err)
	}
}
```

## Uso: Locations y Tariffs

```go
// Publicar una Location (requiere rol CPO en country_code/party_id).
loc, err := client.PutLocation(ctx, tokenB, "CL", "CHG", "LOC-1", ocpi.LocationInput{
	ID:      "LOC-1",
	Publish: true,
	Address: "Av. Andrés Bello 2425",
	City:    "Santiago",
	Country: "CHL",
	Coordinates: ocpi.Coordinates{Latitude: "-33.4182", Longitude: "-70.6061"},
})

// Actualización parcial.
patched, err := client.PatchLocation(ctx, tokenB, "CL", "CHG", "LOC-1", map[string]any{
	"city": "Valparaíso",
})

// Listar Locations publicadas, y obtener una por id.
page, err := client.GetLocations(ctx, tokenB, 0, 50)
one, err := client.GetLocation(ctx, tokenB, "CL", "CHG", "LOC-1")

// Tariffs: put, get, delete.
tariff, err := client.PutTariff(ctx, tokenB, "CL", "CHG", "TAR-1", ocpi.TariffInput{
	ID:       "TAR-1",
	Currency: "USD",
	Elements: []ocpi.TariffElement{
		{PriceComponents: []ocpi.PriceComponent{{Type: "ENERGY", Price: 0.35, StepSize: 1}}},
	},
})
err = client.DeleteTariff(ctx, tokenB, "CL", "CHG", "TAR-1")

// Hub Client Info: visibilidad de qué parties están conectadas al Hub.
list, err := client.ListHubClientInfo(ctx, tokenB, 0, 50)
byRole, err := client.GetHubClientInfo(ctx, tokenB, "CL", "CHG")
```

### Manejo de errores OCPI

El Hub responde con `HTTP 200` (o el código HTTP que corresponda) y un
sobre con `status_code`/`status_message` siguiendo la convención OCPI
(1000 = éxito, 2xxx = error de cliente, 3xxx = error de servidor). El SDK
traduce automáticamente cualquier `status_code` de error en un
`*ocpi.OcpiError`:

```go
var ocpiErr *ocpi.OcpiError
if ocpi.AsOcpiError(err, &ocpiErr) {
	switch ocpiErr.StatusCode {
	case ocpi.StatusUnknownToken:
		// TOKEN_A/TOKEN_B inválido o revocado.
	case ocpi.StatusUnsupportedVersion:
		// El CSMS remoto no soporta OCPI 2.3.0.
	}
}
```

## Módulos disponibles

| Módulo | Métodos del SDK |
|---|---|
| Credentials & Registration | `GetVersions`, `GetDetails`, `RegisterCredentials`, `RenewCredentials`, `TerminateCredentials` |
| Locations | `GetLocations`, `GetLocation`, `PutLocation`, `PatchLocation` |
| Tariffs | `GetTariffs`, `GetTariff`, `PutTariff`, `DeleteTariff` |
| Hub Client Info | `ListHubClientInfo`, `GetHubClientInfo` |
| Sessions | `GetSessions`, `GetSession`, `PutSession`, `PatchSession` |
| CDRs | `GetCdrs`, `GetCdr`, `PostCdr` |
| Tokens & Authorisation | `GetTokens`, `GetToken`, `PutToken`, `PatchToken`, `DeleteToken`, `AuthorizeToken` |
| Commands | `StartSession`, `ReserveNow`, `StopSession`, `UnlockConnector`, `CancelReservation`, `GetCommand` |
| Charging Profiles | `GetActiveChargingProfile`, `SetChargingProfile`, `DeleteChargingProfile`, `GetChargingProfile` |
| Invoice Reconciliation (Ed. 2) | `GetInvoiceReconciliations`, `GetInvoiceReconciliation`, `PutInvoiceReconciliation`, `DeleteInvoiceReconciliation` |

Consultá `docs/Roaming_hub_Latam.md` y `components/ModuleAccordion.tsx` en
el repositorio del Hub para el detalle de cada módulo.

## Desarrollo

```bash
cd sdks/go
go build ./...
go test ./...
```

## Tests de integración

Requieren el Hub real corriendo. Están separados de la suite normal con el
build tag `integration` porque hacen llamadas HTTP reales y disparan un
subproceso (`npx tsx scripts/create-test-registration.ts`) contra la base de
datos del Hub — nunca corren en `go test ./...` sin el tag.

El handshake de Credentials simula un CSMS corriendo en `localhost`, lo cual
la protección SSRF del Hub bloquea por diseño. Por eso el Hub usado para
estos tests necesita levantarse con `OCPI_ALLOW_LOOPBACK=true`, una env var
de solo desarrollo que **nunca debe activarse en producción**. No reutilices
el server de producción de pm2 — levantá una instancia dev dedicada desde la
raíz del repo:

```bash
OCPI_ALLOW_LOOPBACK=true PORT=3948 npm run dev
```

Y corré la suite apuntando a esa instancia:

```bash
OCPI_HUB_TEST_URL=http://localhost:3948 go test -tags=integration ./...
```

Por defecto (sin `OCPI_HUB_TEST_URL`) apuntan a `http://localhost:3947` —
solo válido si esa instancia también tiene `OCPI_ALLOW_LOOPBACK=true`.

## Licencia

Ver licencia del repositorio principal `latam-evra.org`.
