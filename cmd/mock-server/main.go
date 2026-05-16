package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ai-ticket/ai-ticket/internal/model"
)

// 模拟数据
var mockTickets = []model.Ticket{
	{
		ID:        1,
		TicketID:  "T-20260513-001",
		Title:     "申请扩容-订单服务-CPU 4核",
		Scene:     "scale_up",
		Applicant: "zhangsan",
		Approver:  "testuser",
		Reason:    "业务高峰期扩容",
		Status:    "pending",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	},
	{
		ID:        2,
		TicketID:  "T-20260513-002",
		Title:     "申请扩容-支付服务-内存 8G",
		Scene:     "scale_up",
		Applicant: "lisi",
		Approver:  "testuser",
		Reason:    "内存使用率超过80%",
		Status:    "pending",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	},
	{
		ID:        3,
		TicketID:  "T-20260513-003",
		Title:     "申请扩容-用户服务-CPU 2核",
		Scene:     "scale_up",
		Applicant: "wangwu",
		Approver:  "admin",
		Reason:    "CPU使用率持续高位",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	if result == nil {
		result = []model.Ticket{}
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
		if t.TicketID == ticketID {
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
		if mockTickets[i].TicketID == ticketID {
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
	if strings.Contains(r.URL.Path, "/approve") {
		status = "approved"
		message = "工单已批准"
		ticket.Status = "approved"
	} else if strings.Contains(r.URL.Path, "/reject") {
		status = "rejected"
		message = "工单已拒绝"
		ticket.Status = "rejected"
	} else {
		http.Error(w, `{"error":"invalid action"}`, http.StatusBadRequest)
		return
	}

	ticket.Comment = req.Comment
	ticket.UpdatedAt = time.Now()

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
