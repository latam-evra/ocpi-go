# Changelog

Todas las versiones notables de `github.com/latam-evra/ocpi-go` se documentan
en este archivo. El formato sigue aproximadamente
[Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/).

## [0.3.0] - 2026-09-23

### Agregado

- Implementación real del módulo **Hub Client Info**
  (`ocpi/hub_client_info.go`): `ListHubClientInfo`, `GetHubClientInfo`,
  con tipo `HubClientInfoEntry` que mapea `toEntry()`/`STATUS_MAP` en
  `lib/ocpi/hubClientInfo.ts` del Hub (`PLANNED`/`CONNECTED`/
  `SUSPENDED`/`STOPPED`).
- Suite de tests de integración
  (`ocpi/hub_client_info_integration_test.go`, build tag `integration`)
  para el módulo, reutilizando los helpers de `integration_test.go`.
- `ocpi/stubs.go`: se retiró el stub `GetHubClientInfo` y el tipo
  `HubClientInfo` (simplificado, sin `role`/`last_updated`) — reemplazado
  por la implementación real.

## [0.2.0] - 2026-09-23

### Agregado

- Implementación real del módulo **Locations** (`ocpi/locations.go`):
  `GetLocations`, `GetLocation`, `PutLocation`, `PatchLocation`, con tipos
  `Location`, `LocationInput`, `EVSE`, `EvseInput`, `Connector`,
  `ConnectorInput` que mapean el JSON público emitido por
  `toPublicLocation()` en `lib/ocpi/locations.ts` del Hub.
- Implementación real del módulo **Tariffs** (`ocpi/tariffs.go`):
  `GetTariffs`, `GetTariff`, `PutTariff`, `DeleteTariff`, con tipos
  `Tariff`, `TariffInput`, `TariffElement`, `PriceComponent` que mapean
  `toPublicTariff()` en `lib/ocpi/tariffs.ts` del Hub.
- Suite de tests de integración (`ocpi/integration_test.go`, build tag
  `integration`) que ejercita Credentials + Locations + Tariffs contra un
  Hub real corriendo localmente, sin mocks — mismo patrón que
  `sdks/node/tests/integration/` y `sdks/python/tests/integration/`.
- `Coordinates`, `Tariff`, `PriceComponent` y `TariffElement` se movieron
  desde `ocpi/stubs.go` a los nuevos `ocpi/locations.go`/`ocpi/tariffs.go`
  (mismo paquete, sin cambio de import para consumidores).

### Cambiado

- `Location` y `Tariff` ahora reflejan la forma real de la respuesta del
  Hub (antes eran una versión simplificada, escrita antes de que el Hub
  implementara estos módulos server-side).

## [0.1.0] - 2026-09-22

### Agregado

- Cliente `ocpi.Client` con soporte real para el módulo **Credentials &
  Registration** (el único implementado server-side en el Hub hoy):
  - `GetVersions(ctx)`
  - `GetDetails(ctx)`
  - `RegisterCredentials(ctx, tokenA, url, roles)` (POST)
  - `RenewCredentials(ctx, tokenB)` (PUT)
  - `TerminateCredentials(ctx, tokenB)` (DELETE)
- Tipos genéricos `ocpi.Envelope[T]` para el sobre de respuesta OCPI, y
  constantes de `status_code` (`StatusSuccess`, `StatusClientError`,
  `StatusInvalidParameters`, `StatusNotEnoughInfo`, `StatusUnknownToken`,
  `StatusServerError`, `StatusUnableToUseAPI`, `StatusUnsupportedVersion`),
  replicando `lib/ocpi/response.ts` del Hub.
- `ocpi.OcpiError`, que implementa `error` y expone `StatusCode` /
  `StatusMessage` / `HTTPStatus` cuando el Hub responde con un error dentro
  del sobre OCPI.
- Stubs tipados para el resto de los módulos del roadmap OCPI 2.3.0:
  Locations, Sessions, CDRs, Tariffs, Tokens & Authorisation, Commands,
  Hub Client Info, Invoice Reconciliation y Charging Profiles. Cada método
  devuelve `ocpi.ErrNotImplemented` con un mensaje que referencia el
  roadmap, hasta que el Hub los implemente server-side.
- Tests con `testing` + `httptest.NewServer` cubriendo éxito y error para
  Credentials, y verificación de `ErrNotImplemented` para cada stub.
- `README.md` con instrucciones de instalación, ejemplo end-to-end del
  handshake de Credentials, tabla de módulos disponibles vs roadmap, y
  nota de uso local vía `replace` en `go.mod` mientras el módulo no esté
  publicado/taggeado en GitHub.
