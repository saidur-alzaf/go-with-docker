# go-sqlite-api

A minimal Go REST API backed by SQLite (pure-Go driver, no CGO), with a
Dockerfile and a GitHub Actions CI workflow.

## Routes

| Method | Path         | Description         |
|--------|--------------|----------------------|
| GET    | `/health`    | Health check         |
| GET    | `/items`     | List all items       |
| POST   | `/items`     | Create an item        |
| GET    | `/items/{id}`| Get a single item     |
| PUT    | `/items/{id}`| Update an item        |
| DELETE | `/items/{id}`| Delete an item        |

## Run locally

Requires Go 1.22+.

```bash
go mod tidy
go run .
```

The server listens on `:8080` by default and stores data in `./data.db`.
Override with the `ADDR` and `DB_PATH` environment variables.

### Example requests

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"first item"}'

curl http://localhost:8080/items

curl http://localhost:8080/items/1

curl -X PUT http://localhost:8080/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"updated item"}'

curl -X DELETE http://localhost:8080/items/1
```

## Run with Docker

```bash
docker build -t go-sqlite-api .
docker run -p 8080:8080 -v $(pwd)/data:/data go-sqlite-api
```

The container writes its SQLite file to `/data/data.db`; mount a volume
there to persist data across restarts.

## CI

`.github/workflows/ci.yml` runs on every push/PR to `main`:
1. `go mod tidy`, `go vet`, `go build`, `go test`
2. Builds the Docker image to make sure it compiles cleanly

## Notes

- Uses `modernc.org/sqlite`, a pure-Go SQLite driver, so no CGO or system
  SQLite library is needed to build or run the container.
- `go.sum` is not checked in here — run `go mod tidy` once (or let the
  Dockerfile/CI do it) to generate it against your network.
