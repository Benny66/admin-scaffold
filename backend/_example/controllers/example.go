// 这是「黄金路径」唯一范例的 controller 层模板。
// 生成器 gen-module.sh 将本文件复制为 controllers/<name>.go，替换占位符。
// 分层铁律：controller 只做参数校验 + 调 service + 组装响应，不写业务逻辑。
package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetExampleList 示例列表
func GetExampleList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	list, total, err := services.GetExampleList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, page, pageSize)
}

// GetExample 示例详情
func GetExample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的示例ID")
		return
	}

	item, err := services.GetExample(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, item)
}

// CreateExample 创建示例
func CreateExample(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	status := req.Status
	if status == 0 {
		status = 1
	}
	item := &models.Example{
		Name:   req.Name,
		Status: status,
	}
	if err := services.CreateExample(item); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, item)
}

// UpdateExample 更新示例
func UpdateExample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的示例ID")
		return
	}

	var req struct {
		Name   string `json:"name"`
		Status *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := services.UpdateExample(uint(id), updates); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeleteExample 删除示例
func DeleteExample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的示例ID")
		return
	}

	if err := services.DeleteExample(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
