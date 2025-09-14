package api

import (
	"net/http"
	"strconv"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/service"
	"github.com/gin-gonic/gin"
)

// HostHandler 主机处理器
type HostHandler struct {
	hostService service.HostService
}

// NewHostHandler 创建主机处理器实例
func NewHostHandler(hostService service.HostService) *HostHandler {
	return &HostHandler{
		hostService: hostService,
	}
}

// CreateHost 创建主机
func (h *HostHandler) CreateHost(c *gin.Context) {
	var host models.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.hostService.CreateHost(&host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create host",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Host created successfully",
		"host":    host,
	})
}

// GetHost 获取主机详情
func (h *HostHandler) GetHost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	host, err := h.hostService.GetHost(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Host not found",
			"details": err.Error(),
		})
		return
	}

	// 清除敏感信息用于返回
	hostResponse := *host
	hostResponse.Password = ""
	hostResponse.PrivateKey = ""
	hostResponse.Passphrase = ""

	c.JSON(http.StatusOK, gin.H{
		"host": hostResponse,
	})
}

// GetAllHosts 获取所有主机
func (h *HostHandler) GetAllHosts(c *gin.Context) {
	hosts, err := h.hostService.GetAllHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve hosts",
			"details": err.Error(),
		})
		return
	}

	// 清除敏感信息用于返回
	var hostsResponse []models.Host
	for _, host := range hosts {
		hostResponse := host
		hostResponse.Password = ""
		hostResponse.PrivateKey = ""
		hostResponse.Passphrase = ""
		hostsResponse = append(hostsResponse, hostResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"hosts": hostsResponse,
		"count": len(hostsResponse),
	})
}

// UpdateHost 更新主机
func (h *HostHandler) UpdateHost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	var host models.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	host.ID = uint(id)

	if err := h.hostService.UpdateHost(&host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update host",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Host updated successfully",
		"host":    host,
	})
}

// DeleteHost 删除主机
func (h *HostHandler) DeleteHost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	if err := h.hostService.DeleteHost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete host",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Host deleted successfully",
	})
}

// TestConnection 测试SSH连接
func (h *HostHandler) TestConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	if err := h.hostService.TestConnection(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Connection test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection test successful",
	})
}

// CheckStatus 检查主机状态
func (h *HostHandler) CheckStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	if err := h.hostService.CheckHostStatus(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Status check failed",
			"details": err.Error(),
		})
		return
	}

	// 获取更新后的主机信息
	host, err := h.hostService.GetHost(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get updated host status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     host.Status,
		"last_check": host.LastCheck,
		"message":    "Status check completed",
	})
}

// RegisterRoutes 注册路由
func (h *HostHandler) RegisterRoutes(router *gin.RouterGroup) {
	hosts := router.Group("/hosts")
	{
		hosts.POST("", h.CreateHost)
		hosts.GET("", h.GetAllHosts)
		hosts.GET("/:id", h.GetHost)
		hosts.PUT("/:id", h.UpdateHost)
		hosts.DELETE("/:id", h.DeleteHost)
		hosts.POST("/:id/test", h.TestConnection)
		hosts.POST("/:id/status", h.CheckStatus)
	}
}