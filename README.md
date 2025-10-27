# AnyAlert

一个 golang 通知框架，通过插件整合 Slack、短信、电话、有度IM等通知渠道，并提供 HTTP 和 gRPC 服务接口。

## 特性

- 🔌 **插件化架构**：支持多种通知渠道，易于扩展
- 🌐 **多服务支持**：同时提供 HTTP REST API 和 gRPC 服务
- 📡 **统一接口**：标准化的通知消息格式
- 🔄 **并发广播**：支持同时向多个渠道发送通知
- ⚙️ **配置驱动**：通过 JSON 配置文件管理所有渠道

## 支持的通知渠道

- ✅ Slack
- ✅ 短信 (SMS)
- ✅ 电话 (Phone Call)
- ✅ 有度 IM (Youdu)

## 安装

```bash
git clone https://github.com/treerootboy/anyalert.git
cd anyalert
make install
make build
```

## 配置

创建配置文件 `config.json`（可以从 `config.example.json` 复制）：

### Slack 配置

Slack 通道支持两种认证方式：

#### 方式 1：Bot User OAuth Token（推荐）

使用 Bot User OAuth Token 可以更灵活地控制发送者的名称、头像等，并支持发送到用户或频道。

```json
{
  "server": {
    "http": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 8080
    },
    "grpc": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 9090
    }
  },
  "channels": {
    "slack": {
      "enabled": true,
      "type": "slack",
      "config": {
        "bot_token": "xoxb-your-bot-user-oauth-token",
        "username": "AnyAlert",
        "channel": "#alerts",
        "icon_emoji": ":robot_face:",
        "icon_url": ""
      }
    }
  }
}
```

**配置说明：**
- `bot_token`: Slack Bot User OAuth Token（必填，格式：xoxb-...）
- `username`: Bot 显示的用户名（可选，默认：AnyAlert）
- `channel`: 默认发送频道（可选，格式：#channel-name 或 @username）
- `icon_emoji`: Bot 头像 emoji（可选，格式：:emoji_name:）
- `icon_url`: Bot 头像 URL（可选，与 icon_emoji 二选一）

**如何获取 Bot Token：**
1. 访问 [Slack API](https://api.slack.com/apps) 创建或选择应用
2. 在 "OAuth & Permissions" 页面添加 Bot Token Scopes：
   - `chat:write` - 发送消息
   - `chat:write.customize` - 自定义用户名和头像
3. 安装应用到工作区
4. 复制 "Bot User OAuth Token"（格式：xoxb-...）

#### 方式 2：Webhook URL（向后兼容）

```json
{
  "channels": {
    "slack": {
      "enabled": true,
      "type": "slack",
      "config": {
        "webhook_url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
        "username": "AnyAlert",
        "channel": "#alerts"
      }
    }
  }
}
```

## 使用方法

### 启动服务器

```bash
# 使用默认配置文件 config.json
./bin/anyalert-server

# 或指定配置文件
./bin/anyalert-server -config /path/to/config.json
```

### HTTP API 示例

#### 发送单个通知

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["user@example.com"],
      "subject": "测试通知",
      "content": "这是一条测试消息",
      "priority": "high"
    }
  }'
```

#### 发送到指定 Slack 频道或用户

使用 Bot User OAuth Token 模式时，可以灵活指定发送目标：

```bash
# 发送到指定频道
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["#engineering"],
      "subject": "部署通知",
      "content": "新版本已部署到生产环境"
    }
  }'

# 发送到指定用户
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["@john.doe"],
      "subject": "个人通知",
      "content": "您的任务已完成"
    }
  }'

# 通过 metadata 指定频道
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "subject": "告警通知",
      "content": "系统负载过高",
      "metadata": {
        "channel": "#alerts"
      }
    }
  }'
```

#### 广播通知到多个渠道

```bash
curl -X POST http://localhost:8080/api/v1/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "channels": ["slack", "sms"],
    "message": {
      "to": ["user@example.com", "+86-13800138000"],
      "subject": "紧急通知",
      "content": "系统告警：服务器CPU使用率超过90%",
      "priority": "urgent"
    }
  }'
```

#### 列出所有可用渠道

```bash
curl http://localhost:8080/api/v1/channels
```

#### 健康检查

```bash
curl http://localhost:8080/health
```

### gRPC API 示例

使用 gRPC 客户端连接到 `localhost:9090`，参考 proto 定义文件 `proto/notification.proto`。

## 开发

### 项目结构

```
anyalert/
├── api/                    # API 层
│   ├── grpc/              # gRPC 服务实现
│   └── http/              # HTTP 服务实现
├── cmd/                   # 应用程序入口
│   └── server/            # 服务器主程序
├── internal/              # 内部包
│   └── config/            # 配置管理
├── pkg/                   # 公共包
│   ├── notifier/          # 核心通知接口
│   └── plugins/           # 通知插件
│       ├── slack/         # Slack 插件
│       ├── sms/           # 短信插件
│       ├── phone/         # 电话插件
│       └── youdu/         # 有度 IM 插件
├── proto/                 # Protocol Buffers 定义
└── scripts/               # 工具脚本
```

### 用户元数据管理

AnyAlert 提供了用户元数据管理功能，可以配置每个用户在不同通知渠道的通知对象。

#### 用户管理 CLI 工具

首先构建用户管理工具：

```bash
make build-usermgr
```

#### 添加用户

```bash
# 添加用户，配置不同渠道的账号信息
./bin/usermgr add -name user1 \
  -slack user1@example.com \
  -youdu 10232 \
  -phone +8613800138000 \
  -sms +8613800138000
```

#### 列出所有用户

```bash
# 表格格式
./bin/usermgr list

# JSON 格式
./bin/usermgr list -json
```

#### 查询用户

```bash
./bin/usermgr get -name user1
```

#### 更新用户信息

```bash
# 更新用户的 Slack 账号
./bin/usermgr update -name user1 -slack updated@example.com

# 可以同时更新多个字段
./bin/usermgr update -name user1 -slack new@example.com -phone +8613800138001
```

#### 删除用户

```bash
./bin/usermgr delete -name user1
```

#### 自定义数据库路径

默认情况下，用户数据存储在当前目录的 `users.db` 文件中。可以通过 `-db` 参数指定其他路径：

```bash
./bin/usermgr add -db /path/to/users.db -name user1 -slack user1@example.com
```

### 添加新的通知渠道

1. 在 `pkg/plugins/` 下创建新目录
2. 实现 `notifier.Notifier` 接口
3. 在 `cmd/server/main.go` 的 `registerChannels` 函数中注册新插件

### 构建和测试

```bash
# 安装依赖
make install

# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行测试
make test

# 构建
make build

# 生成 proto 文件
make proto
```

## API 文档

### 消息格式

```json
{
  "to": ["recipient1", "recipient2"],
  "subject": "消息主题",
  "content": "消息内容",
  "priority": "normal",
  "metadata": {
    "key": "value"
  }
}
```

### 响应格式

```json
{
  "success": true,
  "message_id": "msg-12345",
  "error": "",
  "details": {
    "info": "additional info"
  }
}
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
