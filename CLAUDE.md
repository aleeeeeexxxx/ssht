# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commit Style

在 commit message 中加上可爱的 emoji 前缀：
- ✨ feat: 新功能
- 🐛 fix: 修复 bug
- 📝 docs: 文档更新
- 🎨 style: 代码格式
- ♻️ refactor: 重构
- 🧪 test: 测试相关
- 🔧 chore: 杂项

## Development Tips

- 测试后台进程（如 `./ssht &`）完成后记得 `pkill ssht`，避免端口占用
- 前端开发用 `cd web && npm run dev`，后端用 `make run-debug`
- 生产构建用 `make build-all`（含前端打包）
- 集成测试需要真实 SSH 服务器
- 每次改动后记得更新相关文档（README.md, docs/, CLAUDE.md）

## Project Overview

`ssht` is an SSH tunnel manager with HTTP API and Web UI for Remote Forwarding (`-R`) tunnels. Features:
- Multiple simultaneous tunnel connections
- Automatic reconnection on disconnect
- Persistent JSON configuration
- RESTful HTTP API (Gin) with `/api` prefix
- Web UI (React + MUI) embedded in binary
- Real-time log streaming via WebSocket
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
cmd/ssht/main.go              CLI entry, HTTP server startup, embed frontend
cmd/ssht/dist/                Frontend build output (embedded via go:embed)
internal/
├── api/
│   ├── server.go             Gin HTTP API handlers (/api prefix)
│   ├── static.go             Serve embedded frontend files
│   └── websocket.go          WebSocket log streaming
├── config/config.go          JSON config load/save
├── logger/logger.go          Zap logger + broadcast hook for WebSocket
└── tunnel/
    ├── tunnel.go             SSH connect, remote forward, auto-reconnect
    └── manager.go            Multi-tunnel orchestration
web/                          React frontend source (Vite + MUI)
e2e-test/
├── docker-compose.yml        Test environment (ssht, ssh-server, echo-server)
└── e2e_test.go               E2E tests in Go
docs/
└── architecture.md           Detailed architecture documentation
```

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tunnels | List all tunnels |
| POST | /api/tunnels | Create tunnel |
| GET | /api/tunnels/:name | Get tunnel status |
| DELETE | /api/tunnels/:name | Delete tunnel |
| POST | /api/tunnels/:name/start | Start tunnel |
| POST | /api/tunnels/:name/stop | Stop tunnel |
| GET | /api/ws/logs | WebSocket log stream |

## Key Patterns

**Config + State 分离**:
- `~/.config/ssht/config.json` — 用户配置模板（只读）
- `~/.local/state/ssht/state.json` — 运行时状态（可写）
- `~/.local/state/ssht/keys/` — 托管的 SSH key 文件
- 启动时 merge：config 覆盖同名，state 独有保留

**SSH Key 托管**: UI 上传的 key 文件保存到 keys 目录，删除 tunnel 时自动清理。

**State management**: Tunnel state changes broadcast via `chan StateChange` from Manager to API.

**Reconnect loop**: On connection loss, waits 5 seconds then retries automatically.

**Auth methods**: SSH key (`auth_method: "key"`) or password (`auth_method: "password"`).

## Config File Format

用户配置: `~/.config/ssht/config.json`（只读模板）
运行状态: `~/.local/state/ssht/state.json`（运行时自动生成）

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
