package services

import (
	"base-backend/database"
	"base-backend/models"
	"errors"
)

// GetPermissionList 分页查询权限列表
func GetPermissionList(page, pageSize int, keyword string) ([]models.Permission, int64, error) {
	var perms []models.Permission
	var total int64

	query := database.DB.Model(&models.Permission{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&perms).Error; err != nil {
		return nil, 0, err
	}
	return perms, total, nil
}

// GetAllPermissions 获取所有权限
func GetAllPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	if err := database.DB.Order("sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// GetPermission 获取单个权限
func GetPermission(id uint) (*models.Permission, error) {
	var perm models.Permission
	if err := database.DB.First(&perm, id).Error; err != nil {
		return nil, errors.New("权限不存在")
	}
	return &perm, nil
}

// CreatePermission 创建权限
func CreatePermission(perm *models.Permission) error {
	var count int64
	database.DB.Model(&models.Permission{}).Where("code = ?", perm.Code).Count(&count)
	if count > 0 {
		return errors.New("权限编码已存在")
	}
	return database.DB.Create(perm).Error
}

// UpdatePermission 更新权限
func UpdatePermission(id uint, updates map[string]interface{}) error {
	var perm models.Permission
	if err := database.DB.First(&perm, id).Error; err != nil {
		return errors.New("权限不存在")
	}
	return database.DB.Model(&perm).Updates(updates).Error
}

// DeletePermission 删除权限
func DeletePermission(id uint) error {
	var perm models.Permission
	if err := database.DB.First(&perm, id).Error; err != nil {
		return errors.New("权限不存在")
	}
	database.DB.Where("permission_id = ?", id).Delete(&models.RolePermission{})
	return database.DB.Delete(&perm).Error
}

// GetRolePermissions 获取角色已分配的权限ID列表
func GetRolePermissions(roleID uint) ([]uint, error) {
	var rolePerms []models.RolePermission
	if err := database.DB.Where("role_id = ?", roleID).Find(&rolePerms).Error; err != nil {
		return nil, err
	}
	var ids []uint
	for _, rp := range rolePerms {
		ids = append(ids, rp.PermissionID)
	}
	return ids, nil
}

// AssignPermissionsToRole 为角色分配权限（先删后增）
func AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		return errors.New("角色不存在")
	}

	database.DB.Where("role_id = ?", roleID).Delete(&models.RolePermission{})
	for _, pid := range permissionIDs {
		database.DB.Create(&models.RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return nil
}
