# Windows 服务安装指南

使用 NSSM (Non-Sucking Service Manager) 将 ssht 安装为 Windows 服务。

## 1. 下载

- **ssht**: 从 [Releases](https://github.com/aleeeeeexxxx/ssht/releases) 下载 `ssht-windows-amd64.exe`
- **NSSM**: 从 https://nssm.cc/download 下载

## 2. 安装

```cmd
# 重命名为 ssht.exe，放到你喜欢的位置
C:\ssht\ssht.exe

# 以管理员身份运行 cmd
nssm install ssht C:\ssht\ssht.exe

# 启动服务
nssm start ssht
```

## 3. 验证

```cmd
# 查看状态
nssm status ssht

# 打开浏览器
start http://localhost:6001
```

## 4. 管理命令

```cmd
nssm start ssht      # 启动
nssm stop ssht       # 停止
nssm restart ssht    # 重启
nssm status ssht     # 状态
nssm remove ssht     # 卸载（会提示确认）
```

## 5. 配置选项（可选）

```cmd
# 使用 GUI 配置
nssm edit ssht
```

在 GUI 中可以设置：
- **Arguments**: `-debug -addr :6001`
- **Startup directory**: `C:\ssht`
- **Log on**: 运行账户
- **I/O**: stdout/stderr 重定向到日志文件

或者命令行设置：

```cmd
# 添加参数
nssm set ssht AppParameters "-debug"

# 设置日志输出
nssm set ssht AppStdout C:\ssht\logs\stdout.log
nssm set ssht AppStderr C:\ssht\logs\stderr.log

# 日志轮转
nssm set ssht AppRotateFiles 1
nssm set ssht AppRotateBytes 1048576
```

## 6. 数据目录

服务运行时数据存储在：
- 配置模板: `%APPDATA%\ssht\config.json`
- 运行状态: `%LOCALAPPDATA%\ssht\state.json`
- SSH Keys: `%LOCALAPPDATA%\ssht\keys\`

注意：服务以 SYSTEM 账户运行时，这些路径会是系统账户的目录。如需使用当前用户的 SSH key，在 NSSM 中设置 "Log on" 为你的用户账户。
