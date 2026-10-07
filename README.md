# Lockwood WebSocket Chat

Small room-based chat server written in Go. Clients connect over WebSocket and exchange JSON commands for authentication, room membership, and chat messages.

## Run

```sh
cp .env.example .env
make run
```

Open [http://localhost:8080/](http://localhost:8080/) in a browser to use the chat UI.

The server exposes:

- `GET /` — simple WebSocket chat UI
- `GET /healthz` — health check
- `GET /ws` — WebSocket chat endpoint

Run tests:

```sh
make test
```

The HTTP layer is generated from [openapi.yaml](./openapi.yaml). It currently exposes health and WebSocket upgrade endpoints, and can be extended with HTTP endpoints such as a room list for a web UI.

The application can evolve as a modular monolith, or the chat functionality can be deployed as a separately scalable service. In the latter case, other services should communicate with it through explicit gRPC endpoints, for example to retrieve the current room list, instead of accessing its storage directly.
