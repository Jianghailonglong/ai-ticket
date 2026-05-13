# 云平台工单审批 MCP 服务设计

## 概述

为云平台提供工单审批MCP服务，让用户可以通过AI（Claude等）查看和审批工单。同时提供标准Skill，描述工单审批流程规范。

## 整体架构

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Claude/AI     │────▶│  MCP Server (Go) │────▶│  工单系统 API    │
│   + Skill       │     │  提供3个Tools     │     │  (现有REST API)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                              │
                        ┌─────┴─────┐
                        │ 安全校验层  │
                        │ 验证工单归属 │
                        └───────────┘
```

**职责划分：**
- **MCP服务**：纯工具层，提供原子操作能力
- **Skill**：使用指南，定义审批流程规范
- **策略**：由用户/部门自己在Claude配置中定义，MCP服务不感知

## MCP Tools 定义

### 1. list_pending_tickets

查询当前用户待审批的工单列表。

**参数：**
```json
{
  "status": "pending",        // 工单状态：pending/approved/rejected
  "page": 1,                  // 分页
  "page_size": 20             // 每页数量，最大100
}
```

**返回：**
```json
{
  "tickets": [
    {
      "id": "T-20260513-001",
      "title": "申请扩容-订单服务-CPU 4核",
      "type": "scale_up",
      "applicant": "zhangsan",
      "created_at": "2026-05-13T10:00:00Z",
      "priority": "medium"
    }
  ],
  "total": 42,
  "page": 1
}
```

### 2. get_ticket_detail

获取单个工单完整信息。

**参数：**
```json
{
  "ticket_id": "T-20260513-001"
}
```

**返回：**
```json
{
  "id": "T-20260513-001",
  "title": "申请扩容-订单服务-CPU 4核",
  "type": "scale_up",
  "applicant": "zhangsan",
  "approver": "lisi",
  "details": {
    "service": "order-service",
    "resource": "cpu",
    "current": 4,
    "target": 8,
    "reason": "业务高峰期扩容"
  },
  "status": "pending",
  "created_at": "2026-05-13T10:00:00Z"
}
```

### 3. review_ticket

审批工单（同意/拒绝）。

**参数：**
```json
{
  "ticket_id": "T-20260513-001",
  "action": "agree",           // 必须是: agree/disagree
  "comment": "同意，符合扩容规范"  // disagree时必填
}
```

**返回：**
```json
{
  "success": true,
  "status": "approved",
  "message": "工单已批准"
}
```

## 安全机制

### MCP服务层

- 每个操作前验证：当前用户 == 工单的 `approver`
- 不匹配则返回错误：`"permission denied: you are not the approver of this ticket"`
- 用户认证：HTTP模式从Bearer token解析用户身份

### Skill层规范

1. **必须先查后审**：调用 `review_ticket` 前，必须先调用 `get_ticket_detail`
2. **安全确认**：只能审批分配给自己的工单
3. **拒绝必须说明原因**：`action=disagree` 时，`comment` 为必填

## Skill 文件

```markdown
---
name: cloud-ticket-approval
description: 云平台工单审批技能，支持查看、同意、拒绝审批工单
---

# 云平台工单审批

## 工具说明

### list_pending_tickets
查询当前用户待审批的工单列表
- 参数: status (pending/approved/rejected), page, page_size

### get_ticket_detail
获取工单详细信息，审批前必须先查看详情
- 参数: ticket_id

### review_ticket
审批工单
- 参数: ticket_id, action (agree/disagree), comment

## 审批流程规范

1. **必须先查后审**
   - 调用 review_ticket 前，必须先调用 get_ticket_detail
   - 确认工单内容合理后再操作

2. **安全确认机制**
   - 只能审批分配给自己的工单
   - 如果工单的 approver 不是当前用户，必须拒绝操作并提示

3. **拒绝必须说明原因**
   - action=disagree 时，comment 为必填
   - 原因必须具体明确，便于申请人理解

## 使用示例

### 场景：查看并审批工单
1. list_pending_tickets() → 获取待审批列表
2. get_ticket_detail(ticket_id="T-001") → 查看详情
3. review_ticket(ticket_id="T-001", action="agree", comment="符合规范")
```

## 技术实现

### 技术栈

- Go
- MCP SDK: `github.com/mark3labs/mcp-go`
- 传输方式: 支持 stdio 和 HTTP/SSE

### 项目结构

```
cloud-mcp/
├── cmd/
│   └── mcp-server/
│       └── main.go            # 入口
├── internal/
│   ├── handler/
│   │   └── tools.go           # MCP工具实现
│   ├── client/
│   │   └── ticket_client.go   # 工单系统API客户端
│   └── auth/
│       └── auth.go            # 用户认证和权限校验
├── skill/
│   └── ticket-approval.md     # Skill文件
├── go.mod
├── go.sum
└── README.md
```

### 认证方案

- HTTP模式：从请求Header的 `Authorization: Bearer <token>` 获取用户身份
- stdio模式：从环境变量或MCP配置获取用户标识

## 后续扩展

- 添加更多工单类型支持（监控报警、扩缩容等）
- 批量审批能力
- 审批历史查询
- 自动审批策略引擎（可选）
