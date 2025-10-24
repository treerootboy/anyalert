# 快速开始指南

本指南将帮助你快速设置和运行 AnyAlert 通知框架。

## 5 分钟快速开始

### 1. 安装

```bash
# 克隆仓库
git clone https://github.com/treerootboy/anyalert.git
cd anyalert

# 安装依赖并构建
make install
make build
```

### 2. 配置

创建配置文件：

```bash
cp config.example.json config.json
```

编辑 `config.json`，配置至少一个通知渠道。以 Slack 为例：

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
        "webhook_url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
        "username": "AnyAlert",
        "channel": "#general"
      }
    }
  }
}
```

获取 Slack Webhook URL：
1. 访问 https://api.slack.com/apps
2. 创建新应用或选择现有应用
3. 启用 "Incoming Webhooks"
4. 添加新的 Webhook 到工作区
5. 复制 Webhook URL

### 3. 启动服务器

```bash
./bin/anyalert-server
```

你应该看到类似输出：

```
2025/10/23 12:00:00 Registered channel: slack (type: slack)
2025/10/23 12:00:00 Registered channels: [slack]
2025/10/23 12:00:00 Starting HTTP server on 0.0.0.0:8080
2025/10/23 12:00:00 Starting gRPC server on 0.0.0.0:9090
```

### 4. 发送第一条通知

打开新的终端窗口，发送测试通知：

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["#general"],
      "subject": "测试通知",
      "content": "Hello from AnyAlert! 🚀",
      "priority": "normal"
    }
  }'
```

你应该在 Slack 频道中看到通知消息！

### 5. 查看可用渠道

```bash
curl http://localhost:8080/api/v1/channels
```

输出：

```json
{
  "channels": ["slack"]
}
```

## 使用 Docker 快速开始

如果你已经安装了 Docker：

### 1. 准备配置文件

```bash
cp config.example.json config.json
# 编辑 config.json，添加你的配置
```

### 2. 使用 Docker Compose 启动

```bash
docker-compose up
```

### 3. 发送测试通知

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["#general"],
      "subject": "Docker 测试",
      "content": "从 Docker 容器发送的消息"
    }
  }'
```

## 配置多个渠道

编辑 `config.json` 添加更多渠道：

```json
{
  "channels": {
    "slack": {
      "enabled": true,
      "type": "slack",
      "config": {
        "webhook_url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
      }
    },
    "sms": {
      "enabled": true,
      "type": "sms",
      "config": {
        "provider": "aliyun",
        "api_key": "your_api_key",
        "api_secret": "your_api_secret",
        "from": "+1234567890"
      }
    },
    "youdu": {
      "enabled": true,
      "type": "youdu",
      "config": {
        "api_url": "https://youdu.example.com",
        "buin": 12345678,
        "app_id": "your_app_id",
        "api_key": "your_api_key"
      }
    }
  }
}
```

## 广播到多个渠道

发送通知到多个渠道：

```bash
curl -X POST http://localhost:8080/api/v1/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "channels": ["slack", "sms"],
    "message": {
      "to": ["#alerts", "+86-13800138000"],
      "subject": "系统告警",
      "content": "服务器 CPU 使用率超过 90%",
      "priority": "urgent"
    }
  }'
```

## 运行示例客户端

项目包含了一个 HTTP 客户端示例：

```bash
cd examples/http-client
go run main.go
```

## 下一步

- 查看 [API 文档](docs/API.md) 了解所有可用的 API 端点
- 阅读 [开发指南](docs/DEVELOPMENT.md) 学习如何添加自定义通知渠道
- 探索 gRPC API 使用 Protocol Buffers

## 常见问题

### 服务器启动失败

检查：
1. 端口 8080 和 9090 是否已被占用
2. 配置文件格式是否正确
3. 至少有一个渠道配置正确

### 通知发送失败

检查：
1. 渠道配置是否正确（API key、webhook URL 等）
2. 网络连接是否正常
3. 消息格式是否符合要求

### 如何停止服务器

按 `Ctrl+C` 优雅停止服务器。

## 获取帮助

- 查看 [README](README.md)
- 查看 [API 文档](docs/API.md)
- 提交 [Issue](https://github.com/treerootboy/anyalert/issues)

祝使用愉快！🎉
