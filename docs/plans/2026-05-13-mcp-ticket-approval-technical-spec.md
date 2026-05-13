# 云平台工单审批 MCP 服务技术方案

## 1. 项目概述

基于设计文档，本文档详细描述MCP服务的技术实现方案。

## 2. 核心接口设计

### 2.1 MCP Tool Schema

#### list_pending_tickets

```json
{
  "name": "list_pending_tickets",
  "description": "查询当前用户待审批的工单列表",
  "inputSchema": {
    "type": "object",
    "properties": {
      "status": {
        "type": "string",
        "enum": ["pending", "approved", "rejected"],
        "description": "工单状态过滤",
        "default": "pending"
      },
      "page": {
        "type": "integer",
        "minimum": 1,
        "description": "页码",
        "default": 1
      },
      "page_size": {
        "type": "integer",
        "minimum": 1,
        "maximum": 100,
        "description": "每页数量",
        "default": 20
      }
    }
  }
}
```

#### get_ticket_detail

```json
{
  "name": "get_ticket_detail",
  "description": "获取工单详细信息，审批前必须先查看详情",
  "inputSchema": {
    "type": "object",
    "properties": {
      "ticket_id": {
        "type": "string",
        "description": "工单ID，格式如 T-20260513-001"
      }
    },
    "required": ["ticket_id"]
  }
}
```

#### review_ticket

```json
{
  "name": "review_ticket",
  "description": "审批工单，同意或拒绝",
  "inputSchema": {
    "type": "object",
    "properties": {
      "ticket_id": {
        "type": "string",
        "description": "工单ID"
      },
      "action": {
        "type": "string",
        "enum": ["agree", "disagree"],
        "description": "审批动作：agree(同意) 或 disagree(拒绝)"
      },
      "comment": {
        "type": "string",
        "description": "审批意见，拒绝时必填"
      }
    },
    "required": ["ticket_id", "action"]
  }
}
```

### 2.2 工单系统API对接

假设现有工单系统提供以下REST API：

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | `/api/v1/tickets` | 查询工单列表 |
| GET | `/api/v1/tickets/{id}` | 获取工单详情 |
| POST | `/api/v1/tickets/{id}/approve` | 同意工单 |
| POST | `/api/v1/tickets/{id}/reject` | 拒绝工单 |

所有请求Header需携带：`Authorization: Bearer <token>`

## 3. 代码结构设计

```
cloud-mcp/
├── cmd/
│   └── mcp-server/
│       └── main.go                 # 程序入口，启动MCP Server
├── internal/
│   ├── config/
│   │   └── config.go               # 配置管理
│   ├── handler/
│   │   ├── tools.go                # MCP工具注册
│   │   ├── list_tickets.go         # list_pending_tickets 处理器
│   │   ├── get_detail.go           # get_ticket_detail 处理器
│   │   └── review_ticket.go        # review_ticket 处理器
│   ├── client/
│   │   └── ticket_client.go        # 工单系统HTTP客户端
│   ├── model/
│   │   └── ticket.go               # 数据模型定义
│   └── errors/
│       └── errors.go               # 错误类型定义
├── skill/
│   └── ticket-approval.md          # Skill文件
├── config.yaml.example             # 配置示例
├── go.mod
├── go.sum
└── README.md
```

## 4. 核心模块设计

### 4.1 配置管理 (config/config.go)

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Ticket   TicketConfig   `yaml:"ticket"`
    Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
    Transport string `yaml:"transport"` // "stdio" or "http"
    HTTPAddr  string `yaml:"http_addr"` // ":8080"
}

type TicketConfig struct {
    BaseURL string `yaml:"base_url"` // 工单系统API地址
    Timeout int    `yaml:"timeout"`  // 请求超时(秒)
}

