package controllers

import (
	"base-backend/services"
	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	result, err := services.Login(req.Username, req.Password)
	if err != nil {
		// 记录登录失败日志
		services.RecordLoginLog(0, req.Username, "", c.ClientIP(), c.Request.UserAgent(), 0, err.Error())
		utils.Fail(c, 401, err.Error())
		return
	}

	// 记录登录成功日志
	services.RecordLoginLog(result.User.ID, result.User.Username, result.User.RealName,
		c.ClientIP(), c.Request.UserAgent(), 1, "")

	utils.Success(c, gin.H{
		"token":       result.Token,
		"user":        result.User,
		"permissions": result.PermissionCodes,
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	utils.Success(c, nil)
}

// GetUserInfo 获取当前用户信息
func GetUserInfo(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	user, err := services.GetUser(userID)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}

	permissions, err := services.GetUserPermissions(userID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"user":        user,
		"permissions": permissions,
	})
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetUint("userID")
	if err := services.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}
