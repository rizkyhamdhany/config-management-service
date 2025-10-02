# Configuration Management Service

A minimal configuration management service, inspired by real-world remote configuration use cases that are commonly required by backend services and applications.

## Functional Requirements

1. **Create a Configuration**
    - Accepts a configuration name and its data (JSON)
    - Validates the data against a schema (specific to the config type)
    - Stores it as version `1`

2. **Update Configuration**
    - Accepts an updated JSON payload
    - Validates the update using the same schema
    - Increments the version number

3. **Rollback a Configuration**
    - Rolls back a configuration by name, restoring from a specific version
    - Creates a new version that mirrors the chosen rollback version

4. **Fetch Configuration**
    - Retrieves the latest version of a configuration by name
    - Optionally retrieves a specific version

5. **List Versions**
    - Returns the full history of versions for a given configuration

## Config Schemas

- **feature_toggle**: Toggles a feature on/off (control flow), with optional rollout/adoption percentage
- **experiment_config**: Used for A/B testing setups
- **service_client**: Defines connection parameters to other services
- **rate_limit_policy**: Configures rate limits for a given service
- **notification_policy**: Stores notification-related settings
- **schedule_rule**: Defines scheduling or cron job rules
- **threshold_policy**: Defines minimum and maximum values for a given process

## Scope

- **Authentication**: Uses simple authentication with `x-api-key` (S2S_STATIC_KEY), assuming the service is only called by internal systems or via an API gateway
- **SQLite**: Provides lightweight persistence with safe concurrent writes, fast queries, and strong data integrity, advantages that in-memory maps or flat JSON files cannot guarantee
- **Explicit schemas**: Enforcing explicit schema types ensures deterministic validation, safer schema evolution, clearer operations, predictable performance, and better error handling—avoiding the ambiguity and risks of auto-detection
- **Versioning (append-only)**: Historical data is preserved by design, but metadata such as `created_by` or `modified_by` is not included
- **ENV**: no need .env file, all config is passed via ENV vars & docker-compose.yml
- **Migration**: automatic migration when service starts, no need to run migrations manually

