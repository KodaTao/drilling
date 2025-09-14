package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/KodaTao/drilling/internal/service"
	"github.com/gin-gonic/gin"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	clashExportService service.ClashExportService
}

// NewExportHandler 创建导出处理器实例
func NewExportHandler(clashExportService service.ClashExportService) *ExportHandler {
	return &ExportHandler{
		clashExportService: clashExportService,
	}
}

// ExportClashConfig 导出Clash配置
// @Summary 导出Clash配置
// @Description 将所有活跃的SOCKS5隧道导出为Clash配置文件
// @Tags export
// @Accept json
// @Produce application/x-yaml
// @Success 200 {string} string "Clash配置YAML文件内容"
// @Failure 404 {object} gin.H "没有找到活跃的SOCKS5隧道"
// @Failure 500 {object} gin.H "生成配置失败"
// @Router /api/v1/export/clash [get]
func (h *ExportHandler) ExportClashConfig(c *gin.Context) {
	// 生成Clash配置
	yamlData, err := h.clashExportService.ExportClashConfigYAML()
	if err != nil {
		// 检查是否是因为没有活跃隧道
		if err.Error() == "no active SOCKS5 tunnels found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No active SOCKS5 tunnels found",
				"message": "Please start at least one SOCKS5 tunnel before exporting Clash configuration",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate Clash configuration",
			"message": "An error occurred while generating the configuration file",
			"details": err.Error(),
		})
		return
	}

	// 设置响应头，使浏览器下载文件
	filename := fmt.Sprintf("clash-config-%s.yaml", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Description", "File Transfer")

	// 返回YAML配置文件内容
	c.Data(http.StatusOK, "application/x-yaml", yamlData)
}

// GetClashConfigPreview 获取Clash配置预览
// @Summary 预览Clash配置
// @Description 获取Clash配置的JSON格式预览，不触发下载
// @Tags export
// @Accept json
// @Produce json
// @Success 200 {object} service.ClashConfig "Clash配置预览"
// @Failure 404 {object} gin.H "没有找到活跃的SOCKS5隧道"
// @Failure 500 {object} gin.H "生成配置失败"
// @Router /api/v1/export/clash/preview [get]
func (h *ExportHandler) GetClashConfigPreview(c *gin.Context) {
	// 生成Clash配置
	config, err := h.clashExportService.GenerateClashConfig()
	if err != nil {
		// 检查是否是因为没有活跃隧道
		if err.Error() == "no active SOCKS5 tunnels found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No active SOCKS5 tunnels found",
				"message": "Please start at least one SOCKS5 tunnel before generating Clash configuration",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate Clash configuration",
			"message": "An error occurred while generating the configuration",
			"details": err.Error(),
		})
		return
	}

	// 获取活跃隧道信息
	tunnels, err := h.clashExportService.GetActiveSocks5Tunnels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tunnel information",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Clash configuration preview generated successfully",
		"config":       config,
		"tunnel_count": len(tunnels),
		"tunnels":      tunnels,
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// GetSocks5TunnelsStatus 获取SOCKS5隧道状态
// @Summary 获取SOCKS5隧道状态
// @Description 获取所有SOCKS5隧道的状态信息，用于判断是否可以导出配置
// @Tags export
// @Accept json
// @Produce json
// @Success 200 {object} gin.H "SOCKS5隧道状态信息"
// @Failure 500 {object} gin.H "获取状态失败"
// @Router /api/v1/export/socks5/status [get]
func (h *ExportHandler) GetSocks5TunnelsStatus(c *gin.Context) {
	// 获取活跃的SOCKS5隧道
	activeTunnels, err := h.clashExportService.GetActiveSocks5Tunnels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get SOCKS5 tunnel status",
			"details": err.Error(),
		})
		return
	}

	// 构建状态信息
	tunnelStatus := make([]gin.H, 0, len(activeTunnels))
	for _, tunnel := range activeTunnels {
		tunnelStatus = append(tunnelStatus, gin.H{
			"id":            tunnel.ID,
			"name":          tunnel.Name,
			"host_id":       tunnel.HostID,
			"local_address": tunnel.LocalAddress,
			"local_port":    tunnel.LocalPort,
			"status":        tunnel.Status,
			"created_at":    tunnel.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "SOCKS5 tunnel status retrieved successfully",
		"active_count":   len(activeTunnels),
		"can_export":     len(activeTunnels) > 0,
		"tunnels":        tunnelStatus,
		"last_check":     time.Now().Format("2006-01-02 15:04:05"),
	})
}