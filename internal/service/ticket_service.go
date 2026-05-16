package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ai-ticket/ai-ticket/internal/database"
	"github.com/ai-ticket/ai-ticket/internal/model"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrTicketProcessed   = errors.New("ticket already processed")
	ErrCommentRequired   = errors.New("comment is required when rejecting")
)

// TicketService 工单服务（MCP和API共享）
type TicketService struct{}

// NewTicketService 创建工单服务
func NewTicketService() *TicketService {
	return &TicketService{}
}

// ListTickets 查询工单列表
// filter: "my_apply" 我发起的, "my_approve" 我审批的, "" 全部（只显示与我相关的）
func (s *TicketService) ListTickets(username, status string, page, pageSize int, filter string) (*model.TicketListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	tickets, total, err := database.ListTicketsByUser(username, status, page, pageSize, filter)
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

// GetTicket 获取工单详情（需要权限校验）
func (s *TicketService) GetTicket(ticketID, username string) (*model.Ticket, error) {
	ticket, err := database.GetTicket(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	// 权限校验：只有申请人或审批人才能查看
	if ticket.Applicant != username && ticket.Approver != username {
		return nil, ErrPermissionDenied
	}

	return ticket, nil
}

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(req model.CreateTicketRequest, applicant string) (*model.Ticket, error) {
	// 生成工单ID
	ticketID := generateTicketID()

	ticket := &model.Ticket{
		TicketID:  ticketID,
		Title:     req.Title,
		Scene:     req.Scene,
		Applicant: applicant,
		Approver:  req.Approver,
		Reason:    req.Reason,
		Status:    "pending",
	}

	if err := database.CreateTicket(ticket); err != nil {
		return nil, err
	}

	return database.GetTicket(ticketID)
}

// ApproveTicket 同意工单
func (s *TicketService) ApproveTicket(ticketID, approver, comment string) (*model.ReviewResponse, error) {
	ticket, err := database.GetTicket(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	if ticket.Approver != approver {
		return nil, ErrPermissionDenied
	}

	if ticket.Status != "pending" {
		return nil, ErrTicketProcessed
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

// RejectTicket 拒绝工单
func (s *TicketService) RejectTicket(ticketID, approver, comment string) (*model.ReviewResponse, error) {
	ticket, err := database.GetTicket(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	if ticket.Approver != approver {
		return nil, ErrPermissionDenied
	}

	if ticket.Status != "pending" {
		return nil, ErrTicketProcessed
	}

	if comment == "" {
		return nil, ErrCommentRequired
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
	seq, _ := database.GetNextSequence(date)
	return fmt.Sprintf("T-%s-%03d", date, seq)
}
