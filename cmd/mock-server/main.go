package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-mcp/cloud-mcp/internal/model"
)

// 模拟数据
var mockTickets = []model.Ticket{
	{
		ID:        "T-20260513-001",
		Title:     "申请扩容-订单服务-CPU 4核",
		Type:      "scale_up",
		Applicant: "zhangsan",
		Approver:  "testuser",
		Details: model.TicketDetail{
			Service:  "order-service",
			Resource: "cpu",
			Current:  4,
			Target:   8,
			Reason:   "业务高峰期扩容",
		},
		Status:    "pending",
		CreatedAt: "2026-05-13T10:00:00Z",
		Priority:  "medium",
	},
	{
		ID:        "T-20260513-002",
		Title:     "申请扩容-支付服务-内存 8G",
		Type:      "scale_up",
		Applicant: "lisi",
		Approver:  "testuser",
		Details: model.TicketDetail{
			Service:  "payment-service",
			Resource: "memory",
			Current:  8,
			Target:   16,
			Reason:   "内存使用率超过80%",
		},
		Status:    "pending",
		CreatedAt: "2026-05-13T11:00:00Z",
		Priority:  "high",
	},
	{
		ID:        "T-20260513-003",
		Title:     "申请扩容-用户服务-CPU 2核",
		Type:      "scale_up",
		Applicant: "wangwu",
		Approver:  "admin",
		Details: model.TicketDetail{
			Service:  "user-service",
			Resource: "cpu",
			Current:  2,
			Target:   4,
			Reason:   "CPU使用率持续高位",
		},
		Status:    "pending",
		CreatedAt: "2026-05-13T12:00:00Z",
		Priority:  "low",
	},
}

func main() {
	mux := http.NewServeMux()

	// 查询工单列表
	mux.HandleFunc("/api/v1/tickets", handleListTickets)

	// 获取工单详情
	mux.HandleFunc("/api/v1/tickets/", handleTicket)

	// 认证接口
	mux.HandleFunc("/api/v1/auth/me", handleAuth)

	addr := ":8081"
	log.Printf("Mock工单系统启动在 %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

func handleListTickets(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}

	status := r.URL.Query().Get("status")
	var result []model.Ticket
	for _, t := range mockTickets {
		if status == "" || t.Status == status {
			result = append(result, t)
		}
	}

	resp := model.TicketListResponse{
		Tickets:  result,
		Total:    len(result),
		Page:     1,
		PageSize: 20,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleTicket(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}

	// 解析路径获取ticketID
	path := r.URL.Path
	ticketID := path[len("/api/v1/tickets/"):]

	// 处理审批请求 (URL格式: /api/v1/tickets/{id}/approve 或 /api/v1/tickets/{id}/reject)
	if r.Method == "POST" {
		// 去掉 /approve 或 /reject 后缀
		if idx := strings.Index(ticketID, "/"); idx != -1 {
			ticketID = ticketID[:idx]
		}
		handleReview(w, r, ticketID)
		return
	}

	// 查询单个工单
	for _, t := range mockTickets {
		if t.ID == ticketID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}

	http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
}

func handleReview(w http.ResponseWriter, r *http.Request, ticketID string) {
	// 查找工单
	var ticket *model.Ticket
	for i := range mockTickets {
		if mockTickets[i].ID == ticketID {
			ticket = &mockTickets[i]
			break
		}
	}

	if ticket == nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	// 检查审批人
	token := r.Header.Get("Authorization")
	currentUser := getUserFromToken(token)
	if ticket.Approver != currentUser {
		http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
		return
	}

	// 解析请求体
	var req struct {
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// 根据路径判断是同意还是拒绝
	var status, message string
	if contains(r.URL.Path, "/approve") {
		status = "approved"
		message = "工单已批准"
		ticket.Status = "approved"
	} else if contains(r.URL.Path, "/reject") {
		status = "rejected"
		message = "工单已拒绝"
		ticket.Status = "rejected"
	} else {
		http.Error(w, `{"error":"invalid action"}`, http.StatusBadRequest)
		return
	}

	resp := model.ReviewResponse{
		Success: true,
		Status:  status,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}

	token := r.Header.Get("Authorization")
	user := getUserFromToken(token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user": user})
}

func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" || token == "Bearer " {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return false
	}
	return true
}

func getUserFromToken(token string) string {
	// 简单模拟：token格式为 "Bearer test-token"
	if token == "Bearer test-token" {
		return "testuser"
	}
	return "unknown"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

func init() {
	// 重置mock数据的时间
	mockTickets[0].CreatedAt = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	mockTickets[1].CreatedAt = time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	mockTickets[2].CreatedAt = time.Now().Format(time.RFC3339)

	fmt.Println("Mock工单系统初始化完成")
}