type AuthConfig struct {
    TokenHeader string `yaml:"token_header"` // 默认 "Authorization"
}
```

### 4.2 数据模型 (model/ticket.go)

```go
type Ticket struct {
    ID        string      `json:"id"`
    Title     string      `json:"title"`
    Type      string      `json:"type"`
    Applicant string      `json:"applicant"`
    Approver  string      `json:"approver"`
    Details   TicketDetail `json:"details"`
    Status    string      `json:"status"`
    CreatedAt string      `json:"created_at"`
    Priority  string      `json:"priority"`
}

type TicketDetail struct {
    Service  string `json:"service"`
    Resource string `json:"resource"`
    Current  int    `json:"current"`
    Target   int    `json:"target"`
    Reason   string `json:"reason"`
}

type TicketListResponse struct {
    Tickets  []Ticket `json:"tickets"`
    Total    int      `json:"total"`
    Page     int      `json:"page"`
    PageSize int      `json:"page_size"`
}

type ReviewResponse struct {
    Success bool   `json:"success"`
    Status  string `json:"status"`
    Message string `json:"message"`
}
```

### 4.3 错误处理 (errors/errors.go)

```go
var (
    ErrTicketNotFound    = errors.New("ticket not found")
    ErrPermissionDenied  = errors.New("permission denied: you are not the approver of this ticket")
    ErrInvalidAction     = errors.New("invalid action: must be 'agree' or 'disagree'")
    ErrCommentRequired   = errors.New("comment is required when rejecting a ticket")
    ErrUpstreamError     = errors.New("failed to call upstream ticket system")
)

type MCPError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

### 4.4 工单客户端 (client/ticket_client.go)

```go
type TicketClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *TicketClient

// ListTickets 查询工单列表
func (c *TicketClient) ListTickets(token string, status string, page, pageSize int) (*TicketListResponse, error)

// GetTicket 获取工单详情
func (c *TicketClient) GetTicket(token string, ticketID string) (*Ticket, error)

// ApproveTicket 同意工单
func (c *TicketClient) ApproveTicket(token string, ticketID string, comment string) (*ReviewResponse, error)

// RejectTicket 拒绝工单
func (c *TicketClient) RejectTicket(token string, ticketID string, comment string) (*ReviewResponse, error)
```

### 4.5 工具处理器 (handler/)

每个处理器需要：
1. 解析MCP请求参数
2. 从上下文获取用户token
3. 调用TicketClient
4. 处理错误并返回格式化响应

```go
// handler/tools.go - 工具注册
func RegisterTools(server *mcp.Server, client *TicketClient) {
    server.AddTool(listPendingTicketsTool(), handleListPendingTickets(client))
    server.AddTool(getTicketDetailTool(), handleGetTicketDetail(client))
    server.AddTool(reviewTicketTool(), handleReviewTicket(client))
}

// handler/review_ticket.go - 审批处理核心逻辑
func handleReviewTicket(client *TicketClient) mcp.ToolHandler {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. 解析参数
        ticketID := request.Params.Arguments["ticket_id"].(string)
        action := request.Params.Arguments["action"].(string)
        comment, _ := request.Params.Arguments["comment"].(string)

        // 2. 验证参数
        if action != "agree" && action != "disagree" {
            return nil, ErrInvalidAction
        }
        if action == "disagree" && comment == "" {
            return nil, ErrCommentRequired
        }

        // 3. 获取用户token (从context)
        token := getTokenFromContext(ctx)

        // 4. 先查询工单验证权限
        ticket, err := client.GetTicket(token, ticketID)
        if err != nil {
            return nil, err
        }

        // 5. 校验审批人
        currentUser := getUserFromToken(token)
        if ticket.Approver != currentUser {
            return nil, ErrPermissionDenied
        }

        // 6. 执行审批
        var resp *ReviewResponse
        if action == "agree" {
            resp, err = client.ApproveTicket(token, ticketID, comment)
        } else {
            resp, err = client.RejectTicket(token, ticketID, comment)
        }

        // 7. 返回结果
        return formatResponse(resp)
    }
}
```

## 5. 传输层设计

### 5.1 stdio 模式

适用于本地CLI场景，通过标准输入输出通信。

