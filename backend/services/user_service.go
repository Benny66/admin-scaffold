package services

import (
	"base-backend/database"
	"base-backend/models"
	"base-backend/utils"
	"errors"
)

// GetUserList 分页查询用户列表
func GetUserList(page, pageSize int, keyword string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := database.DB.Model(&models.User{}).Preload("Roles")
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// GetUser 获取单个用户
func GetUser(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

// CreateUser 创建用户
func CreateUser(user *models.User, roleIDs []uint) error {
	// 检查用户名唯一
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 密码加密
	if user.Password == "" {
		user.Password = "123456"
	}
	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash

	if err := database.DB.Create(user).Error; err != nil {
		return err
	}

	// 绑定角色
	if len(roleIDs) > 0 {
		for _, roleID := range roleIDs {
			database.DB.Create(&models.UserRole{UserID: user.ID, RoleID: roleID})
		}
	}
	return nil
}

// UpdateUser 更新用户（不含密码）
func UpdateUser(id uint, updates map[string]interface{}, roleIDs []uint) error {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return err
	}

	// 更新角色（先删后增）
	if roleIDs != nil {
		database.DB.Where("user_id = ?", id).Delete(&models.UserRole{})
		for _, roleID := range roleIDs {
			database.DB.Create(&models.UserRole{UserID: id, RoleID: roleID})
		}
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(id uint) error {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.IsAdmin {
		return errors.New("不能删除超级管理员")
	}

	database.DB.Where("user_id = ?", id).Delete(&models.UserRole{})
	return database.DB.Delete(&user).Error
}

// ResetPassword 重置密码
func ResetPassword(id uint, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return database.DB.Model(&user).Update("password", hash).Error
}

// ToggleUserStatus 启用/禁用用户
func ToggleUserStatus(id uint, status int) error {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.IsAdmin && status == 0 {
		return errors.New("不能禁用超级管理员")
	}
	return database.DB.Model(&user).Update("status", status).Error
}
