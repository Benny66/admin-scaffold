package middleware

import (
	"base-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "请先登录，缺少Authorization头")
			return
		}

		// 支持 "Bearer <token>" 格式
		parts := strings.SplitN(authHeader, " ", 2)
		var tokenStr string
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		} else {
			tokenStr = authHeader
		}

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			utils.Unauthorized(c, "令牌无效或已过期，请重新登录")
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isAdmin", claims.IsAdmin)
		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			utils.Forbidden(c, "需要管理员权限")
			return
		}
		c.Next()
	}
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) uint {
	if id, exists := c.Get("userID"); exists {
		return id.(uint)
	}
	return 0
}
