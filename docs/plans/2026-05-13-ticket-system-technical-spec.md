# 工单系统技术方案

## 1. 后端实现

### 1.1 依赖库

```go
// go.mod
require (
    github.com/gin-gonic/gin v1.9.1          // HTTP框架
    github.com/mattn/go-sqlite3 v1.14.22      // SQLite驱动
    github.com/golang-jwt/jwt/v5 v5.2.1       // JWT
    golang.org/x/crypto v0.21.0               // 密码加密
    github.com/mark3labs/mcp-go v0.6.0        // MCP SDK
)
```

### 1.2 目录结构

```
internal/
├── database/
│   ├── db.go              # 数据库初始化
│   ├── user_repo.go       # 用户数据操作
│   └── ticket_repo.go     # 工单数据操作
├── service/
│   ├── auth_service.go    # 认证业务逻辑
│   └── ticket_service.go  # 工单业务逻辑（MCP和API共享）
├── handler/
│   ├── auth_handler.go    # 认证API处理器
│   └── ticket_handler.go  # 工单API处理器
├── auth/
│   └── jwt.go             # JWT工具函数
├── middleware/
│   └── auth.go            # 认证中间件
└── model/
    └── ticket.go          # 数据模型（已存在，需更新）
```

### 1.3 核心代码实现

#### database/db.go - 数据库初始化

```go
package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
    var err error
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    // 创建表
    return createTables()
}

func createTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            display_name TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS tickets (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ticket_id TEXT UNIQUE NOT NULL,
            title TEXT NOT NULL,
            type TEXT NOT NULL,
            scene TEXT NOT NULL,
            applicant TEXT NOT NULL,
            approver TEXT NOT NULL,
            reason TEXT,
            status TEXT DEFAULT 'pending',
            comment TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
    }

    for _, q := range queries {
        if _, err := DB.Exec(q); err != nil {
            return err
        }
    }
    return nil
}
```

#### service/ticket_service.go - 工单业务逻辑

```go
package service

import (
    "fmt"
    "time"

    "github.com/cloud-mcp/cloud-mcp/internal/database"
    "github.com/cloud-mcp/cloud-mcp/internal/model"
)

type TicketService struct{}

func NewTicketService() *TicketService {
    return &TicketService{}
}

func (s *TicketService) ListTickets(approver, status string, page, pageSize int) (*model.TicketListResponse, error) {
    // 查询逻辑
    tickets, total, err := database.ListTickets(approver, status, page, pageSize)
    if err != nil {
        return nil, err
    }

    return &model.TicketListResponse{
        Tickets:  tickets,
        Total:    total,
        Page:     page,
        PageSize: pageSize,
    }, nil
}

func (s *TicketService) GetTicket(ticketID string) (*model.Ticket, error) {
    return database.GetTicket(ticketID)
}

func (s *TicketService) CreateTicket(req model.CreateTicketRequest, applicant string) (*model.Ticket, error) {
    // 生成工单ID
    ticketID := generateTicketID()

    ticket := &model.Ticket{
        TicketID:  ticketID,
        Title:     req.Title,
        Type:      req.Type,
        Scene:     req.Scene,
        Applicant: applicant,
        Approver:  req.Approver,
        Reason:    req.Reason,
        Status:    "pending",
    }

    if err := database.CreateTicket(ticket); err != nil {
        return nil, err
    }

    return ticket, nil
}

func (s *TicketService) ApproveTicket(ticketID, approver, comment string) (*model.ReviewResponse, error) {
    ticket, err := database.GetTicket(ticketID)
    if err != nil {
        return nil, err
    }

    if ticket.Approver != approver {
        return nil, fmt.Errorf("permission denied")
    }

    if err := database.UpdateTicketStatus(ticketID, "approved", comment); err != nil {
        return nil, err
    }

    return &model.ReviewResponse{
        Success: true,
        Status:  "approved",
        Message: "工单已批准",
    }, nil
}

func (s *TicketService) RejectTicket(ticketID, approver, comment string) (*model.ReviewResponse, error) {
    ticket, err := database.GetTicket(ticketID)
    if err != nil {
        return nil, err
    }

    if ticket.Approver != approver {
        return nil, fmt.Errorf("permission denied")
    }

    if comment == "" {
        return nil, fmt.Errorf("comment is required")
    }

    if err := database.UpdateTicketStatus(ticketID, "rejected", comment); err != nil {
        return nil, err
    }

    return &model.ReviewResponse{
        Success: true,
        Status:  "rejected",
        Message: "工单已拒绝",
    }, nil
}

func generateTicketID() string {
    now := time.Now()
    date := now.Format("20060102")
    // 查询当天最大序号
    seq, _ := database.GetNextSequence(date)
    return fmt.Sprintf("T-%s-%03d", date, seq)
}
```

