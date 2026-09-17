# ssht

SSH tunnel manager with HTTP API and Web UI for Remote Forwarding (`-R`) tunnels.

## Features

- Multiple simultaneous tunnel connections
- Automatic reconnection on disconnect (max 5 retries)
- Config + State separation (user config vs runtime state)
- SSH key file upload via Web UI
- RESTful HTTP API with dynamic log level control
- Web UI (React + MUI) with real-time logs
- Single binary deployment (frontend embedded)
- Docker support
- Windows service support via NSSM

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

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tunnels | List all tunnels |
| POST | /api/tunnels | Create tunnel |
| GET | /api/tunnels/:name | Get tunnel status |
| DELETE | /api/tunnels/:name | Delete tunnel |
| POST | /api/tunnels/:name/start | Start tunnel |
| POST | /api/tunnels/:name/stop | Stop tunnel |
| GET | /api/log-level | Get current log level |
| PUT | /api/log-level | Set log level (debug/info/warn/error) |

### Examples

```bash
# List tunnels
curl http://localhost:6001/api/tunnels

# Create tunnel with key file
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

# Create tunnel with key content (uploaded via API)
curl -X POST http://localhost:6001/api/tunnels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-tunnel",
    "host": "example.com",
    "user": "root",
    "auth_method": "key",
    "key_content": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
    "remote_port": 8080,
    "local_port": 3000
  }'

# Start/Stop tunnel
curl -X POST http://localhost:6001/api/tunnels/my-tunnel/start
curl -X POST http://localhost:6001/api/tunnels/my-tunnel/stop

# Change log level
curl -X PUT http://localhost:6001/api/log-level \
  -H "Content-Type: application/json" \
  -d '{"level": "debug"}'
```

## Configuration

ssht uses config + state separation:

| File | Path | Purpose |
|------|------|---------|
| Config | `~/.config/ssht/config.json` | User config template (read-only) |
| State | `~/.local/state/ssht/state.json` | Runtime state (auto-generated) |
| Keys | `~/.local/state/ssht/keys/` | Uploaded SSH key files |

On startup, config and state are merged: config overwrites same-name tunnels, state-only tunnels are preserved.

### Config Format

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

## Windows Service

See [docs/windows-service.md](docs/windows-service.md) for installing ssht as a Windows service using NSSM.

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

