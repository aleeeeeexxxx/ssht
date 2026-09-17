# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ssht` is an SSH tunnel manager for Remote Forwarding (`-R`) tunnels. Features:
- Multiple simultaneous tunnel connections
- Automatic reconnection on disconnect
- Persistent JSON configuration

## Development Commands

```bash
# Build
go build ./cmd/ssht

# Run
go run ./cmd/ssht

# Run with custom config
go run ./cmd/ssht -config /path/to/config.json

# Format code
go fmt ./...

# Run tests
go test ./...

# Tidy dependencies
go mod tidy
```

## Architecture

```
cmd/ssht/main.go          CLI entry point, signal handling
internal/
├── config/config.go      JSON config load/save (~/.config/ssht/config.json)
└── tunnel/
    ├── tunnel.go         Single tunnel: SSH connect, remote forward, auto-reconnect
    └── manager.go        Multi-tunnel orchestration, state change notifications
```

## Key Patterns

**State management**: Tunnel state changes broadcast via `chan StateChange` from Manager.

**Reconnect loop**: On connection loss, waits 5 seconds then retries automatically.

**Auth methods**: Supports SSH key (`auth_method: "key"`) or password (`auth_method: "password"`).

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
