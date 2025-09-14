package main

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/KodaTao/drilling/internal/api"
	"github.com/KodaTao/drilling/internal/config"
	"github.com/KodaTao/drilling/internal/database"
	"github.com/KodaTao/drilling/internal/middleware"
	"github.com/KodaTao/drilling/internal/repository"
	"github.com/KodaTao/drilling/internal/service"
	"github.com/KodaTao/drilling/web"
	"github.com/gin-gonic/gin"
)

// isAPI 检查路径是否是API请求
func isAPI(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/health")
}

// isStaticFile 检查路径是否是静态文件
func isStaticFile(path string) bool {
	staticExtensions := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map", ".js.map", ".css.map"}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func main() {
	// 初始化日志
	middleware.InitLogger()

	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 确保数据库连接在程序结束时关闭
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// 执行数据库迁移
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 重启后重置所有隧道状态为未启动状态
	log.Println("Resetting all tunnel status to inactive on startup...")
	if err := db.Exec("UPDATE tunnels SET status = ? WHERE status = ?", "inactive", "active").Error; err != nil {
		log.Printf("Warning: Failed to reset tunnel status: %v", err)
	} else {
		log.Println("All tunnel status reset to inactive successfully")
	}

	// 设置 Gin 模式
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 实例
	r := gin.New()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Drilling SSH Tunnel Manager is running",
		})
	})

	// 初始化仓库层
	hostRepo := repository.NewHostRepository(db)
	tunnelRepo := repository.NewTunnelRepository(db)

	// 初始化服务层
	encryptKey := cfg.Security.EncryptKey
	if encryptKey == "" {
		encryptKey = "default-encryption-key-change-in-production"
	}
	hostService := service.NewHostService(hostRepo, encryptKey)
	tunnelService := service.NewTunnelService(tunnelRepo, hostService)
	clashExportService := service.NewClashExportService(tunnelRepo, hostRepo)

	// 初始化API处理器
	hostHandler := api.NewHostHandler(hostService)
	tunnelHandler := api.NewTunnelHandler(tunnelService)
	exportHandler := api.NewExportHandler(clashExportService)

	// API 路由组
	apiV1 := r.Group("/api/v1")
	{
		// 系统状态
		apiV1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":   "running",
				"database": "connected",
				"version":  "1.0.0",
			})
		})

		// 注册主机管理路由
		hostHandler.RegisterRoutes(apiV1)

		// 注册隧道管理路由
		tunnelHandler.RegisterRoutes(apiV1)

		// 注册导出路由
		exportGroup := apiV1.Group("/export")
		{
			exportGroup.GET("/clash", exportHandler.ExportClashConfig)
			exportGroup.GET("/clash/preview", exportHandler.GetClashConfigPreview)
			exportGroup.GET("/socks5/status", exportHandler.GetSocks5TunnelsStatus)
		}
	}

	// 静态文件服务 - 嵌入前端资源
	staticFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		log.Fatalf("Failed to create static filesystem: %v", err)
	}

	// 调试：列出嵌入的文件
	log.Println("Embedded files:")
	fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil {
			log.Printf("  %s (dir: %v)", path, d.IsDir())
		}
		return nil
	})

	// 前端路由处理
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API请求，返回404
		if isAPI(path) {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// 尝试提供静态文件
		if isStaticFile(path) {
			// 移除开头的 "/"
			filePath := path[1:]
			if data, err := fs.ReadFile(staticFS, filePath); err == nil {
				// 设置正确的Content-Type
				if strings.HasSuffix(path, ".js.map") || strings.HasSuffix(path, ".css.map") || strings.HasSuffix(path, ".map") {
					c.Header("Content-Type", "application/json")
				} else if strings.HasSuffix(path, ".js") {
					c.Header("Content-Type", "application/javascript")
				} else if strings.HasSuffix(path, ".css") {
					c.Header("Content-Type", "text/css")
				}
				c.Data(http.StatusOK, "", data)
				return
			}
		}

		// 对于所有其他路由，返回index.html（SPA路由）
		c.Header("Content-Type", "text/html")
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
			return
		}
		c.Data(http.StatusOK, "text/html", data)
	})

	// 启动服务器
	address := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting Drilling SSH Tunnel Manager on %s", address)
	log.Fatal(r.Run(address))
}
