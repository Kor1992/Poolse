# Pulse

Pulse is a lightweight website monitoring service written in Go. It periodically checks configured websites, stores check results in PostgreSQL, and exposes HTTP endpoints for managing monitored sites and retrieving monitoring statistics.

## Features

- Add websites for monitoring
- Activate/deactivate monitoring for a site
- View active websites
- Get site details and check history
- Delete websites with cascade cleanup of related checks
- Calculate uptime statistics for a website
- Run background monitoring checks on a schedule

## Architecture

The project uses a simple layered structure:

- `cmd/pulse` — application entry point
- `internal/domain` — domain models
- `internal/repository` — repository contracts and PostgreSQL implementation
- `internal/service` — business logic and validation
- `internal/handler` — HTTP handlers and request DTOs
- `internal/monitor` — scheduled checks and monitoring loop

## Requirements

- Go 1.25+
- PostgreSQL
- Docker Compose (for local database setup)

## Local PostgreSQL setup

1. Copy the environment template if needed:

```bash
cp .env.example .env
```

2. Fill in the required values in `.env`. The key value is `DATABASE_URL`, which must match the actual PostgreSQL password and port used by the local database.
3. Start PostgreSQL locally:

```bash
docker compose up -d --build
```

By default the local database is expected on `localhost:5433`.

## Running the app

Before starting the Go app, load the database URL from `.env` into the current shell session:

```powershell
$databaseUrlLine = Get-Content .env | Where-Object { $_ -match '^DATABASE_URL=' } | Select-Object -First 1
if (-not $databaseUrlLine) { throw "DATABASE_URL is missing from .env" }
$env:DATABASE_URL = $databaseUrlLine.Substring("DATABASE_URL=".Length)
go run ./cmd/pulse
```

The application starts an HTTP server on port `8080`.

## HTTP API

### Websites

- `POST /api/websites` — create a website
- `GET /api/websites` — list active websites
- `GET /api/websites/{id}` — get a website by ID
- `PATCH /api/websites/{id}` — update `is_active`
- `DELETE /api/websites/{id}` — delete a website
- `GET /api/websites/{id}/checks` — get check history for a website
- `GET /api/websites/{id}/stats` — get website statistics

Example update request:

```http
PATCH /api/websites/2
Content-Type: application/json

{
  "is_active": false
}
```

## Database notes

- The database data is stored in the `postgres_data` volume.
- To reset the local database completely:

```bash
docker compose down -v
```

This removes all local database contents permanently.

## Testing

Run unit tests for the service layer:

```bash
go test ./internal/service -v
```

Run the full project tests:

```bash
go test ./...
```

## License

This project is currently used for local development and internal monitoring workflows. Update this section if you decide to publish it under a specific license.
