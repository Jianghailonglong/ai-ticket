package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/cloud-mcp/cloud-mcp/internal/model"
	"github.com/cloud-mcp/cloud-mcp/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	service *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		service: service.NewAuthService(),
	}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		DisplayName string `json:"display_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Register(model.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}, req.DisplayName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
