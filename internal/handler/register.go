package handler

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
)

// ToolSet MCP工具集接口
type ToolSet interface {
	Register(s *server.MCPServer)
}

// RegisterAllTools 注册所有MCP工具集
func RegisterAllTools(s *server.MCPServer, token string, logger *slog.Logger) {
	toolSets := []ToolSet{
		NewTicketTools(token, logger),
		// 未来添加: NewMonitorTools(...), NewScaleTools(...)
	}

	for _, ts := range toolSets {
		ts.Register(s)
	}
}
