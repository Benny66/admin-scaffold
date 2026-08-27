package controllers

import (
	"base-backend/config"
	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetSystemInfo 获取系统信息（供前端展示系统名称）
func GetSystemInfo(c *gin.Context) {
	name := config.GlobalConfig.App.Name
	if name == "" {
		name = config.GlobalConfig.App.Subtitle
	}
	if name == "" {
		name = "Base Admin"
	}
	utils.Success(c, gin.H{
		"name":     name,
		"subtitle": config.GlobalConfig.App.Subtitle,
	})
}
