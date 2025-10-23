# AnyAlert API 文档

## HTTP REST API

### 基础 URL
```
http://localhost:8080
```

### 端点

#### 1. 发送单个通知

**端点:** `POST /api/v1/send`

**请求体:**
```json
{
  "channel": "slack",
  "message": {
    "to": ["user@example.com"],
    "subject": "通知标题",
    "content": "通知内容",
    "priority": "normal",
    "metadata": {
      "custom_key": "custom_value"
    }
  }
}
```

**响应:**
```json
{
  "success": true,
  "message_id": "msg-12345",
  "error": "",
  "details": {
    "status_code": "200"
  }
}
```

**示例:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "slack",
    "message": {
      "to": ["#general"],
      "subject": "系统通知",
      "content": "服务器正常运行",
      "priority": "normal"
    }
  }'
```

#### 2. 广播到多个渠道

**端点:** `POST /api/v1/broadcast`

**请求体:**
```json
{
  "channels": ["slack", "sms", "youdu"],
  "message": {
    "to": ["user@example.com", "+86-13800138000"],
    "subject": "紧急通知",
    "content": "系统告警",
    "priority": "urgent"
  }
}
```

**响应:**
```json
{
  "results": {
    "slack": {
      "success": true,
      "message_id": "slack-123",
      "error": ""
    },
    "sms": {
      "success": true,
      "message_id": "sms-456",
      "error": ""
    },
    "youdu": {
      "success": false,
      "error": "连接失败"
    }
  }
}
```

**示例:**
```bash
curl -X POST http://localhost:8080/api/v1/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "channels": ["slack", "sms"],
    "message": {
      "to": ["ops@example.com"],
      "subject": "告警",
      "content": "CPU使用率超过90%"
    }
  }'
```

#### 3. 列出可用渠道

**端点:** `GET /api/v1/channels`

**响应:**
```json
{
  "channels": ["slack", "sms", "phone", "youdu"]
}
```

**示例:**
```bash
curl http://localhost:8080/api/v1/channels
```

#### 4. 健康检查

**端点:** `GET /health`

**响应:**
```json
{
  "status": "ok"
}
```

**示例:**
```bash
curl http://localhost:8080/health
```

## gRPC API

### 基础地址
```
localhost:9090
```

### Proto 定义

参考 `proto/notification.proto` 文件。

### 服务方法

#### 1. Send
发送单个通知到指定渠道。

```protobuf
rpc Send(SendRequest) returns (SendResponse);
```

#### 2. Broadcast
向多个渠道广播通知。

```protobuf
rpc Broadcast(BroadcastRequest) returns (BroadcastResponse);
```

#### 3. ListChannels
列出所有可用的通知渠道。

```protobuf
rpc ListChannels(ListChannelsRequest) returns (ListChannelsResponse);
```

## 消息格式

### Message 对象

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| to | []string | 是 | 接收者列表（邮箱、电话号码、频道名等） |
| subject | string | 否 | 消息主题/标题 |
| content | string | 是 | 消息内容 |
| priority | string | 否 | 优先级：low, normal, high, urgent |
| metadata | map[string]string | 否 | 自定义元数据 |

### Response 对象

| 字段 | 类型 | 描述 |
|------|------|------|
| success | bool | 是否发送成功 |
| message_id | string | 消息ID（如果成功） |
| error | string | 错误信息（如果失败） |
| details | map[string]interface{} | 额外的响应详情 |

## 渠道配置

### Slack

```json
{
  "enabled": true,
  "type": "slack",
  "config": {
    "webhook_url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
    "username": "AnyAlert",
    "channel": "#alerts"
  }
}
```

### SMS

```json
{
  "enabled": true,
  "type": "sms",
  "config": {
    "provider": "aliyun",
    "api_key": "your_api_key",
    "api_secret": "your_api_secret",
    "from": "+1234567890"
  }
}
```

### Phone

```json
{
  "enabled": true,
  "type": "phone",
  "config": {
    "provider": "twilio",
    "api_key": "your_api_key",
    "api_secret": "your_api_secret",
    "from": "+1234567890"
  }
}
```

### Youdu IM

```json
{
  "enabled": true,
  "type": "youdu",
  "config": {
    "api_url": "https://youdu.example.com",
    "buin": 12345678,
    "app_id": "your_app_id",
    "api_key": "your_api_key"
  }
}
```

## 错误码

| HTTP 状态码 | 描述 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源未找到 |
| 405 | 方法不允许 |
| 500 | 服务器内部错误 |

## 使用示例

### Go 客户端

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func sendNotification() error {
    payload := map[string]interface{}{
        "channel": "slack",
        "message": map[string]interface{}{
            "to":      []string{"#general"},
            "subject": "测试",
            "content": "Hello from Go",
        },
    }

    data, _ := json.Marshal(payload)
    resp, err := http.Post(
        "http://localhost:8080/api/v1/send",
        "application/json",
        bytes.NewBuffer(data),
    )
    defer resp.Body.Close()
    return err
}
```

### Python 客户端

```python
import requests

def send_notification():
    payload = {
        "channel": "slack",
        "message": {
            "to": ["#general"],
            "subject": "测试",
            "content": "Hello from Python"
        }
    }
    
    response = requests.post(
        "http://localhost:8080/api/v1/send",
        json=payload
    )
    return response.json()
```

### JavaScript 客户端

```javascript
async function sendNotification() {
    const payload = {
        channel: "slack",
        message: {
            to: ["#general"],
            subject: "测试",
            content: "Hello from JavaScript"
        }
    };
    
    const response = await fetch("http://localhost:8080/api/v1/send", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(payload)
    });
    
    return await response.json();
}
```

## 最佳实践

1. **优先级设置**: 根据消息的重要性合理设置优先级
2. **错误处理**: 始终检查响应的 `success` 字段和 `error` 信息
3. **重试机制**: 对于关键通知，建议实现重试逻辑
4. **超时设置**: 为 API 调用设置合理的超时时间
5. **并发控制**: 使用 broadcast 而不是多次调用 send 来提高效率
