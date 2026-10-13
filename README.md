# go-websocket-service

A local-first Go WebSocket service with a health endpoint and a minimal text-message echo endpoint.

## Development

Requires Go 1.27.1 or newer. Run tests with `go test ./...` and start the service with `go run .`.

- Health check: `http://localhost:8080/healthz`
- WebSocket echo: `ws://localhost:8080/ws`

For example, connect to `ws://localhost:8080/ws` with a WebSocket client and send a text message; the service echoes it to the same connection. The endpoint accepts valid WebSocket upgrades and rejects ordinary HTTP requests.

## Architecture

The service uses Go's standard library `net/http` with `github.com/coder/websocket` for WebSocket transport. Tests use `httptest` and an in-process WebSocket client; no Docker or external services are needed. Rooms, broadcast/fanout, authentication, persistence, and production deployment are out of scope for this initial vertical slice.