---

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Project Layout](#project-layout)
3. [Quickstart (Docker — recommended)](#quickstart-docker--recommended)
4. [Local Dev (Go)](#local-dev-go)
5. [Configuration (env)](#configuration-env)
6. [Build](#build)
7. [Run](#run)
8. [Testing](#testing)
   - [Unit Tests](#unit-tests)
   - [API / Integration Tests](#api--integration-tests)
9. [API Reference](#api-reference)
10. [List of Test Case (Unit Test)](#listoftestCase)
11. [Data Model](#data-model)
12. [Troubleshooting](#troubleshooting)
13. [Dependencies](#dependencies)

---

## Prerequisites

You can choose **Docker** (no Go toolchain needed) or **local Go**:

**Option A — Docker (recommended)**
- Docker 24+ and Docker Compose 2+

**Option B — Local Go**
- macOS/Linux
- **Go 1.23+** (recommended)
  - If you must use Go 1.21.x, you’ll need a compatible `modernc.org/sqlite` version or use Docker instead.

**Common**
- `make` (optional)
- `curl` for API testing

---

## Project Layout

```
.
├─ api/                 # API contract / swagger
├─ cmd/                 # Main Service Entrypoint
├─ data/                # Default SQLite file location (created at runtime)
├─ db/                  # Database Connection & Migration
├─ internal/
│  └─ remote_config/
│     ├─ handler/        # HTTP handlers (Echo)
│     ├─ repository/     # DB repo + mocks (gomock)
│     ├─ service/        # business logic
│     └─ validator/      # JSON schema validation
├─ docker-compose.yml
├─ Dockerfile
├─ Makefile
└─ README.md


```

## Quickstart (Docker — recommended)

This path avoids local Go toolchain/version issues.

1) **Build and start:**
```bash
docker compose build
docker compose up -d
```

or 

```bash
make build
make up
```

2) **Health check:**
```bash
curl -i http://localhost:8080/healthz
```

---

## Local Dev (Go)

> Prefer Go **1.23+** to avoid dependency constraints.

1) **Run the API locally:**
```bash
go mod download
go run ./cmd
```

Server listens on `:8080` by default (base path `/api`).

---

## Configuration (env)

Edit ENV vars in docker-compose.yml

```docker-compose.yml
...
    environment:
      SERVICE_NAME: "configuration-management-service"
      SERVICE_VERSION: "0.1.0"
      DATABASE_URL: "file:/srv/data/configs.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL"
      S2S_STATIC_KEY: "super-secret-123"
...
```

---

## Build

**Docker:**
```bash
docker compose build
```
or 

**Makefile:**
```bash
make build
```

---

## Run

**Docker:**
```bash
docker compose up
```

or 

**Makefile:**
```bash
make up
```

or

**Local:**
```bash
go run ./cmd
```

## Stop

**Docker:**
```bash
docker compose down
```

or

**Makefile:**
```bash
make down
```

---

## Testing

### Unit Tests

```bash
make test
```

or 

```bash
go test ./...
```

### Test Coverage

```bash
make test-coverage
```

or 

### API / Integration Tests

Set a shell var for convenience:
```bash
API=http://localhost:8080
KEY=super-secret-123
```

**1) Health**
```bash
curl -i $API/healthz
```

**2) Create config (default version = 1, use query param to set version)**
```bash
curl -i -X POST "$API/api/configs"   -H "x-api-key: $KEY"   -H "Content-Type: application/json"   -d '{
    "name": "payment-qris-toggle",
    "type": "feature_toggle",
    "data": { "enabled": true, "description": "Enable QRIS payments in prod" }
  }'
```

**3) Get latest**
```bash
curl -i "$API/api/configs/payment-qris-toggle" -H "x-api-key: $KEY"
```

**4) Append a new version**
```bash
curl -i -X PUT "$API/api/configs/payment-qris-toggle"   -H "x-api-key: $KEY"   -H "Content-Type: application/json"   -d '{ "data": { "enabled": false, "description": "Temporarily disable" } }'
```

**5) Rollback**
```bash
curl -i -X POST "$API/api/configs/payment-qris-toggle/rollback"   -H "x-api-key: $KEY"   -H "Content-Type: application/json"   -d '{ "version": 1 }'
```

---

## API Reference

- See **`api/openapi.yml`** in repo.
- Key notes:
  - `/healthz` is public (no API key / S2S_STATIC_KEY).
  - All `/configs` endpoints require `x-api-key: <S2S_STATIC_KEY>`.
  - Write endpoints also require `Content-Type: application/json`.

---

## List of Test Case (Unit Test)

### Handler
#### create config handler
- when unsupported media type should status code 415
- when invalid json should status code 400 and error message
- when missing type/name should status code 400 and error message
- when config is already exists should status code 409 and error message
- when success

#### get handler
- when missing name should status code 400
- when latest success, no If-None-Match
- when If-None-Match matches should 304
- when by version success

#### rollback handler
- when missing config name should status code 400
- when invalid json should status code 400 and error message
- when invalid version should status code 400 and error message
- when service not found should status code 404 and error message
- when success

#### list version handler
- when missing config name should status code 400
- when success

#### Service
#### create service
- when invalid input - empty schema or name should return ErrInvalidInput
- when validator error should return ErrInvalidInput
- when already exists maps should return ErrAlreadyExists
- when repo not found maps should return ErrNotFound
- when success

##### get service
- when invalid input - empty name should return ErrInvalidInput
- when version nil should return ErrNotFound
- when version nil should return latest
- when version provided should ByVersion not found
- when version provided should ByVersion success

##### list by version serivce
- when invalid input empty name or bad version should return ErrInvalidInput
- when target not found should return ErrNotFound
- when append not found ErrNotFound should return
- when success

##### update service
- when invalid input - empty name should return ErrInvalidInput
- when latest not found maps should return ErrNotFound
- when validator returns error should return ErrInvalidInput
- when append not found maps should return ErrNotFound
- when success

#### Schema Validator
- when unknown schema type should return error
- when malformed json should return error
- when feature_toggle valid should return nil
- when feature_toggle missing required 'enabled' should return error
- when feature_toggle additional property rejected should return error
- when experiment_config invalid variants (<2) should return error
- when experiment_config valid should return nil
- when service_client invalid uri should return error
- when rate_limit_policy invalid identifier_type should return error
- when notification_policy valid
- when schedule_rule invalid window - missing end should return error
- when threshold_policy valid with null min/max

### Repository
##### append repository
- when type missing should return ErrNotFound
- when success
- when insert error
- when not found
- when success

##### create repository
- when unique violation should return ErrAlreadyExists
- when other exec error should return error
- when success

##### latest repository
- when not found should return ErrNotFound
- when success

##### list repository
- when query error should return error
- when success empty should return empty
- when success with rows should return rows

---

## Data Model

### Table: `configs`
- `id` (PK)
- `name` (TEXT)
- `type` (TEXT)
- `version` (INTEGER)
- `data` (JSON)
- `created_at` (TIMESTAMP)

---

## Troubleshooting

- **403 on /configs**: missing or wrong `x-api-key`.
- **Go version errors**: use Go 1.23+ or Docker.
- **Database locked**: check WAL readers/writers.
- **Port conflict**: stop other processes on 8080.

---

## Dependencies

- **github.com/golang/mock v1.6.0**  
  Used for generating mocks in unit tests, ensuring isolation and reproducibility when testing services and repositories.

- **github.com/labstack/echo/v4 v4.11.4**  
  A fast and minimalistic web framework for Go, providing a clean API for routing, middleware, and request/response handling.

- **github.com/xeipuuv/gojsonschema v1.2.0**  
  Used for JSON Schema validation to enforce deterministic config structure and prevent invalid configurations from being stored.

- **modernc.org/sqlite v1.35.0**  
  A pure-Go SQLite driver, chosen to avoid CGO dependencies and simplify portability while still providing transactional persistence.  