#### handler/ticket_handler.go - API处理器

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/cloud-mcp/cloud-mcp/internal/service"
)

type TicketHandler struct {
    service *service.TicketService
}

func NewTicketHandler() *TicketHandler {
    return &TicketHandler{
        service: service.NewTicketService(),
    }
}

func (h *TicketHandler) ListTickets(c *gin.Context) {
    username := c.GetString("username")
    status := c.DefaultQuery("status", "")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    resp, err := h.service.ListTickets(username, status, page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) GetTicket(c *gin.Context) {
    ticketID := c.Param("id")

    ticket, err := h.service.GetTicket(ticketID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
        return
    }

    c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
    username := c.GetString("username")

    var req model.CreateTicketRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ticket, err := h.service.CreateTicket(req, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) ApproveTicket(c *gin.Context) {
    ticketID := c.Param("id")
    username := c.GetString("username")

    var req struct {
        Comment string `json:"comment"`
    }
    c.ShouldBindJSON(&req)

    resp, err := h.service.ApproveTicket(ticketID, username, req.Comment)
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) RejectTicket(c *gin.Context) {
    ticketID := c.Param("id")
    username := c.GetString("username")

    var req struct {
        Comment string `json:"comment"`
    }
    c.ShouldBindJSON(&req)

    resp, err := h.service.RejectTicket(ticketID, username, req.Comment)
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}
```

#### auth/jwt.go - JWT工具

```go
package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your-secret-key-change-in-production")

func GenerateToken(username string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "exp":      time.Now().Add(24 * time.Hour).Unix(),
    })

    return token.SignedString(secretKey)
}

func ParseToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })

    if err != nil {
        return "", err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", jwt.ErrSignatureInvalid
    }

    username, _ := claims["username"].(string)
    return username, nil
}
```

#### middleware/auth.go - 认证中间件

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/cloud-mcp/cloud-mcp/internal/auth"
)

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            c.Abort()
            return
        }

        // 去掉 Bearer 前缀
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")

        username, err := auth.ParseToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        c.Set("username", username)
        c.Next()
    }
}
```

#### cmd/api-server/main.go - API服务入口

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/cloud-mcp/cloud-mcp/internal/database"
    "github.com/cloud-mcp/cloud-mcp/internal/handler"
    "github.com/cloud-mcp/cloud-mcp/internal/middleware"
)

