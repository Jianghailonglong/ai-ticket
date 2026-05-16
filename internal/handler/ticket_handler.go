package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ai-ticket/ai-ticket/internal/model"
	"github.com/ai-ticket/ai-ticket/internal/service"
)

// TicketHandler 工单处理器
type TicketHandler struct {
	service *service.TicketService
}

// NewTicketHandler 创建工单处理器
func NewTicketHandler() *TicketHandler {
	return &TicketHandler{
		service: service.NewTicketService(),
	}
}

// ListTickets 查询工单列表
func (h *TicketHandler) ListTickets(c *gin.Context) {
	username := c.GetString("username")
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filter := c.DefaultQuery("filter", "") // "my_apply" 或 "my_approve" 或空

	resp, err := h.service.ListTickets(username, status, page, pageSize, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTicket 获取工单详情
func (h *TicketHandler) GetTicket(c *gin.Context) {
	ticketID := c.Param("id")
	username := c.GetString("username")

	ticket, err := h.service.GetTicket(ticketID, username)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此工单"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// CreateTicket 创建工单
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	username := c.GetString("username")

	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.CreateTicket(req, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// ApproveTicket 同意工单
func (h *TicketHandler) ApproveTicket(c *gin.Context) {
	ticketID := c.Param("id")
	username := c.GetString("username")

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.ApproveTicket(ticketID, username, req.Comment)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RejectTicket 拒绝工单
func (h *TicketHandler) RejectTicket(c *gin.Context) {
	ticketID := c.Param("id")
	username := c.GetString("username")

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.RejectTicket(ticketID, username, req.Comment)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
