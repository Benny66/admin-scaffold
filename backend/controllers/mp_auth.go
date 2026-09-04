package controllers

import (
	"base-backend/services"
	"base-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// MpLoginRequest 微信小程序登录请求
type MpLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// MpLogin 微信小程序登录
//
// 与 Login 同在公开路由组（无需 token 即可访问）。客户端用 wx.login 拿 code 后
// 调本接口，后端用 code 换 openid 后签发与账号密码登录完全相同的 JWT。
//
// 错误码约定：
//   - 400：code 非法 / 已失效（微信侧 errcode != 0），或请求参数缺失
//   - 500：未配置 wechat 段、DB 异常等其他服务端问题
func MpLogin(c *gin.Context) {
	var req MpLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	result, err := services.MpLogin(req.Code)
	if err != nil {
		// 微信侧 code 失效（errmsg 含"微信登录失败"）返回 400；未配置返回 500；其他 500
		code := 500
		if strings.Contains(err.Error(), "微信登录失败") {
			code = 400
		}
		utils.Fail(c, code, err.Error())
		return
	}

	// 记录登录成功日志（与 Login 同路径，但标记来源为 mp-login 以便审计区分）
	services.RecordLoginLog(result.User.ID, result.User.Username, result.User.RealName,
		c.ClientIP(), c.Request.UserAgent(), 1, "mp-login")

	utils.Success(c, gin.H{
		"token":       result.Token,
		"user":        result.User,
		"permissions": result.PermissionCodes,
	})
}
