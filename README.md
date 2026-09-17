# ssht

SSH tunnel manager with HTTP API and Web UI for Remote Forwarding (`-R`) tunnels.

## Features

- Multiple simultaneous tunnel connections
- Automatic reconnection on disconnect
- Persistent JSON configuration
- RESTful HTTP API
- Web UI (React + MUI) with real-time logs
- Single binary deployment (frontend embedded)
- Docker support

## Quick Start

```bash
# Build (includes frontend)
make build-all

# Run (default port 6001)
./bin/ssht -debug

# Open Web UI
open http://localhost:6001

# Or with Docker
make docker
make docker-run
```

## HTTP API

Base URL: `http://localhost:6001/api`

```bash
# List tunnels
curl http://localhost:6001/api/tunnels

# Create tunnel
curl -X POST http://localhost:6001/api/tunnels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-tunnel",
    "host": "example.com",
    "port": 22,
    "user": "root",
    "auth_method": "key",
    "key_path": "~/.ssh/id_rsa",
    "remote_host": "0.0.0.0",
    "remote_port": 8080,
    "local_host": "127.0.0.1",
    "local_port": 3000
  }'

# Start tunnel
curl -X POST http://localhost:6001/api/tunnels/my-tunnel/start

# Get status
curl http://localhost:6001/api/tunnels/my-tunnel

# Stop tunnel
curl -X POST http://localhost:6001/api/tunnels/my-tunnel/stop

# Delete tunnel
curl -X DELETE http://localhost:6001/api/tunnels/my-tunnel
```

## Configuration

Config file location: `~/.config/ssht/config.json`

```json
{
  "tunnels": [
    {
      "name": "dev-server",
      "host": "example.com",
      "port": 22,
      "user": "root",
      "auth_method": "key",
      "key_path": "~/.ssh/id_rsa",
      "remote_host": "0.0.0.0",
      "remote_port": 8080,
      "local_host": "127.0.0.1",
      "local_port": 3000
    }
  ]
}
```

## CLI Options

```
-addr string    HTTP server address (default ":6001")
-config string  Config file path (default "~/.config/ssht/config.json")
-debug          Enable debug logging
```

## Development

```bash
# Backend only
make build
make run-debug

# Frontend dev (hot reload)
cd web && npm install && npm run dev
# Then visit http://localhost:5173

# Full build (frontend + backend)
make build-all

# Tests
make test              # Unit tests
make test-integration  # Integration tests
make test-e2e          # E2E tests with docker-compose
```

## License

MIT
