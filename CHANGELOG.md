# Changelog

Todas las versiones notables de `github.com/latam-evra/ocpi-go` se documentan
en este archivo. El formato sigue aproximadamente
[Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/).

## [0.5.0] - 2026-09-23

### Agregado

- Implementación real del módulo **Charging Profiles**
  (`ocpi/charging_profiles.go`): `GetActiveChargingProfile`,
  `SetChargingProfile`, `DeleteChargingProfile`, `GetChargingProfile`.
  Al igual que Commands, no es CRUD simétrico: un método por acción
  (`GET_ACTIVE_CHARGING_PROFILE`/`PUT_CHARGING_PROFILE`/
  `DELETE_CHARGING_PROFILE`) sobre una sesión existente, más
  `GetChargingProfile` cuyo `GET` vive en
  `/chargingprofiles/callback/{id}`, no en
  `/chargingprofiles/{session_id}`.
- Tests unitarios (`ocpi/charging_profiles_test.go`) y de integración
  (build tag `integration`: `ocpi/charging_profiles_integration_test.go`,
  con un `httptest.NewServer` actuando de CPO externo).
- `ocpi/stubs.go` se elimina: con Charging Profiles implementado, ya no
  queda ningún módulo en el roadmap del Hub. `ErrNotImplemented`,
  `IsNotImplemented` y `notImplementedError` se retiran de
  `ocpi/errors.go` por no tener ya ningún uso.

### Breaking changes

- `SetChargingProfile` cambia de firma: antes tomaba
  `(ctx, sessionID, ChargingProfile)` y siempre devolvía
  `ErrNotImplemented`; ahora toma
  `(ctx, tokenB, countryCode, partyID, sessionID, responseURL, ChargingProfile)`
  y devuelve `(*ChargingProfileAckResponse, error)` contra el Hub real.
- `ChargingProfile.Limit` (un único límite) se reemplaza por
  `ChargingProfilePeriod` (lista de `{start_period, limit}`), reflejando
  el objeto `charging_profile` real de OCPI 2.3.0.

## [0.4.0] - 2026-09-23

### Agregado

- Implementación real de 5 módulos: **Sessions** (`ocpi/sessions.go`:
  `GetSessions`, `GetSession`, `PutSession`, `PatchSession`), **CDRs**
  (`ocpi/cdrs.go`: `GetCdrs`, `GetCdr`, `PostCdr` — sin Put/Patch/Delete,
  CDRs son inmutables), **Tokens & Authorisation** (`ocpi/tokens.go`:
  `GetTokens`, `GetToken`, `PutToken`, `PatchToken`, `DeleteToken`, más
  `AuthorizeToken` como método propio, no CRUD, que hace POST a
  `/tokens/{country_code}/{party_id}/{token_uid}/authorize` con body JSON
  — nunca surge un error de negocio, cualquier fallo interno resuelve a
  `AuthorizeResult{Allowed: "BLOCKED"}`), **Commands**
  (`ocpi/commands.go`: `StartSession`, `ReserveNow`, `StopSession`,
  `UnlockConnector`, `CancelReservation` — cada uno un `POST
  /commands/{TYPE}` con su struct de payload tipado, más `GetCommand`
  cuyo `GET` vive en `/commands/callback/{command_id}`, no en
  `/commands/{command_type}`) e **Invoice Reconciliation**
  (`ocpi/invoice_reconciliation.go`: `GetInvoiceReconciliations`,
  `GetInvoiceReconciliation`, `PutInvoiceReconciliation` — upsert, sin
  POST — y `DeleteInvoiceReconciliation`; `InvoiceReconciliation` incluye
  los campos opcionales de conversión FX server-side,
  `DiscrepancyCurrency`/`DiscrepancyAmountUSD`/`ExchangeRateUsed`,
  ausentes cuando el `PUT` no incluyó `discrepancy_amount`).
- `CdrToken`, `ChargingPeriodDimension` y `ChargingPeriod` se definen en
  `ocpi/sessions.go` y se reusan desde `ocpi/cdrs.go`; `PriceComponent` y
  `TariffElement` (ya existentes en `ocpi/tariffs.go`) se reusan desde
  `CdrTariff` en `ocpi/cdrs.go` — `TariffElement` ahora también expone
  `Restrictions`, usado por CDRs.
- Tests unitarios por módulo (`ocpi/sessions_test.go`,
  `ocpi/cdrs_test.go`, `ocpi/tokens_test.go`, `ocpi/commands_test.go`,
  `ocpi/invoice_reconciliation_test.go`), mismo patrón que
  `ocpi/locations_test.go` (`httptest.NewServer` mockeando el Hub).
- Tests de integración por módulo (build tag `integration`:
  `ocpi/sessions_integration_test.go`, `ocpi/cdrs_integration_test.go`,
  `ocpi/tokens_integration_test.go`, `ocpi/commands_integration_test.go`
  — con un `httptest.NewServer` actuando de CPO externo, mismo patrón que
  el equivalente Node/Python —,
  `ocpi/invoice_reconciliation_integration_test.go` — incluye un caso de
  conversión FX real vía Frankfurter/BCCh sobre un CDR en CLP).
- `ocpi/stubs.go`: se retiraron los stubs `GetActiveSession`, `GetCdrs`,
  `SubmitCdr`, `AuthorizeToken`, `StartSession`, `StopSession`,
  `UnlockConnector`, `GetInvoiceReconciliations` y los structs
  placeholder correspondientes (`Session`, `Cdr`, `CdrToken`,
  `CdrLocation`, `ChargingPeriodDimension`, `ChargingPeriod`, `Price`,
  `Token`, `CommandTokenRef`, `StartSessionCommand`,
  `InvoiceReconciliation`) — reemplazados por las implementaciones
  reales. `SetChargingProfile`/`ChargingProfile` son el único stub que
  queda (Charging Profiles no está en el roadmap del Hub).

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
