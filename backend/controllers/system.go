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
		name = "企业管理系统"
	}
	utils.Success(c, gin.H{
		"name":     name,
		"subtitle": config.GlobalConfig.App.Subtitle,
	})
}
