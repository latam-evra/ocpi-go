# ocpi-go

Cliente Go ligero para el Hub de roaming OCPI 2.3.0 de **LATAM EV Roaming
Alliance**. Usa exclusivamente la librería estándar de Go (`net/http`,
`encoding/json`) — sin dependencias externas.

> **Estado actual del Hub**: hoy el Hub solo implementa realmente el módulo
> **Credentials & Registration**. Este SDK expone un cliente completo y
> funcional para ese módulo, y stubs tipados para el resto de los módulos
> del roadmap OCPI (Locations, Sessions, CDRs, Tariffs, Tokens, Commands,
> Hub Client Info, Invoice Reconciliation, Charging Profiles), que hoy
> devuelven `ErrNotImplemented`.

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
require github.com/latam-evra/ocpi-go v0.1.0

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

## Módulos disponibles vs roadmap

| Módulo | Estado | Métodos del SDK |
|---|---|---|
| Credentials & Registration | **Disponible** (implementado en el Hub) | `GetVersions`, `GetDetails`, `RegisterCredentials`, `RenewCredentials`, `TerminateCredentials` |
| Locations | Roadmap | `GetLocations`, `GetLocation` (devuelven `ErrNotImplemented`) |
| Sessions | Roadmap | `GetActiveSession` |
| CDRs | Roadmap | `GetCdrs`, `SubmitCdr` |
| Tariffs | Roadmap | `GetTariffs` |
| Tokens & Authorisation | Roadmap | `AuthorizeToken` |
| Commands | Roadmap | `StartSession`, `StopSession`, `UnlockConnector` |
| Hub Client Info | Roadmap | `GetHubClientInfo` |
| Invoice Reconciliation (Ed. 2) | Roadmap | `GetInvoiceReconciliations` |
| Charging Profiles | Roadmap | `SetChargingProfile` |

Todos los stubs de módulos "Roadmap" ya tienen sus tipos Go completos
(con tags `json:"..."` que reflejan el objeto OCPI real), listos para usar
en cuanto el Hub implemente el módulo server-side. Llamarlos hoy devuelve
`ocpi.ErrNotImplemented` (verificable con `ocpi.IsNotImplemented(err)`), con
un mensaje que referencia el roadmap (`docs/Roaming_hub_Latam.md`).

## Desarrollo

```bash
cd sdks/go
go build ./...
go test ./...
```

## Licencia

Ver licencia del repositorio principal `latam-evra.org`.
