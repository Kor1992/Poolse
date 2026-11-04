# Poolse

## Local PostgreSQL

Add the variables from `.env.example` to `.env` (preserve any existing settings)
and set a local `POSTGRES_PASSWORD`. Keep the password in `DATABASE_URL` in sync.
Start PostgreSQL with `docker compose up -d --build`. The database is available
at `localhost:5433`; the initial schema migration runs when the database volume
is created for the first time.

Compose reads `.env` automatically, but the Go application does not. Load the
URL from `.env` into the PowerShell process before starting Pulse. This also
replaces any stale `DATABASE_URL` already set in that terminal:

```powershell
$databaseUrlLine = Get-Content .env | Where-Object { $_ -match '^DATABASE_URL=' } | Select-Object -First 1
if (-not $databaseUrlLine) { throw "DATABASE_URL is missing from .env" }
$env:DATABASE_URL = $databaseUrlLine.Substring("DATABASE_URL=".Length)
go run ./cmd/pulse
```

The database data is stored in the `postgres_data` volume. To remove it and
recreate a fresh database, run `docker compose down -v`; this permanently
deletes the local database contents.
