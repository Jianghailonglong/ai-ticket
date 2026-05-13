package model

import "time"

// Ticket 工单模型
type Ticket struct {
	ID        int64     `json:"id"`
	TicketID  string    `json:"ticket_id"`
	Title     string    `json:"title"`
	Scene     string    `json:"scene"`
	Applicant string    `json:"applicant"`
	Approver  string    `json:"approver"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TicketListResponse 工单列表响应
type TicketListResponse struct {
	Tickets  []Ticket `json:"tickets"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// ReviewResponse 审批响应
type ReviewResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	Title    string `json:"title" binding:"required"`
	Scene    string `json:"scene" binding:"required"`
	Approver string `json:"approver" binding:"required"`
	Reason   string `json:"reason"`
}

// User 用户模型
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}
