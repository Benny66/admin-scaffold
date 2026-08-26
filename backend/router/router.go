package router

import (
	"base-backend/controllers"
	"base-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有路由
func SetupRouter(r *gin.Engine) {
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

	// 用户相关
	protected.GET("/auth/info", controllers.GetUserInfo)
	protected.PUT("/auth/password", controllers.ChangePassword)

	// ==================== 系统管理 ====================
	// 用户管理
	userGroup := protected.Group("/users")
	{
		userGroup.GET("", controllers.GetUserList)
		userGroup.GET("/:id", controllers.GetUser)
		userGroup.POST("", controllers.CreateUser)
		userGroup.PUT("/:id", controllers.UpdateUser)
		userGroup.DELETE("/:id", controllers.DeleteUser)
		userGroup.PUT("/:id/password", controllers.ResetPassword)
		userGroup.PUT("/:id/status", controllers.ToggleUserStatus)
	}

	// 角色管理
	roleGroup := protected.Group("/roles")
	{
		roleGroup.GET("", controllers.GetRoleList)
		roleGroup.GET("/all", controllers.GetAllRoles)
		roleGroup.GET("/:id", controllers.GetRole)
		roleGroup.POST("", controllers.CreateRole)
		roleGroup.PUT("/:id", controllers.UpdateRole)
		roleGroup.DELETE("/:id", controllers.DeleteRole)
	}

	// 权限管理
	permGroup := protected.Group("/permissions")
	{
		permGroup.GET("", controllers.GetPermissionList)
		permGroup.GET("/all", controllers.GetAllPermissions)
		permGroup.GET("/:id", controllers.GetPermission)
		permGroup.POST("", controllers.CreatePermission)
		permGroup.PUT("/:id", controllers.UpdatePermission)
		permGroup.DELETE("/:id", controllers.DeletePermission)
		permGroup.GET("/role/:roleID", controllers.GetRolePermissions)
		permGroup.POST("/role/:roleID/assign", controllers.AssignPermissionsToRole)
	}

	// 日志管理
	logGroup := protected.Group("/logs")
	{
		logGroup.GET("/operation", controllers.GetOperationLogList)
		logGroup.GET("/login", controllers.GetLoginLogList)
		logGroup.DELETE("/operation", controllers.ClearOperationLog)
	}

	// 字典管理
	dictGroup := protected.Group("/dict")
	{
		dictGroup.GET("/types", controllers.GetDictTypeList)
		dictGroup.GET("/types/:id", controllers.GetDictType)
		dictGroup.POST("/types", controllers.CreateDictType)
		dictGroup.PUT("/types/:id", controllers.UpdateDictType)
		dictGroup.DELETE("/types/:id", controllers.DeleteDictType)
		dictGroup.GET("/items", controllers.GetDictItemList)
		dictGroup.POST("/items", controllers.CreateDictItem)
		dictGroup.PUT("/items/:id", controllers.UpdateDictItem)
		dictGroup.DELETE("/items/:id", controllers.DeleteDictItem)
		dictGroup.GET("/code/:code", controllers.GetDictItemsByCode)
	}
}
