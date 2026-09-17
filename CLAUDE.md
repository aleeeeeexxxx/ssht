# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ssht` is an SSH tunnel manager with HTTP API for Remote Forwarding (`-R`) tunnels. Features:
- Multiple simultaneous tunnel connections
- Automatic reconnection on disconnect
- Persistent JSON configuration
- RESTful HTTP API (Gin)
- Structured logging (Zap)

## Development Commands

```bash
# Build
make build

# Run
make run
make run-debug    # with debug logging

# Test
make test              # unit tests
make test-integration  # integration tests (requires SSH server)
make test-e2e          # e2e tests with docker-compose
make test-e2e-cleanup  # cleanup e2e environment

# Docker
make docker       # build image
make docker-run   # run container

# Other
make fmt          # format code
make tidy         # tidy dependencies
make lint         # run linter
make clean        # clean build artifacts
```

## Architecture

```
cmd/ssht/main.go              CLI entry, HTTP server startup
internal/
├── api/server.go             Gin HTTP API handlers
├── config/config.go          JSON config load/save
├── logger/logger.go          Zap logger setup
└── tunnel/
    ├── tunnel.go             SSH connect, remote forward, auto-reconnect
    └── manager.go            Multi-tunnel orchestration
e2e-test/
├── docker-compose.yml        Test environment (ssht, ssh-server, echo-server)
└── e2e_test.go               E2E tests in Go
```

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| GET | /tunnels | List all tunnels |
| POST | /tunnels | Create tunnel |
| GET | /tunnels/:name | Get tunnel status |
| DELETE | /tunnels/:name | Delete tunnel |
| POST | /tunnels/:name/start | Start tunnel |
| POST | /tunnels/:name/stop | Stop tunnel |

## Key Patterns

**State management**: Tunnel state changes broadcast via `chan StateChange` from Manager to API.

**Reconnect loop**: On connection loss, waits 5 seconds then retries automatically.

**Auth methods**: SSH key (`auth_method: "key"`) or password (`auth_method: "password"`).

## Config File Format

Location: `~/.config/ssht/config.json`

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
