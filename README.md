# go-websocket-service

A local-first Go WebSocket service with health, echo, and named-room messaging endpoints.

## Development

Requires Go 1.27.1 or newer. Run tests with `go test ./...` and start the service with `go run .`.

- Health check: `http://localhost:8080/healthz`
- WebSocket echo: `ws://localhost:8080/ws`
- Room messaging: `ws://localhost:8080/rooms/{room-name}`

The `/ws` endpoint echoes text to the same connection. Connect two clients to the same `/rooms/{room-name}` path to exchange text messages; messages are not delivered to the sender or clients in other rooms. Reconnecting clients join a room by opening its URL again. Room names must be one non-empty path segment.

Each room client has a bounded outbound queue of 16 messages. If a recipient's queue is full, that slow client is disconnected and removed from the room; other clients continue receiving messages. This favors bounded memory and progress over guaranteed delivery.

## Architecture

The service uses Go's standard library `net/http` with `github.com/coder/websocket` for WebSocket transport. Tests use `httptest` and in-process WebSocket clients; no Docker or external services are needed. Authentication, persistence, and production deployment are out of scope.
