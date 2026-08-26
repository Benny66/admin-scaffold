package services

import (
	"base-backend/database"
	"base-backend/models"
	"errors"
)

// GetRoleList 分页查询角色列表
func GetRoleList(page, pageSize int, keyword string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := database.DB.Model(&models.Role{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

// GetAllRoles 获取所有角色
func GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	if err := database.DB.Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRole 获取单个角色
func GetRole(id uint) (*models.Role, error) {
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		return nil, errors.New("角色不存在")
	}
	return &role, nil
}

// CreateRole 创建角色
func CreateRole(role *models.Role) error {
	var count int64
	database.DB.Model(&models.Role{}).Where("code = ?", role.Code).Count(&count)
	if count > 0 {
		return errors.New("角色编码已存在")
	}
	return database.DB.Create(role).Error
}

// UpdateRole 更新角色
func UpdateRole(id uint, updates map[string]interface{}) error {
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		return errors.New("角色不存在")
	}
	return database.DB.Model(&role).Updates(updates).Error
}

// DeleteRole 删除角色
func DeleteRole(id uint) error {
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		return errors.New("角色不存在")
	}
	if role.Code == "admin" {
		return errors.New("不能删除超级管理员角色")
	}

	database.DB.Where("role_id = ?", id).Delete(&models.UserRole{})
	database.DB.Where("role_id = ?", id).Delete(&models.RolePermission{})
	return database.DB.Delete(&role).Error
}