```go
// main.go
if config.Server.Transport == "stdio" {
    server := mcp.NewServer(...)
    // 注册工具
    // 启动stdio server
    server.ServeStdio()
}
```

### 5.2 HTTP/SSE 模式

适用于远程调用场景。

```go
// main.go
if config.Server.Transport == "http" {
    server := mcp.NewServer(...)
    // 注册工具
    // 启动HTTP server
    handler := mcp.NewHTTPHandler(server)
    http.ListenAndServe(config.Server.HTTPAddr, handler)
}
```

**HTTP模式下的认证流程：**

```
Client Request
    │
    ▼
┌──────────────────┐
│ HTTP Handler     │
│ 从Header提取Token │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Context注入Token  │
│ ctx = context.WithValue(ctx, "token", token)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Tool Handler     │
│ 从Context获取Token │
└──────────────────┘
```

## 6. 配置文件示例

```yaml
# config.yaml
server:
  transport: "http"      # "stdio" 或 "http"
  http_addr: ":8080"

ticket:
  base_url: "https://ticket.internal.company.com/api/v1"
  timeout: 30

auth:
  token_header: "Authorization"  # Bearer token
```

## 7. 错误码定义

| 错误码 | 说明 | HTTP状态码 |
|-------|------|-----------|
| TICKET_NOT_FOUND | 工单不存在 | 404 |
| PERMISSION_DENIED | 无权操作此工单 | 403 |
| INVALID_ACTION | 无效的审批动作 | 400 |
| COMMENT_REQUIRED | 拒绝时必须填写原因 | 400 |
| UPSTREAM_ERROR | 上游系统调用失败 | 502 |
| INVALID_TOKEN | 认证失败 | 401 |

## 8. 日志规范

使用 `slog` (Go 1.21+ 标准库) 进行结构化日志输出。

```go
slog.Info("review ticket",
    "ticket_id", ticketID,
    "action", action,
    "user", currentUser,
    "result", "success",
)
```

## 9. 测试策略

### 9.1 单元测试

- 每个Tool Handler独立测试
- Mock TicketClient进行隔离测试
- 覆盖正常流程和异常场景

### 9.2 集成测试

- 启动真实MCP Server
- 模拟工单系统API响应
- 验证端到端流程

### 9.3 测试用例清单

| 场景 | 预期结果 |
|-----|---------|
| 查询待审批工单 | 返回工单列表 |
| 查看工单详情 | 返回完整工单信息 |
| 同意自己的工单 | 审批成功 |
| 拒绝自己的工单(带原因) | 审批成功 |
| 操作他人的工单 | 返回PERMISSION_DENIED |
| 拒绝工单不填原因 | 返回COMMENT_REQUIRED |
| 工单不存在 | 返回TICKET_NOT_FOUND |

## 10. 部署方案

### 10.1 构建

```bash
go build -o cloud-mcp-server ./cmd/mcp-server/
```

### 10.2 Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/mcp-server/

FROM alpine:latest
COPY --from=builder /app/server /server
COPY config.yaml /config.yaml
EXPOSE 8080
CMD ["/server", "-config", "/config.yaml"]
```

### 10.3 Claude Desktop配置示例

```json
{
  "mcpServers": {
    "cloud-ticket": {
      "command": "/path/to/cloud-mcp-server",
      "args": ["-config", "/path/to/config.yaml"],
      "env": {
        "TICKET_TOKEN": "your-bearer-token"
      }
    }
  }
}
```

## 11. 安全清单

- [ ] Bearer token验证
- [ ] 工单归属校验 (approver == currentUser)
- [ ] 输入参数校验 (ticketID格式、action枚举值)
- [ ] 请求超时控制
- [ ] 日志脱敏 (不记录完整token)
- [ ] HTTPS (生产环境)

## 12. 后续迭代方向

| 版本 | 功能 |
|-----|------|
| v1.0 | 基础审批流程 (当前) |
| v1.1 | 批量审批、审批历史 |
| v1.2 | 更多工单类型 (报警、扩缩容) |
| v2.0 | 可选的自动审批策略引擎 |
