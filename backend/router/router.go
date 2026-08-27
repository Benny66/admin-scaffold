package router

import (
	"base-backend/controllers"
	"base-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有路由
func SetupRouter(r *gin.Engine) {
	// 静态文件服务：品牌图片（logo/favicon）放在 backend/static/，通过 /static/<file> 访问
	r.Static("/static", "./static")

	// API路由组
	api := r.Group("/api")

	// ==================== 公开接口（无需鉴权）====================
	api.GET("/system/info", controllers.GetSystemInfo)

	auth := api.Group("/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", controllers.Logout)
	}

	// ==================== 需要JWT鉴权的接口 ====================
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	protected.Use(middleware.OperationLogger())

	// 用户相关（当前用户自身信息，无需权限码，JWT 即可）
	protected.GET("/auth/info", controllers.GetUserInfo)
	protected.PUT("/auth/password", controllers.ChangePassword)

	// ==================== 系统管理（按权限码 RBAC 接线）====================
	// 用户管理
	userGroup := protected.Group("/users")
	{
		userGroup.GET("", middleware.PermissionRequired("users:view"), controllers.GetUserList)
		userGroup.GET("/:id", middleware.PermissionRequired("users:view"), controllers.GetUser)
		userGroup.POST("", middleware.PermissionRequired("users:create"), controllers.CreateUser)
		userGroup.PUT("/:id", middleware.PermissionRequired("users:edit"), controllers.UpdateUser)
		userGroup.DELETE("/:id", middleware.PermissionRequired("users:delete"), controllers.DeleteUser)
		userGroup.PUT("/:id/password", middleware.PermissionRequired("users:edit"), controllers.ResetPassword)
		userGroup.PUT("/:id/status", middleware.PermissionRequired("users:edit"), controllers.ToggleUserStatus)
	}

	// 角色管理
	roleGroup := protected.Group("/roles")
	{
		roleGroup.GET("", middleware.PermissionRequired("roles:view"), controllers.GetRoleList)
		roleGroup.GET("/all", middleware.PermissionRequired("roles:view"), controllers.GetAllRoles)
		roleGroup.GET("/:id", middleware.PermissionRequired("roles:view"), controllers.GetRole)
		roleGroup.POST("", middleware.PermissionRequired("roles:create"), controllers.CreateRole)
		roleGroup.PUT("/:id", middleware.PermissionRequired("roles:edit"), controllers.UpdateRole)
		roleGroup.DELETE("/:id", middleware.PermissionRequired("roles:delete"), controllers.DeleteRole)
	}

	// 权限管理
	permGroup := protected.Group("/permissions")
	{
		permGroup.GET("", middleware.PermissionRequired("permissions:view"), controllers.GetPermissionList)
		permGroup.GET("/all", middleware.PermissionRequired("permissions:view"), controllers.GetAllPermissions)
		permGroup.GET("/:id", middleware.PermissionRequired("permissions:view"), controllers.GetPermission)
		permGroup.POST("", middleware.PermissionRequired("permissions:create"), controllers.CreatePermission)
		permGroup.PUT("/:id", middleware.PermissionRequired("permissions:edit"), controllers.UpdatePermission)
		permGroup.DELETE("/:id", middleware.PermissionRequired("permissions:delete"), controllers.DeletePermission)
		permGroup.GET("/role/:roleID", middleware.PermissionRequired("permissions:view"), controllers.GetRolePermissions)
		permGroup.POST("/role/:roleID/assign", middleware.PermissionRequired("permissions:edit"), controllers.AssignPermissionsToRole)
	}

	// 日志管理
	logGroup := protected.Group("/logs")
	{
		logGroup.GET("/operation", middleware.PermissionRequired("logs:view"), controllers.GetOperationLogList)
		logGroup.GET("/login", middleware.PermissionRequired("logs:view"), controllers.GetLoginLogList)
		// 清空操作日志属于破坏性高危操作，仅超级管理员可执行
		logGroup.DELETE("/operation", middleware.AdminRequired(), controllers.ClearOperationLog)
	}

	// 字典管理
	dictGroup := protected.Group("/dict")
	{
		dictGroup.GET("/types", middleware.PermissionRequired("dict:view"), controllers.GetDictTypeList)
		dictGroup.GET("/types/:id", middleware.PermissionRequired("dict:view"), controllers.GetDictType)
		dictGroup.POST("/types", middleware.PermissionRequired("dict:create"), controllers.CreateDictType)
		dictGroup.PUT("/types/:id", middleware.PermissionRequired("dict:edit"), controllers.UpdateDictType)
		dictGroup.DELETE("/types/:id", middleware.PermissionRequired("dict:delete"), controllers.DeleteDictType)
		dictGroup.GET("/items", middleware.PermissionRequired("dict:view"), controllers.GetDictItemList)
		dictGroup.POST("/items", middleware.PermissionRequired("dict:create"), controllers.CreateDictItem)
		dictGroup.PUT("/items/:id", middleware.PermissionRequired("dict:edit"), controllers.UpdateDictItem)
		dictGroup.DELETE("/items/:id", middleware.PermissionRequired("dict:delete"), controllers.DeleteDictItem)
		dictGroup.GET("/code/:code", middleware.PermissionRequired("dict:view"), controllers.GetDictItemsByCode)
	}

	// 【gen:routes】
}
