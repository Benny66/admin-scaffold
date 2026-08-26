package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserList 用户列表
func GetUserList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	users, total, err := services.GetUserList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, users, total, page, pageSize)
}

// GetUser 用户详情
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	user, err := services.GetUser(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, user)
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password"`
		RealName string `json:"real_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Status   int    `json:"status"`
		Remark   string `json:"remark"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	status := req.Status
	if status == 0 {
		status = 1
	}
	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   status,
		Remark:   req.Remark,
	}

	if err := services.CreateUser(user, req.RoleIDs); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, user)
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		RealName string `json:"real_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Status   *int   `json:"status"`
		Remark   string `json:"remark"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{
		"real_name": req.RealName,
		"email":     req.Email,
		"phone":     req.Phone,
		"remark":    req.Remark,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := services.UpdateUser(uint(id), updates, req.RoleIDs); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	if err := services.DeleteUser(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ResetPassword 重置密码
func ResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := services.ResetPassword(uint(id), req.Password); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ToggleUserStatus 启用/禁用用户
func ToggleUserStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: status 必须为 0 或 1")
		return
	}

	if err := services.ToggleUserStatus(uint(id), req.Status); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
