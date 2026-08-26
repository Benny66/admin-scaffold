package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:   []string{"Content-Length", "Content-Disposition"},
		MaxAge:          12 * time.Hour,
	})
}
