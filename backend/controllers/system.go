package controllers

import (
	"base-backend/config"
	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetSystemInfo 获取系统信息（供前端展示品牌名称、logo、favicon、页脚）
func GetSystemInfo(c *gin.Context) {
	app := config.GlobalConfig.App

	name := app.Name
	if name == "" {
		name = app.Subtitle
	}
	if name == "" {
		name = "Base Admin"
	}

	// logo/favicon：文件名拼成静态路径；favicon 缺省回退 logo
	logo := staticURL(app.Logo)
	favicon := staticURL(app.Favicon)
	if favicon == "" {
		favicon = logo
	}

	utils.Success(c, gin.H{
		"name":            name,
		"subtitle":        app.Subtitle,
		"logo":            logo,
		"favicon":         favicon,
		"footer":          app.Footer,
		"login_bg":        staticURL(app.LoginBg),
		"login_bg_mobile": staticURL(app.LoginBgMobile),
	})
}

// staticURL 把配置里的文件名拼成静态资源路径；空文件名返回空串。
func staticURL(file string) string {
	if file == "" {
		return ""
	}
	return "/static/" + file
}
