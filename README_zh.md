[English](README.md) | **中文**

# peerclaw-cli

PeerClaw 命令行工具。通过 REST API 与 PeerClaw Server 交互，管理 Agent、发送消息、检查服务状态。

## 安装

```bash
cd cli
go build -o peerclaw ./cmd/peerclaw
```

## 使用

### 配置

默认连接 `http://localhost:8080`。可通过环境变量或配置文件修改：

```bash
# 环境变量
export PEERCLAW_SERVER=http://my-server:8080

# 或配置文件
peerclaw config set server http://my-server:8080
peerclaw config show
```

### 通过 Claim Token 注册 Agent（推荐）

最简单的 Agent 注册方式——无需编写代码：

```bash
# Claim 一个从 Provider Console 生成的 token
peerclaw agent claim --token PCW-XXXX-XXXX

# 自定义服务器和密钥路径
peerclaw agent claim --token PCW-XXXX-XXXX --server https://peerclaw.ai --keypair ./my-agent.key
```

该命令自动生成 Ed25519 密钥对、签名 token 并向服务器注册。Agent 名称和元数据来自 token（在 Web UI 中设置）。

### Agent 管理

```bash
# 列出所有 Agent
peerclaw agent list

# 按协议过滤
peerclaw agent list -protocol a2a

# 查看 Agent 详情
peerclaw agent get <agent-id>

# 注册 Agent（手动方式——生产环境推荐使用 claim）
peerclaw agent register -name "MyAgent" -url http://localhost:3000 -protocols a2a,mcp

# 删除 Agent
peerclaw agent delete <agent-id>
```

### Agent 发现

```bash
# 按能力搜索 Agent
peerclaw agent discover -capabilities code-review,summarize

# 按协议过滤
peerclaw agent discover -capabilities translate -protocol a2a
```

### Agent 心跳

```bash
# 发送心跳（默认状态：online）
peerclaw agent heartbeat <agent-id>

# 指定状态
peerclaw agent heartbeat <agent-id> -status degraded
```

### Agent 端点验证

```bash
# 验证 Agent 端点是否可达且拥有对应密钥
peerclaw agent verify <agent-id>
```

### Agent 联系人白名单

```bash
# 列出 Agent 的联系人
peerclaw agent contacts list <agent-id>

# 添加联系人（允许另一个 Agent 发送消息）
peerclaw agent contacts add <agent-id> --contact <contact-agent-id> --alias "我的伙伴"

# 移除联系人
peerclaw agent contacts remove <agent-id> --contact <contact-agent-id>
```

### 发送消息

```bash
peerclaw send -from agent-a -to agent-b -protocol a2a -payload '{"message": "hello"}'
```

### 健康检查

```bash
peerclaw health

# JSON 输出
peerclaw health -output json
```

### 输出格式

所有列表命令支持 `-output` 参数：

- `table`（默认）：表格格式
- `json`：JSON 格式

```bash
peerclaw agent list -output json
```
