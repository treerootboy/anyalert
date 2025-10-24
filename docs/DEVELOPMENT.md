# 开发指南

## 项目结构

```
anyalert/
├── api/                    # API 层
│   ├── grpc/              # gRPC 服务实现
│   │   └── server.go      # gRPC 服务器
│   └── http/              # HTTP 服务实现
│       └── server.go      # HTTP REST API 服务器
├── cmd/                   # 应用程序入口
│   └── server/            # 服务器主程序
│       └── main.go        # 主入口文件
├── internal/              # 内部包（不可被外部导入）
│   └── config/            # 配置管理
│       └── config.go      # 配置结构和加载
├── pkg/                   # 公共包（可被外部导入）
│   ├── notifier/          # 核心通知接口
│   │   ├── types.go       # 数据类型定义
│   │   ├── manager.go     # 通知管理器
│   │   └── manager_test.go # 管理器测试
│   └── plugins/           # 通知插件
│       ├── slack/         # Slack 插件
│       │   ├── slack.go
│       │   └── slack_test.go
│       ├── sms/           # 短信插件
│       │   └── sms.go
│       ├── phone/         # 电话插件
│       │   └── phone.go
│       └── youdu/         # 有度 IM 插件
│           └── youdu.go
├── proto/                 # Protocol Buffers 定义
│   ├── notification.proto # Proto 定义文件
│   ├── notification.pb.go # 生成的 Go 代码
│   └── notification_grpc.pb.go # 生成的 gRPC 代码
├── scripts/               # 工具脚本
│   └── generate-proto.sh  # Proto 文件生成脚本
├── examples/              # 示例代码
│   └── http-client/       # HTTP 客户端示例
│       └── main.go
├── docs/                  # 文档
│   └── API.md            # API 文档
├── Dockerfile            # Docker 镜像构建文件
├── docker-compose.yml    # Docker Compose 配置
├── Makefile              # 构建脚本
├── config.example.json   # 配置示例
├── .gitignore            # Git 忽略文件
├── LICENSE               # 许可证
├── README.md             # 项目说明
├── go.mod                # Go 模块定义
└── go.sum                # Go 依赖锁定
```

## 开发环境设置

### 前置要求

- Go 1.21 或更高版本
- Protocol Buffers 编译器 (protoc)
- Git

### 安装依赖

```bash
# 克隆仓库
git clone https://github.com/treerootboy/anyalert.git
cd anyalert

# 安装 Go 依赖
make install

# 安装 protoc 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 开发工作流

### 1. 创建配置文件

```bash
cp config.example.json config.json
# 编辑 config.json，设置你的渠道配置
```

### 2. 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./pkg/notifier/...
```

### 3. 构建项目

```bash
# 构建服务器
make build

# 或直接使用 go build
go build -o bin/anyalert-server cmd/server/main.go
```

### 4. 运行服务器

```bash
# 使用默认配置
./bin/anyalert-server

# 使用自定义配置
./bin/anyalert-server -config /path/to/config.json
```

### 5. 代码格式化和检查

```bash
# 格式化代码
make fmt

# 运行代码检查
make lint
```

## 添加新的通知渠道

### 步骤 1: 创建插件目录

```bash
mkdir -p pkg/plugins/mynewchannel
```

### 步骤 2: 实现 Notifier 接口

创建文件 `pkg/plugins/mynewchannel/mynewchannel.go`:

```go
package mynewchannel

import (
    "context"
    "fmt"
    "github.com/treerootboy/anyalert/pkg/notifier"
)

type MyNewChannel struct {
    apiKey string
    // 添加其他配置字段
}

func NewMyNewChannel() *MyNewChannel {
    return &MyNewChannel{}
}

func (m *MyNewChannel) Name() string {
    return "mynewchannel"
}

func (m *MyNewChannel) Initialize(config map[string]interface{}) error {
    if apiKey, ok := config["api_key"].(string); ok {
        m.apiKey = apiKey
    } else {
        return fmt.Errorf("api_key is required")
    }
    return nil
}

func (m *MyNewChannel) Validate(msg *notifier.Message) error {
    if msg == nil {
        return fmt.Errorf("message cannot be nil")
    }
    if msg.Content == "" {
        return fmt.Errorf("content cannot be empty")
    }
    return nil
}

func (m *MyNewChannel) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
    // 实现发送逻辑
    return &notifier.Response{
        Success: true,
        MessageID: "msg-123",
    }, nil
}
```

