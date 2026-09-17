# Architecture

## Overview

ssht 是一个 SSH 隧道管理工具，支持 Remote Forwarding (`-R`) 模式。

```
┌─────────────────────────────────────────────────────────────────┐
│                         ssht binary                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  HTTP API   │  │   Tunnel    │  │      Embedded Web UI    │  │
│  │    (Gin)    │  │   Manager   │  │   (React + MUI + Vite)  │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         │                │                     │                 │
│         └────────────────┼─────────────────────┘                 │
│                          │                                       │
│                   ┌──────┴──────┐                                │
│                   │   Config    │                                │
│                   │   (JSON)    │                                │
│                   └─────────────┘                                │
└─────────────────────────────────────────────────────────────────┘
                           │
                           │ SSH Remote Forward (-R)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Remote SSH Server                           │
│                                                                  │
│   remote_host:remote_port  ──────►  local_host:local_port       │
│        (0.0.0.0:8080)                  (127.0.0.1:3000)         │
└─────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
ssht/
├── cmd/ssht/
│   ├── main.go              # CLI 入口，启动 HTTP 服务
│   └── dist/                # 前端构建产物 (embedded)
├── internal/
│   ├── api/
│   │   ├── server.go        # Gin HTTP 服务，路由定义
│   │   ├── static.go        # 静态文件服务 (embed.FS)
│   │   └── websocket.go     # WebSocket 日志推送
│   ├── config/
│   │   └── config.go        # JSON 配置加载/保存
│   ├── logger/
│   │   └── logger.go        # Zap logger + broadcast hook
│   └── tunnel/
│       ├── tunnel.go        # SSH 连接，Remote Forward，自动重连
│       └── manager.go       # 多隧道管理，状态追踪
├── web/                     # React 前端源码
│   ├── src/
│   │   ├── App.tsx
│   │   ├── api/tunnels.ts   # API 调用封装
│   │   └── components/
│   │       ├── TunnelList.tsx
│   │       ├── TunnelForm.tsx
│   │       └── LogViewer.tsx
│   ├── vite.config.ts
│   └── package.json
├── e2e-test/                # E2E 测试 (docker-compose)
└── docs/                    # 文档
```

## Core Components

### 1. Tunnel (`internal/tunnel/tunnel.go`)

单个 SSH 隧道的生命周期管理：

```go
type Tunnel struct {
    config   Config
    client   *ssh.Client
    listener net.Listener
    state    State
    stopCh   chan struct{}
}
```

**状态机**：
```
Stopped ──Start()──► Connecting ──success──► Connected
    ▲                    │                       │
    │                    │ fail                  │ disconnect
    │                    ▼                       ▼
    └──────Stop()────  Error  ◄──────────── Reconnecting
                         │                       │
                         └───── 5s delay ────────┘
```

**Remote Forward 流程**：
1. SSH 连接到远程服务器
2. 调用 `client.Listen("tcp", "remote_host:remote_port")`
3. 远程服务器监听端口，接受连接
4. 每个连接 dial 到本地 `local_host:local_port`
5. 双向数据转发

### 2. Manager (`internal/tunnel/manager.go`)

多隧道编排：

```go
type Manager struct {
    tunnels map[string]*Tunnel
    mu      sync.RWMutex
}
```

- `Add(config)` - 添加隧道配置
- `Start(name)` - 启动指定隧道
- `Stop(name)` - 停止指定隧道
- `Remove(name)` - 删除隧道
- `GetState(name)` - 获取隧道状态

### 3. HTTP API (`internal/api/server.go`)

RESTful API，使用 Gin 框架：

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tunnels | 列出所有隧道 |
| POST | /api/tunnels | 创建隧道 |
| GET | /api/tunnels/:name | 获取隧道状态 |
| DELETE | /api/tunnels/:name | 删除隧道 |
| POST | /api/tunnels/:name/start | 启动隧道 |
| POST | /api/tunnels/:name/stop | 停止隧道 |
| GET | /api/ws/logs | WebSocket 日志流 |

### 4. Logger (`internal/logger/logger.go`)

Zap logger + broadcast hook：

```go
type broadcastHook struct{}

func (h *broadcastHook) Write(p []byte) (n int, err error) {
    BroadcastLog(string(p))  // 推送到 WebSocket 客户端
    return len(p), nil
}
```

日志同时输出到：
- 终端 (stdout)
- WebSocket 客户端 (实时)

### 5. Config (`internal/config/config.go`)

JSON 配置持久化：

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

默认路径：`~/.config/ssht/config.json`

## Frontend Architecture

### Tech Stack

- React 18 + TypeScript
- MUI (Material UI)
- Vite (构建工具)
- WebSocket (实时日志)

### Build & Embed

```
web/src/*.tsx
     │
     │ npm run build (Vite)
     ▼
cmd/ssht/dist/
├── index.html
└── assets/
    └── index-xxx.js
     │
     │ go:embed
     ▼
ssht binary (单文件)
```

**关键代码** (`cmd/ssht/main.go`)：
```go
//go:embed all:dist
var distFS embed.FS

func main() {
    if subFS, err := fs.Sub(distFS, "dist"); err == nil {
        api.SetStaticFS(subFS)
    }
}
```

### Dev vs Prod

| 模式 | 前端 | 后端 | 访问地址 |
|------|------|------|----------|
| 开发 | Vite dev server (5173) | Go (6001) | localhost:5173 |
| 生产 | 嵌入 Go binary | Go (6001) | localhost:6001 |

开发模式下，Vite 代理 `/api` 请求到后端：

```ts
// vite.config.ts
proxy: {
  '/api': {
    target: 'http://localhost:6001',
    changeOrigin: true,
    ws: true,
  },
}
```

## Data Flow

### 创建并启动隧道

```
User ──► Web UI ──► POST /api/tunnels ──► API Server
                                              │
                                              ▼
                                         Config.Add()
                                         Manager.Add()
                                         Config.Save()
                                              │
User ◄── Web UI ◄── 201 Created ◄────────────┘

User ──► Web UI ──► POST /api/tunnels/:name/start ──► API Server
                                                          │
                                                          ▼
                                                     Manager.Start()
                                                          │
                                                          ▼
                                                     Tunnel.Start()
                                                          │
                                                          ▼
                                                     SSH Connect
                                                     Remote Listen
                                                          │
User ◄── Web UI ◄── 200 OK ◄──────────────────────────────┘
```

### 实时日志

```
Tunnel ──► Logger.Infow() ──► Zap ──► broadcastHook
                                           │
                                           ▼
                                      BroadcastLog()
                                           │
                                           ▼
                              WebSocket ──► LogViewer (React)
```

## Authentication

支持两种 SSH 认证方式：

### 1. Key-based (`auth_method: "key"`)

```json
{
  "auth_method": "key",
  "key_path": "~/.ssh/id_rsa"
}
```

- 支持 `~` 展开
- 读取私钥文件
- 使用 `ssh.PublicKeys()` 认证

### 2. Password (`auth_method: "password"`)

```json
{
  "auth_method": "password",
  "password": "your-password"
}
```

- 使用 `ssh.Password()` 认证
- 不推荐，仅用于测试

## Auto Reconnect

隧道断开后自动重连：

```go
func (t *Tunnel) reconnectLoop() {
    for {
        select {
        case <-t.stopCh:
            return
        default:
            time.Sleep(5 * time.Second)
            if err := t.connect(); err == nil {
                return
            }
        }
    }
}
```

- 断开后等待 5 秒
- 自动重试连接
- 直到成功或手动 Stop
