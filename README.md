# keycloak-broker

An [Open Service Broker](https://www.openservicebrokerapi.org/) (OSB) API implementation that wraps the Keycloak Admin REST API, enabling platforms like Cloud Foundry or the Swisscom DevOps portal to provision and manage OIDC clients in Keycloak through the standard OSB interface.
Swisscom developers can leverage this to use OIDC integrations into their applications in a fully self-service mode.

## Architecture

```
main.go → router → Echo HTTP server
                  ├── /health, /healthz        → Keycloak realm connectivity check
                  ├── /metrics                 → Prometheus metrics
                  └── /v2/* (basic auth)       → OSB API
                        ├── GET    /catalog
                        ├── PUT    /service_instances/:id          → Provision
                        ├── PATCH  /service_instances/:id          → Update
                        ├── GET    /service_instances/:id          → Fetch
                        ├── DELETE /service_instances/:id          → Deprovision
                        ├── PUT    /service_instances/:id/service_bindings/:bid  → Bind
                        ├── GET    /service_instances/:id/service_bindings/:bid  → Fetch binding
                        └── DELETE /service_instances/:id/service_bindings/:bid  → Unbind
```

### Packages

| Package | Purpose |
|---------|---------|
| `config` | Singleton configuration from environment variables |
| `logger` | Thin wrapper around `slog` with printf-style helpers |
| `catalog` | Loads `catalog.yaml` at startup, provides service/plan lookups |
| `validation` | UUID format and catalog membership validation |
| `keycloak` | HTTP client for Keycloak Admin API (token management, CRUD, OIDC discovery caching) |
| `broker` | OSB API handlers + Keycloak ↔ OSB response mapping |
| `router` | Echo setup with middleware (security headers, basic auth, request logging) |
| `health` | Health endpoint verifying Keycloak connectivity |
| `metrics` | Prometheus `/metrics` endpoint |

## Design Decisions

- **Fully stateless** — No local database. The OSB `instance_id` becomes the Keycloak `clientId`. Service and plan IDs are stored as Keycloak client attributes for later retrieval.
- **Idempotent provisioning** — PUT checks if the client already exists. Returns `200` with existing data if so, `201` on new creation.
- **In-place updates** — PATCH merges updated parameters (e.g. `redirectURIs`, `implicitFlowEnabled`, `directAccessGrantsEnabled`, `consentRequired`) into the existing client without changing `clientId` or `clientSecret`.
- **Binding is a no-op** — Bind/unbind intentionally do not cycle client credentials. They return the existing client credentials in a format well-suited for Cloud Foundry compatibility.
- **Token caching** — Admin tokens are cached with a 15-second safety buffer before expiry.
- **Discovery caching** — The OIDC `.well-known` discovery response is fetched once and cached for the lifetime of the process.
- **Shared Keycloak client** — A single `keycloak.Client` instance is shared across health and broker handlers, providing one token cache and one discovery cache.

## Service Catalog

Defined in `catalog.yaml`. Ships with one service (`corpid`) and two plans:

| Plan | Type | Description |
|------|------|-------------|
| `standard-client` | Confidential | Standard OIDC client with client secret |
| `public-client` | Public | Public OIDC client (no client secret) |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `BROKER_USERNAME` | _(empty)_ | Basic auth username (auth disabled if empty) |
| `BROKER_PASSWORD` | _(empty)_ | Basic auth password |
| `BROKER_LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `BROKER_LOG_TIMESTAMP` | `false` | Include timestamps in log output |
| `KEYCLOAK_URL` | `http://localhost:8080` | Keycloak base URL |
| `KEYCLOAK_REALM` | _(empty)_ | Target Keycloak realm |
| `KEYCLOAK_ADMIN` | _(empty)_ | Keycloak admin username |
| `KEYCLOAK_PASSWORD` | _(empty)_ | Keycloak admin password |

## Development

### Prerequisites

- Go 1.25+
- Docker (for local Keycloak)
- [air](https://github.com/air-verse/air) (optional, for hot-reload)

### Quick Start

```bash
# start a local Keycloak dev server with a pre-configured realm
make start-keycloak

# run the broker with hot-reload
make run

# or run directly with race detector
make dev
```

### Example OSB Operations

```bash
# provision an OIDC client
make provision-instance

# update parameters on an existing instance (e.g. redirectURIs)
make update-instance

# fetch the instance
make fetch-instance

# bind (returns credentials)
make bind-instance

# fetch the binding (returns credentials)
make fetch-binding

# unbind (does nothing except to satisfy the OSB spec)
make delete-binding

# deprovision (deletes the OIDC client)
make deprovision-instance
```

### Other Targets

```bash
make health-check    # check broker health
make metrics-check   # check Prometheus metrics
make build           # build binary
make test            # run tests with race detector
make init            # initializes and updates all golang module vendoring
make install-air     # installs "air" hot-reloader
```

## Complete test flow with an OIDC client

```bash
make start-keycloak
make run
make provision-instance
make bind-instance
make test-oidc-login
```
