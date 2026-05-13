# 轻量级工单系统设计

## 概述

实现一个轻量级工单系统，包含Go后端API、SQLite存储、Vue 3前端，替换现有mock接口，实现真正的工单流转。

## 整体架构

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Vue 3 前端    │────▶│  Go API 服务     │────▶│  SQLite         │
│   :5173 (dev)   │     │  :8080           │     │  tickets.db     │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                              │
                              ├── REST API (工单CRUD)
                              ├── MCP Server (AI工具)
                              └── JWT 认证
```

## 数据库设计

### 用户表

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 工单表

```sql
CREATE TABLE tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id TEXT UNIQUE NOT NULL,         -- 工单编号 T-20260513-001
    title TEXT NOT NULL,
    type TEXT NOT NULL,                     -- scale_up/scale_down
    scene TEXT NOT NULL,                    -- 业务场景
    applicant TEXT NOT NULL,                -- 申请人
    approver TEXT NOT NULL,                 -- 审批人
    reason TEXT,
    status TEXT DEFAULT 'pending',          -- pending/approved/rejected
    comment TEXT,                           -- 审批意见
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## API接口

### 认证接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录，返回JWT |

### 工单接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | `/api/tickets` | 查询工单列表 |
| GET | `/api/tickets/:id` | 获取工单详情 |
| POST | `/api/tickets` | 创建工单 |
| POST | `/api/tickets/:id/approve` | 同意工单 |
| POST | `/api/tickets/:id/reject` | 拒绝工单 |

## MCP复用

MCP工具与REST API共享Service层：

| MCP工具 | REST API |
|---------|----------|
| `list_pending_tickets` | `GET /api/tickets` |
| `get_ticket_detail` | `GET /api/tickets/:id` |
| `review_ticket` | `POST /api/tickets/:id/approve` |

## 前端设计

### 技术栈

- Vue 3 + TypeScript
- Vite
- Pinia 状态管理
- Vue Router
- Element Plus

### 页面

| 页面 | 功能 |
|-----|------|
| Login | 登录/注册 |
| TicketList | 工单列表，支持筛选 |
| TicketDetail | 工单详情+审批操作 |
| TicketCreate | 创建工单 |

## 项目结构

```
cloud-mcp/
├── cmd/
│   ├── api-server/          # API服务入口
│   ├── mcp-server/          # MCP服务入口
│   └── mock-server/         # Mock服务（测试用）
├── internal/
│   ├── database/            # SQLite操作
│   ├── handler/             # HTTP API处理器
│   ├── service/             # 业务逻辑层
│   ├── auth/                # JWT认证
│   ├── client/              # MCP客户端接口
│   ├── config/              # 配置
│   ├── errors/              # 错误码
│   └── model/               # 数据模型
├── web/                     # Vue 3 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile
└── docker-compose.yml
```

## 部署方案

### 开发环境

```bash
# 后端
go run ./cmd/api-server/

# 前端
cd web && npm run dev
```

### Docker部署

```yaml
version: '3'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data

  web:
    build: ./web
    ports:
      - "80:80"
    depends_on:
      - api
```
