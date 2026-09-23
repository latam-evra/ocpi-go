# Changelog

Todas las versiones notables de `github.com/latam-evra/ocpi-go` se documentan
en este archivo. El formato sigue aproximadamente
[Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/).

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
