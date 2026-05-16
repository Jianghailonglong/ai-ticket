package main

import (
	"log"

	"github.com/ai-ticket/ai-ticket/internal/database"
	"github.com/ai-ticket/ai-ticket/internal/handler"
	"github.com/ai-ticket/ai-ticket/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	if err := database.Init("./data/tickets.db"); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	// CORS中间件
	r.Use(corsMiddleware())
	r.Use(middleware.FixEncoding())

	// 公开接口
	authHandler := handler.NewAuthHandler()
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// 需要认证的接口
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		ticketHandler := handler.NewTicketHandler()
		api.GET("/tickets", ticketHandler.ListTickets)
		api.GET("/tickets/:id", ticketHandler.GetTicket)
		api.POST("/tickets", ticketHandler.CreateTicket)
		api.POST("/tickets/:id/approve", ticketHandler.ApproveTicket)
		api.POST("/tickets/:id/reject", ticketHandler.RejectTicket)
	}

	log.Println("API服务启动在 :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
