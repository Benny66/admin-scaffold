package middleware

import (
	"base-backend/database"
	"base-backend/models"
	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// PermissionRequired 权限检查中间件
func PermissionRequired(requiredCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			utils.Unauthorized(c, "用户未登录")
			return
		}

		isAdmin, _ := c.Get("isAdmin")
		// 管理员跳过权限检查
		if isAdmin.(bool) {
			c.Next()
			return
		}

		// 检查用户是否有该权限
		var user models.User
		if err := database.DB.Preload("Roles").First(&user, userID.(uint)).Error; err != nil {
			utils.Forbidden(c, "用户不存在")
			return
		}

		// 获取用户所有角色的权限
		var roleIDs []uint
		for _, role := range user.Roles {
			roleIDs = append(roleIDs, role.ID)
		}

		if len(roleIDs) == 0 {
			utils.Forbidden(c, "您没有权限访问此资源")
			return
		}

		var permissions []models.Permission
		database.DB.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
			Where("role_permissions.role_id IN ?", roleIDs).
			Find(&permissions)

		// 检查是否有所需权限
		hasPermission := false
		for _, perm := range permissions {
			if perm.Code == requiredCode {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.Forbidden(c, "您没有权限执行此操作")
			return
		}

		c.Next()
	}
}
