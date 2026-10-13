# go-websocket-service

A local-first Go WebSocket service. The initial scaffold exposes a health endpoint; the first feature issue will define the WebSocket vertical slice.

## Development

Requires Go 1.27.1 or newer. Run tests with `go test ./...` and start the service with `go run .`; health check: `http://localhost:8080/healthz`.

## Architecture

The initial service uses Go's standard library `net/http`, avoiding external services and dependencies. WebSocket transport, room management, and fanout will be introduced in scoped follow-up work with tests.
