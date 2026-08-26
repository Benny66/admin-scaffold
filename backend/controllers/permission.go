package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPermissionList 权限列表
func GetPermissionList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	perms, total, err := services.GetPermissionList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, perms, total, page, pageSize)
}

// GetAllPermissions 所有权限
func GetAllPermissions(c *gin.Context) {
	perms, err := services.GetAllPermissions()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, perms)
}

// GetPermission 权限详情
func GetPermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的权限ID")
		return
	}

	perm, err := services.GetPermission(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, perm)
}

// CreatePermission 创建权限
func CreatePermission(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Type     string `json:"type"`
		ParentID uint   `json:"parent_id"`
		Path     string `json:"path"`
		Icon     string `json:"icon"`
		Sort     int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	permType := req.Type
	if permType == "" {
		permType = "api"
	}
	perm := &models.Permission{
		Name:     req.Name,
		Code:     req.Code,
		Type:     permType,
		ParentID: req.ParentID,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   1,
	}
	if err := services.CreatePermission(perm); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, perm)
}

// UpdatePermission 更新权限
func UpdatePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的权限ID")
		return
	}

	var req struct {
		Name     string `json:"name"`
		Code     string `json:"code"`
		Type     string `json:"type"`
		ParentID *uint  `json:"parent_id"`
		Path     string `json:"path"`
		Icon     string `json:"icon"`
		Sort     *int   `json:"sort"`
		Status   *int   `json:"status"`
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
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := services.UpdatePermission(uint(id), updates); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeletePermission 删除权限
func DeletePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的权限ID")
		return
	}

	if err := services.DeletePermission(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetRolePermissions 获取角色已分配的权限ID列表
func GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		utils.Fail(c, 400, "无效的角色ID")
		return
	}

	ids, err := services.GetRolePermissions(uint(roleID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, ids)
}

// AssignPermissionsToRole 为角色分配权限
func AssignPermissionsToRole(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		utils.Fail(c, 400, "无效的角色ID")
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := services.AssignPermissionsToRole(uint(roleID), req.PermissionIDs); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