### 步骤 3: 注册新渠道

在 `cmd/server/main.go` 的 `registerChannels` 函数中添加:

```go
case "mynewchannel":
    n = mynewchannel.NewMyNewChannel()
```

并添加导入:

```go
import (
    // ... 其他导入
    "github.com/treerootboy/anyalert/pkg/plugins/mynewchannel"
)
```

### 步骤 4: 添加配置

在 `config.json` 中添加:

```json
{
  "channels": {
    "mynewchannel": {
      "enabled": true,
      "type": "mynewchannel",
      "config": {
        "api_key": "your_api_key"
      }
    }
  }
}
```

### 步骤 5: 编写测试

创建文件 `pkg/plugins/mynewchannel/mynewchannel_test.go`:

```go
package mynewchannel

import (
    "context"
    "testing"
    "github.com/treerootboy/anyalert/pkg/notifier"
)

func TestMyNewChannel_Initialize(t *testing.T) {
    channel := NewMyNewChannel()
    config := map[string]interface{}{
        "api_key": "test-key",
    }
    
    err := channel.Initialize(config)
    if err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }
}

func TestMyNewChannel_Send(t *testing.T) {
    channel := NewMyNewChannel()
    channel.Initialize(map[string]interface{}{
        "api_key": "test-key",
    })
    
    msg := &notifier.Message{
        To:      []string{"test@example.com"},
        Content: "Test message",
    }
    
    resp, err := channel.Send(context.Background(), msg)
    if err != nil {
        t.Fatalf("Failed to send: %v", err)
    }
    
    if !resp.Success {
        t.Error("Expected success")
    }
}
```

## 修改 gRPC 接口

### 1. 更新 Proto 定义

编辑 `proto/notification.proto`，添加或修改消息定义。

### 2. 重新生成代码

```bash
make proto
```

### 3. 更新服务实现

在 `api/grpc/server.go` 中实现新的方法。

## 测试策略

### 单元测试

```bash
# 运行所有单元测试
go test ./...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

### 集成测试

1. 启动服务器
2. 使用示例客户端或 curl 测试 API

```bash
# 启动服务器
./bin/anyalert-server &

# 测试 API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/channels
```

## 调试

### 使用 Delve

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试服务器
dlv debug cmd/server/main.go
```

### 日志输出

在代码中添加日志:

```go
import "log"

log.Printf("Debug: %v", variable)
```

## Docker 开发

### 构建镜像

```bash
docker build -t anyalert:latest .
```

### 运行容器

```bash
docker run -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/config.json:/root/config.json \
  anyalert:latest
```

### 使用 Docker Compose

```bash
docker-compose up
```

## 贡献指南

### 提交代码

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 官方代码风格
- 运行 `go fmt` 格式化代码
- 运行 `go vet` 检查代码
- 为新功能编写测试
- 更新相关文档

### Commit 消息规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型 (type):
- feat: 新功能
- fix: 修复 bug
- docs: 文档更新
- style: 代码格式（不影响代码运行的变动）
- refactor: 重构
- test: 添加测试
- chore: 构建过程或辅助工具的变动

## 常见问题

### Q: 如何添加自定义元数据到消息？

A: 使用 `metadata` 字段:

```json
{
  "message": {
    "to": ["user@example.com"],
    "content": "Test",
    "metadata": {
      "custom_field": "custom_value"
    }
  }
}
```

### Q: 如何实现消息重试？

A: 在客户端实现重试逻辑，或在插件的 `Send` 方法中添加重试。

### Q: 支持异步发送吗？

A: 广播功能 (broadcast) 是并发发送的。单个发送 (send) 是同步的，但你可以在客户端使用 goroutine 实现异步。

## 资源

- [Go 官方文档](https://golang.org/doc/)
- [gRPC Go 快速开始](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
