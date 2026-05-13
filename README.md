# Cloud MCP - 云平台工单系统

基于MCP协议的云平台工单系统，支持Web界面和AI助手两种方式管理工单。

## 功能特性

- 用户注册/登录（JWT认证）
- 工单创建、查询、审批
- 支持stdio和HTTP/SSE两种MCP传输模式
- Vue 3 前端界面
- SQLite轻量级存储

## 快速开始

### 前置条件

- Go 1.21+
- Node.js 18+
- npm 或 pnpm

### 后端启动

```bash
# 安装依赖
go mod tidy

# 启动API服务
go run ./cmd/api-server/
```

API服务启动在 http://localhost:8080

### 前端启动

```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端启动在 http://localhost:5173

### MCP服务启动

```bash
# 设置认证token
set TICKET_TOKEN=<your-jwt-token>

# 启动MCP服务
go run ./cmd/mcp-server/
```

## API接口

### 认证接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |

### 工单接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | `/api/tickets` | 查询工单列表 |
| GET | `/api/tickets/:id` | 获取工单详情 |
| POST | `/api/tickets` | 创建工单 |
| POST | `/api/tickets/:id/approve` | 同意工单 |
| POST | `/api/tickets/:id/reject` | 拒绝工单 |

## MCP Tools

| 工具 | 说明 | 参数 |
|-----|------|------|
| `list_pending_tickets` | 查询待审批工单 | status, page, page_size |
| `get_ticket_detail` | 获取工单详情 | ticket_id |
| `review_ticket` | 审批工单 | ticket_id, action, comment |

## 项目结构

```
cloud-mcp/
├── cmd/
│   ├── api-server/          # API服务入口
│   ├── mcp-server/          # MCP服务入口
│   └── mock-server/         # Mock服务（测试用）
├── internal/
│   ├── auth/                # JWT认证
│   ├── database/            # SQLite数据库
│   ├── handler/             # HTTP/MCP处理器
│   ├── middleware/           # 中间件
│   ├── model/               # 数据模型
│   ├── service/             # 业务逻辑层
│   └── ...
├── web/                     # Vue 3 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile
└── docker-compose.yml
```

## Docker部署

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 测试

### 后端测试

```bash
go test ./...
```

### 前端测试

```bash
cd web
npm run dev
# 浏览器访问 http://localhost:5173
```

## License

MIT
