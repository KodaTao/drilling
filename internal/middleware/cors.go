package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 配置 CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// 允许的源
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080", "http://127.0.0.1:8080"}

	// 允许的方法
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	// 允许的头部
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	// 允许凭据
	config.AllowCredentials = true

	return cors.New(config)
}