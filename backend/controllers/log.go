package controllers

import (
	"base-backend/services"
	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetOperationLogList 操作日志列表
func GetOperationLogList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	logs, total, err := services.GetOperationLogList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, logs, total, page, pageSize)
}

// GetLoginLogList 登录日志列表
func GetLoginLogList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	logs, total, err := services.GetLoginLogList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, logs, total, page, pageSize)
}

// ClearOperationLog 清空操作日志
func ClearOperationLog(c *gin.Context) {
	if err := services.ClearOperationLog(); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
