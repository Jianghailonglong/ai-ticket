package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ai-ticket/ai-ticket/internal/model"
)

// TicketClient 工单系统客户端接口
type TicketClient interface {
	ListTickets(token string, status string, page, pageSize int) (*model.TicketListResponse, error)
	GetTicket(token string, ticketID string) (*model.Ticket, error)
	ReviewTicket(token string, ticketID string, action string, comment string) (*model.ReviewResponse, error)
	GetCurrentUser(token string) (string, error)
}

// ticketClient 工单系统HTTP客户端实现
type ticketClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewTicketClient 创建工单客户端
func NewTicketClient(baseURL string, timeout time.Duration) TicketClient {
	return &ticketClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *ticketClient) ListTickets(token string, status string, page, pageSize int) (*model.TicketListResponse, error) {
	url := fmt.Sprintf("%s/tickets?status=%s&page=%d&page_size=%d", c.baseURL, status, page, pageSize)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	var result model.TicketListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &result, nil
}

func (c *ticketClient) GetTicket(token string, ticketID string) (*model.Ticket, error) {
	url := fmt.Sprintf("%s/tickets/%s", c.baseURL, ticketID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	var ticket model.Ticket
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &ticket, nil
}

func (c *ticketClient) ReviewTicket(token string, ticketID string, action string, comment string) (*model.ReviewResponse, error) {
	var url string
	if action == "agree" {
		url = fmt.Sprintf("%s/tickets/%s/approve", c.baseURL, ticketID)
	} else {
		url = fmt.Sprintf("%s/tickets/%s/reject", c.baseURL, ticketID)
	}

	body := map[string]string{"comment": comment}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upstream error: %s", string(bodyBytes))
	}

	var result model.ReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &result, nil
}

func (c *ticketClient) GetCurrentUser(token string) (string, error) {
	url := fmt.Sprintf("%s/auth/me", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth failed with status %d", resp.StatusCode)
	}

	var result struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	return result.User, nil
}
