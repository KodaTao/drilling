package api

import (
	"net/http"
	"strconv"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/service"
	"github.com/gin-gonic/gin"
)

// TunnelHandler 隧道处理器
type TunnelHandler struct {
	tunnelService service.TunnelService
}

// NewTunnelHandler 创建隧道处理器实例
func NewTunnelHandler(tunnelService service.TunnelService) *TunnelHandler {
	return &TunnelHandler{
		tunnelService: tunnelService,
	}
}

// CreateTunnel 创建隧道
func (h *TunnelHandler) CreateTunnel(c *gin.Context) {
	var tunnel models.Tunnel
	if err := c.ShouldBindJSON(&tunnel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.tunnelService.CreateTunnel(&tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tunnel created successfully",
		"tunnel":  tunnel,
	})
}

// GetTunnel 获取隧道详情
func (h *TunnelHandler) GetTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	tunnel, err := h.tunnelService.GetTunnel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Tunnel not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnel": tunnel,
	})
}

// GetAllTunnels 获取所有隧道
func (h *TunnelHandler) GetAllTunnels(c *gin.Context) {
	tunnels, err := h.tunnelService.GetAllTunnels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve tunnels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnels": tunnels,
		"count":   len(tunnels),
	})
}

// GetTunnelsByHost 根据主机获取隧道
func (h *TunnelHandler) GetTunnelsByHost(c *gin.Context) {
	hostIDStr := c.Param("hostId")
	hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	tunnels, err := h.tunnelService.GetTunnelsByHost(uint(hostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve tunnels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnels": tunnels,
		"count":   len(tunnels),
	})
}

// UpdateTunnel 更新隧道
func (h *TunnelHandler) UpdateTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	var tunnel models.Tunnel
	if err := c.ShouldBindJSON(&tunnel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	tunnel.ID = uint(id)

	if err := h.tunnelService.UpdateTunnel(&tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tunnel updated successfully",
		"tunnel":  tunnel,
	})
}

// DeleteTunnel 删除隧道
func (h *TunnelHandler) DeleteTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	if err := h.tunnelService.DeleteTunnel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tunnel deleted successfully",
	})
}

// StartTunnel 启动隧道
func (h *TunnelHandler) StartTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	if err := h.tunnelService.StartTunnel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to start tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tunnel started successfully",
	})
}

// StopTunnel 停止隧道
func (h *TunnelHandler) StopTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	if err := h.tunnelService.StopTunnel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to stop tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tunnel stopped successfully",
	})
}

// RestartTunnel 重启隧道
func (h *TunnelHandler) RestartTunnel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	if err := h.tunnelService.RestartTunnel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to restart tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tunnel restarted successfully",
	})
}

// GetTunnelStatus 获取隧道状态
func (h *TunnelHandler) GetTunnelStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	status, err := h.tunnelService.GetTunnelStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tunnel status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// GetConnectionLogs 获取连接日志
func (h *TunnelHandler) GetConnectionLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tunnel ID",
		})
		return
	}

	// 获取limit参数
	limitStr := c.Query("limit")
	limit := 100 // 默认限制
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs, err := h.tunnelService.GetConnectionLogs(uint(id), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get connection logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// StartAutoTunnels 启动自动启动的隧道
func (h *TunnelHandler) StartAutoTunnels(c *gin.Context) {
	if err := h.tunnelService.StartAutoTunnels(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to start auto tunnels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Auto tunnels started successfully",
	})
}

// StopAllTunnels 停止所有隧道
func (h *TunnelHandler) StopAllTunnels(c *gin.Context) {
	if err := h.tunnelService.StopAllTunnels(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to stop all tunnels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All tunnels stopped successfully",
	})
}

// CreateMultipleLocalForwards 批量创建本地服务映射
func (h *TunnelHandler) CreateMultipleLocalForwards(c *gin.Context) {
	hostIDStr := c.Param("hostId")
	hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	var req struct {
		Services []service.LocalServiceMapping `json:"services"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if len(req.Services) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No services provided",
		})
		return
	}

	tunnels, err := h.tunnelService.CreateMultipleLocalForwards(uint(hostID), req.Services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create tunnels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Tunnels created successfully",
		"tunnels":        tunnels,
		"created_count":  len(tunnels),
		"total_services": len(req.Services),
	})
}

// CheckServiceHealth 检查本地服务健康状态
func (h *TunnelHandler) CheckServiceHealth(c *gin.Context) {
	var req struct {
		LocalAddress string `json:"local_address"`
		LocalPort    int    `json:"local_port"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.tunnelService.CheckServiceHealth(req.LocalAddress, req.LocalPort); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"healthy": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"healthy": true,
		"message": "Service is available",
	})
}

// CreateDynamicSOCKS5Tunnel 创建动态SOCKS5隧道
func (h *TunnelHandler) CreateDynamicSOCKS5Tunnel(c *gin.Context) {
	hostIDStr := c.Param("hostId")
	hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid host ID",
		})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		AutoStart   bool   `json:"auto_start"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	tunnel, err := h.tunnelService.CreateDynamicSOCKS5Tunnel(uint(hostID), req.Name, req.Description, req.AutoStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create SOCKS5 tunnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SOCKS5 tunnel created successfully",
		"tunnel":  tunnel,
	})
}

// FindAvailablePort 查找可用端口
func (h *TunnelHandler) FindAvailablePort(c *gin.Context) {
	var req struct {
		StartPort int    `json:"start_port" binding:"required,min=1,max=65535"`
		EndPort   int    `json:"end_port" binding:"required,min=1,max=65535"`
		Address   string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if req.StartPort > req.EndPort {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Start port must be less than or equal to end port",
		})
		return
	}

	port, err := h.tunnelService.FindAvailablePort(req.StartPort, req.EndPort, req.Address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No available port found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available_port": port,
		"address":        req.Address,
	})
}

// RegisterRoutes 注册路由
func (h *TunnelHandler) RegisterRoutes(router *gin.RouterGroup) {
	tunnels := router.Group("/tunnels")
	{
		tunnels.POST("", h.CreateTunnel)
		tunnels.GET("", h.GetAllTunnels)
		tunnels.GET("/:id", h.GetTunnel)
		tunnels.PUT("/:id", h.UpdateTunnel)
		tunnels.DELETE("/:id", h.DeleteTunnel)
		tunnels.POST("/:id/start", h.StartTunnel)
		tunnels.POST("/:id/stop", h.StopTunnel)
		tunnels.POST("/:id/restart", h.RestartTunnel)
		tunnels.GET("/:id/status", h.GetTunnelStatus)
		tunnels.GET("/:id/logs", h.GetConnectionLogs)
	}

	// 主机相关的隧道路由 - 使用不同的路径避免冲突
	tunnels.GET("/by-host/:hostId", h.GetTunnelsByHost)

	// 批量创建本地服务映射
	tunnels.POST("/by-host/:hostId/multiple", h.CreateMultipleLocalForwards)

	// 动态SOCKS5隧道
	tunnels.POST("/by-host/:hostId/socks5", h.CreateDynamicSOCKS5Tunnel)

	// 端口管理
	router.POST("/port/find-available", h.FindAvailablePort)

	// 服务健康检查
	router.POST("/service/health-check", h.CheckServiceHealth)

	// 全局操作
	router.POST("/tunnels/auto-start", h.StartAutoTunnels)
	router.POST("/tunnels/stop-all", h.StopAllTunnels)
}