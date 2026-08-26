package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRoleList 角色列表
func GetRoleList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	roles, total, err := services.GetRoleList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, roles, total, page, pageSize)
}

// GetAllRoles 所有角色
func GetAllRoles(c *gin.Context) {
	roles, err := services.GetAllRoles()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, roles)
}

// GetRole 角色详情
func GetRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的角色ID")
		return
	}

	role, err := services.GetRole(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, role)
}

// CreateRole 创建角色
func CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Sort        int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	role := &models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      1,
	}
	if err := services.CreateRole(role); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, role)
}

// UpdateRole 更新角色
func UpdateRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的角色ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		Sort        *int   `json:"sort"`
		Status      *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := services.UpdateRole(uint(id), updates); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeleteRole 删除角色
func DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的角色ID")
		return
	}

	if err := services.DeleteRole(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
