package services

import (
	"base-backend/database"
	"base-backend/models"
	"base-backend/utils"
	"errors"
	"time"
)

// LoginResult 登录结果
type LoginResult struct {
	Token           string   `json:"token"`
	User            ginUser  `json:"user"`
	PermissionCodes []string `json:"permissions"`
}

// ginUser 登录返回的用户信息（脱敏）
type ginUser struct {
	ID       uint         `json:"id"`
	Username string       `json:"username"`
	RealName string       `json:"real_name"`
	Email    string       `json:"email"`
	Phone    string       `json:"phone"`
	Avatar   string       `json:"avatar"`
	IsAdmin  bool         `json:"is_admin"`
	Roles    []models.Role `json:"roles"`
}

// Login 用户登录，返回 token 与用户信息；失败时返回错误
func Login(username, password string) (*LoginResult, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status == 0 {
		return nil, errors.New("账号已被禁用，请联系管理员")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成 JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, errors.New("令牌生成失败")
	}

	// 更新最后登录信息
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
	})

	// 查询用户角色
	var roles []models.Role
	database.DB.Model(&user).Association("Roles").Find(&roles)

	// 查询用户权限
	permissionCodes := getUserPermissionCodes(user)

	return &LoginResult{
		Token: token,
		User: ginUser{
			ID:       user.ID,
			Username: user.Username,
			RealName: user.RealName,
			Email:    user.Email,
			Phone:    user.Phone,
			Avatar:   user.Avatar,
			IsAdmin:  user.IsAdmin,
			Roles:    roles,
		},
		PermissionCodes: permissionCodes,
	}, nil
}

// GetUserPermissions 获取用户的权限码列表
func GetUserPermissions(userID uint) ([]string, error) {
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return getUserPermissionCodes(user), nil
}

// ChangePassword 修改密码
func ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	return database.DB.Model(&user).Update("password", hash).Error
}

// getUserPermissionCodes 计算用户的权限码列表
func getUserPermissionCodes(user models.User) []string {
	var roles []models.Role
	database.DB.Model(&user).Association("Roles").Find(&roles)

	var roleIDs []uint
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	var permissions []models.Permission
	if len(roleIDs) > 0 {
		database.DB.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
			Where("role_permissions.role_id IN ?", roleIDs).
			Find(&permissions)
	}

	var codes []string
	for _, perm := range permissions {
		codes = append(codes, perm.Code)
	}
	return codes
}
