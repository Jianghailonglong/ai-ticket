package client

import (
	"github.com/ai-ticket/ai-ticket/internal/model"
)

// MockTicketClient 用于测试的Mock客户端
type MockTicketClient struct {
	ListTicketsFunc    func(token string, status string, page, pageSize int) (*model.TicketListResponse, error)
	GetTicketFunc      func(token string, ticketID string) (*model.Ticket, error)
	ReviewTicketFunc   func(token string, ticketID string, action string, comment string) (*model.ReviewResponse, error)
	GetCurrentUserFunc func(token string) (string, error)
}

func (m *MockTicketClient) ListTickets(token string, status string, page, pageSize int) (*model.TicketListResponse, error) {
	if m.ListTicketsFunc != nil {
		return m.ListTicketsFunc(token, status, page, pageSize)
	}
	return &model.TicketListResponse{}, nil
}

func (m *MockTicketClient) GetTicket(token string, ticketID string) (*model.Ticket, error) {
	if m.GetTicketFunc != nil {
		return m.GetTicketFunc(token, ticketID)
	}
	return &model.Ticket{}, nil
}

func (m *MockTicketClient) ReviewTicket(token string, ticketID string, action string, comment string) (*model.ReviewResponse, error) {
	if m.ReviewTicketFunc != nil {
		return m.ReviewTicketFunc(token, ticketID, action, comment)
	}
	return &model.ReviewResponse{Success: true}, nil
}

func (m *MockTicketClient) GetCurrentUser(token string) (string, error) {
	if m.GetCurrentUserFunc != nil {
		return m.GetCurrentUserFunc(token)
	}
	return "testuser", nil
}