func main() {
    // 初始化数据库
    if err := database.Init("./data/tickets.db"); err != nil {
        log.Fatalf("数据库初始化失败: %v", err)
    }

    r := gin.Default()

    // CORS中间件
    r.Use(corsMiddleware())

    // 公开接口
    auth := handler.NewAuthHandler()
    r.POST("/api/auth/register", auth.Register)
    r.POST("/api/auth/login", auth.Login)

    // 需要认证的接口
    api := r.Group("/api")
    api.Use(middleware.AuthRequired())
    {
        tickets := handler.NewTicketHandler()
        api.GET("/tickets", tickets.ListTickets)
        api.GET("/tickets/:id", tickets.GetTicket)
        api.POST("/tickets", tickets.CreateTicket)
        api.POST("/tickets/:id/approve", tickets.ApproveTicket)
        api.POST("/tickets/:id/reject", tickets.RejectTicket)
    }

    log.Println("API服务启动在 :8080")
    r.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

## 2. 前端实现

### 2.1 初始化项目

```bash
cd D:/project/cloud-mcp
npm create vite@latest web -- --template vue-ts
cd web
npm install
npm install element-plus vue-router@4 pinia axios
```

### 2.2 核心文件

#### src/api/ticket.ts

```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api'
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export const ticketApi = {
  list(params: { status?: string; page?: number; page_size?: number }) {
    return api.get('/tickets', { params })
  },
  get(id: string) {
    return api.get(`/tickets/${id}`)
  },
  create(data: CreateTicketRequest) {
    return api.post('/tickets', data)
  },
  approve(id: string, comment: string) {
    return api.post(`/tickets/${id}/approve`, { comment })
  },
  reject(id: string, comment: string) {
    return api.post(`/tickets/${id}/reject`, { comment })
  }
}
```

#### src/views/TicketList.vue

```vue
<template>
  <div class="ticket-list">
    <el-card>
      <template #header>
        <div class="header">
          <span>工单列表</span>
          <el-button type="primary" @click="$router.push('/tickets/create')">
            创建工单
          </el-button>
        </div>
      </template>

      <el-tabs v-model="activeStatus" @tab-change="loadTickets">
        <el-tab-pane label="待审批" name="pending" />
        <el-tab-pane label="已批准" name="approved" />
        <el-tab-pane label="已拒绝" name="rejected" />
        <el-tab-pane label="全部" name="" />
      </el-tabs>

      <el-table :data="tickets" style="width: 100%">
        <el-table-column prop="ticket_id" label="工单编号" width="150" />
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="scene" label="场景" width="120" />
        <el-table-column prop="applicant" label="申请人" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link @click="$router.push(`/tickets/${row.ticket_id}`)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        @current-change="loadTickets"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ticketApi } from '@/api/ticket'

const tickets = ref([])
const activeStatus = ref('pending')
const page = ref(1)
const total = ref(0)

const loadTickets = async () => {
  const { data } = await ticketApi.list({
    status: activeStatus.value,
    page: page.value
  })
  tickets.value = data.tickets
  total.value = data.total
}

const statusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审批',
    approved: '已批准',
    rejected: '已拒绝'
  }
  return map[status] || status
}

onMounted(loadTickets)
</script>
```

#### src/views/TicketDetail.vue

```vue
<template>
  <div class="ticket-detail">
    <el-card v-if="ticket">
      <template #header>
        <span>工单详情 - {{ ticket.ticket_id }}</span>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="标题">{{ ticket.title }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ ticket.type }}</el-descriptions-item>
        <el-descriptions-item label="场景">{{ ticket.scene }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType(ticket.status)">{{ statusText(ticket.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请人">{{ ticket.applicant }}</el-descriptions-item>
        <el-descriptions-item label="审批人">{{ ticket.approver }}</el-descriptions-item>
        <el-descriptions-item label="申请原因" :span="2">{{ ticket.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="ticket.comment" label="审批意见" :span="2">
          {{ ticket.comment }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ ticket.created_at }}</el-descriptions-item>
      </el-descriptions>

      <div v-if="ticket.status === 'pending' && isApprover" class="actions">
        <el-input v-model="comment" type="textarea" placeholder="审批意见" />
        <div class="buttons">
          <el-button type="success" @click="handleApprove">同意</el-button>
          <el-button type="danger" @click="handleReject">拒绝</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ticketApi } from '@/api/ticket'

const route = useRoute()
const router = useRouter()
const ticket = ref<any>(null)
const comment = ref('')

const isApprover = computed(() => {
  const username = localStorage.getItem('username')
  return ticket.value?.approver === username
})

const loadTicket = async () => {
  const { data } = await ticketApi.get(route.params.id as string)
  ticket.value = data
}

const handleApprove = async () => {
  await ticketApi.approve(ticket.value.ticket_id, comment.value)
  ElMessage.success('工单已批准')
  router.push('/tickets')
}

const handleReject = async () => {
  if (!comment.value) {
    ElMessage.warning('拒绝时必须填写审批意见')
    return
  }
  await ticketApi.reject(ticket.value.ticket_id, comment.value)
  ElMessage.success('工单已拒绝')
  router.push('/tickets')
}

onMounted(loadTicket)
</script>
```

### 2.3 Vite配置

#### vite.config.ts

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

## 3. 测试方案

### 3.1 后端测试

```bash
# 运行所有测试
go test ./...

# 运行特定包
go test ./internal/service/ -v
go test ./internal/handler/ -v
```

### 3.2 前端测试

```bash
cd web
npm run dev
# 浏览器访问 http://localhost:5173
```

### 3.3 集成测试步骤

1. 启动后端：`go run ./cmd/api-server/`
2. 启动前端：`cd web && npm run dev`
3. 浏览器访问 http://localhost:5173
4. 注册用户 -> 登录 -> 创建工单 -> 审批工单

### 3.4 MCP测试

```bash
# 启动MCP服务（复用新的Service层）
set TICKET_TOKEN=<jwt-token>
go run ./cmd/mcp-server/ -config config.yaml
```

## 4. 部署启动

### 4.1 开发环境

```bash
# 终端1：后端
cd D:/project/cloud-mcp
go run ./cmd/api-server/

# 终端2：前端
cd D:/project/cloud-mcp/web
npm run dev
```

### 4.2 生产环境 (Docker)

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 4.3 docker-compose.yml

```yaml
version: '3'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - JWT_SECRET=your-production-secret

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
      - api
```

### 4.4 web/Dockerfile

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### 4.5 nginx.conf

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://api:8080;
    }
}
```
