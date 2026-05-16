package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/ai-ticket/ai-ticket/internal/auth"
	"github.com/ai-ticket/ai-ticket/internal/errors"
	"github.com/ai-ticket/ai-ticket/internal/service"
)

// TicketTools MCP工单工具集
type TicketTools struct {
	service *service.TicketService
	token   string
	logger  *slog.Logger
}

// NewTicketTools 创建MCP工单工具集
func NewTicketTools(token string, logger *slog.Logger) *TicketTools {
	return &TicketTools{
		service: service.NewTicketService(),
		token:   token,
		logger:  logger,
	}
}

// Register 注册工单相关工具到MCP服务器
func (t *TicketTools) Register(s *server.MCPServer) {
	s.AddTool(t.listTool(), t.handleList)
	s.AddTool(t.detailTool(), t.handleDetail)
	s.AddTool(t.reviewTool(), t.handleReview)
}

func (t *TicketTools) listTool() mcp.Tool {
	return mcp.NewTool("list_pending_tickets",
		mcp.WithDescription("查询当前用户待审批的工单列表"),
		mcp.WithString("status",
			mcp.Description("工单状态过滤: pending/approved/rejected"),
			mcp.Enum("pending", "approved", "rejected"),
		),
		mcp.WithString("filter",
			mcp.Description("过滤类型: my_apply(我发起的), my_approve(我审批的), 空(全部相关)"),
			mcp.Enum("", "my_apply", "my_approve"),
		),
		mcp.WithNumber("page", mcp.Description("页码，从1开始")),
		mcp.WithNumber("page_size", mcp.Description("每页数量，最大100")),
	)
}

func (t *TicketTools) detailTool() mcp.Tool {
	return mcp.NewTool("get_ticket_detail",
		mcp.WithDescription("获取工单详细信息，审批前必须先查看详情"),
		mcp.WithString("ticket_id",
			mcp.Required(),
			mcp.Description("工单ID，格式如 T-20260513-001"),
		),
	)
}

func (t *TicketTools) reviewTool() mcp.Tool {
	return mcp.NewTool("review_ticket",
		mcp.WithDescription("审批工单，同意或拒绝"),
		mcp.WithString("ticket_id", mcp.Required(), mcp.Description("工单ID")),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("审批动作：agree(同意) 或 disagree(拒绝)"),
			mcp.Enum("agree", "disagree"),
		),
		mcp.WithString("comment", mcp.Description("审批意见，拒绝时必填")),
	)
}

func (t *TicketTools) handleList(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if t.token == "" {
		return nil, errors.ErrInvalidToken
	}

	// 从token解析用户名
	username, err := parseUsername(t.token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	status := getStringArg(arguments, "status", "pending")
	filter := getStringArg(arguments, "filter", "")
	page := getIntArg(arguments, "page", 1)
	pageSize := getIntArg(arguments, "page_size", 20)

	result, err := t.service.ListTickets(username, status, page, pageSize, filter)
	if err != nil {
		t.logger.Error("list tickets failed", "error", err)
		return nil, errors.Wrap("UPSTREAM_ERROR", err)
	}

	t.logger.Info("list tickets", "user", username, "status", status, "count", len(result.Tickets))
	return toJSONResult(result)
}

func (t *TicketTools) handleDetail(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if t.token == "" {
		return nil, errors.ErrInvalidToken
	}

	// 从token解析用户名
	username, err := parseUsername(t.token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	ticketID, ok := arguments["ticket_id"].(string)
	if !ok || ticketID == "" {
		return nil, errors.New("INVALID_PARAM", "ticket_id is required")
	}

	ticket, err := t.service.GetTicket(ticketID, username)
	if err != nil {
		t.logger.Error("get ticket detail failed", "ticket_id", ticketID, "error", err)
		return nil, errors.Wrap("TICKET_NOT_FOUND", err)
	}

	t.logger.Info("get ticket detail", "ticket_id", ticketID)
	return toJSONResult(ticket)
}

func (t *TicketTools) handleReview(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if t.token == "" {
		return nil, errors.ErrInvalidToken
	}

	// 从token解析用户名
	username, err := parseUsername(t.token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	ticketID, ok := arguments["ticket_id"].(string)
	if !ok || ticketID == "" {
		return nil, errors.New("INVALID_PARAM", "ticket_id is required")
	}

	action, ok := arguments["action"].(string)
	if !ok || (action != "agree" && action != "disagree") {
		return nil, errors.ErrInvalidAction
	}

	comment, _ := arguments["comment"].(string)
	if action == "disagree" && comment == "" {
		return nil, errors.ErrCommentRequired
	}

	// 使用Service层审批
	var result interface{}
	if action == "agree" {
		result, err = t.service.ApproveTicket(ticketID, username, comment)
	} else {
		result, err = t.service.RejectTicket(ticketID, username, comment)
	}

	if err != nil {
		t.logger.Error("review ticket failed", "ticket_id", ticketID, "action", action, "error", err)
		return nil, errors.Wrap("REVIEW_FAILED", err)
	}

	t.logger.Info("review ticket success", "ticket_id", ticketID, "action", action, "user", username)
	return toJSONResult(result)
}

// parseUsername 从JWT token解析用户名
func parseUsername(token string) (string, error) {
	return auth.ParseToken(token)
}

// 工具函数

func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok && v > 0 {
		return int(v)
	}
	return defaultVal
}

func toJSONResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{Type: "text", Text: string(data)},
		},
	}, nil
}
