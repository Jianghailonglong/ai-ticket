package errors

import "fmt"

// MCPError MCP协议错误
type MCPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *MCPError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 标准错误码
var (
	ErrTicketNotFound  = &MCPError{Code: "TICKET_NOT_FOUND", Message: "ticket not found"}
	ErrPermissionDenied = &MCPError{Code: "PERMISSION_DENIED", Message: "permission denied: you are not the approver of this ticket"}
	ErrInvalidAction   = &MCPError{Code: "INVALID_ACTION", Message: "invalid action: must be 'agree' or 'disagree'"}
	ErrCommentRequired = &MCPError{Code: "COMMENT_REQUIRED", Message: "comment is required when rejecting a ticket"}
	ErrUpstreamError   = &MCPError{Code: "UPSTREAM_ERROR", Message: "failed to call upstream ticket system"}
	ErrInvalidToken    = &MCPError{Code: "INVALID_TOKEN", Message: "authentication failed"}
)

// New 创建自定义错误
func New(code, message string) *MCPError {
	return &MCPError{Code: code, Message: message}
}

// Wrap 包装上游错误
func Wrap(code string, err error) *MCPError {
	return &MCPError{Code: code, Message: err.Error()}
}
