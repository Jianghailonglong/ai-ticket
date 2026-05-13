package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/cloud-mcp/cloud-mcp/internal/database"
	"github.com/cloud-mcp/cloud-mcp/internal/handler"
)

func main() {

	// 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 初始化数据库（复用同一个数据库）
	if err := database.Init("./data/tickets.db"); err != nil {
		logger.Error("数据库初始化失败", "error", err)
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	token := os.Getenv("TICKET_TOKEN")

	s := server.NewMCPServer("cloud-ticket-mcp", "1.0.0")
	handler.RegisterAllTools(s, token, logger)

	logger.Info("starting mcp server", "transport", "stdio")

	if err := server.ServeStdio(s); err != nil {
		logger.Error("mcp server failed", "error", err)
		log.Fatalf("MCP服务启动失败: %v", err)
	}
}
