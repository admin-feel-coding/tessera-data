# tessera-data

Data service for Tessera — Go / chi / PostgreSQL + pgvector.

## Run

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
```

## Test

```bash
go test ./...
gofmt -l .
go vet ./...
```
