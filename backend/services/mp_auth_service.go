package services

import (
	"base-backend/config"
	"base-backend/database"
	"base-backend/models"
	"base-backend/utils"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// MpLogin 用微信小程序 code 完成登录，返回与账号密码登录相同的 LoginResult。
//
// 流程（design D4 + D5）：
//  1. 取 config.Wechat.AppID/Secret；空则返回明确错误（业务码 500，由 controller 判断）
//  2. 调 utils.JsCode2Session 拿 openid（标准库 net/http，不引第三方 SDK）
//  3. 按 openid 查 User：未找到则创建（username=mp_<openid 前 8 位>，随机密码——
//     小程序用户不会用 username/password 登录；绑定「普通用户」角色，与默认账号对齐）
//  4. 签发与 /api/auth/login 完全相同的 JWT（同一 GenerateToken、同一过期时间）
//  5. 返回 { token, user, permissions }，与原 Login 结构一致
//
// 鉴权衔接：mp-login 签发的 token 与 username/password 登录签发的 token 等价，
// 后续 JWTAuth / PermissionRequired / AdminRequired 中间件不区分 token 来源。
// 与 in-progress 的 auth-token-lifecycle 协同：等其落地后 token 自动带 Ver 字段，
// 享受版本吊销与刷新令牌能力（见 design D10）。
func MpLogin(code string) (*LoginResult, error) {
	appid := config.GlobalConfig.Wechat.AppID
	secret := config.GlobalConfig.Wechat.Secret
	if appid == "" || secret == "" {
		return nil, errors.New("微信小程序登录未配置：请在 config.yaml 的 wechat 段填写 app_id 与 secret")
	}

	openid, _, err := utils.JsCode2Session(appid, secret, code)
	if err != nil {
		return nil, err
	}

	user, created, err := findOrCreateMpUser(openid)
	if err != nil {
		return nil, err
	}

	if user.Status == 0 {
		return nil, errors.New("账号已被禁用，请联系管理员")
	}

	if created {
		// 新建用户绑定「普通用户」角色（与 initBaseData 的默认账号对齐，避免越权）
		bindDefaultMpRole(user.ID)
	}

	// 签发与 /api/auth/login 完全相同的 JWT（同一 GenerateToken、同一过期时间）
	token, err := utils.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, errors.New("令牌生成失败")
	}

	// 更新最后登录信息（与 Login 同路径）
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
	})

	// 查询用户角色与权限码（与 Login 同路径，保证 RBAC 一致）
	var roles []models.Role
	database.DB.Model(&user).Association("Roles").Find(&roles)
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

// findOrCreateMpUser 按 openid 查 User；不存在则创建。返回 (user, created, err)。
//
// 用 strings.Contains(err.Error(), "record not found") 区分「未找到」与「其他 DB 错误」，
// 与既有 services 不 import gorm 的风格保持一致（auth_service.go 同款写法）。
// 并发场景：两个请求同时拿同一 openid 时，唯一索引兜底，失败的请求回查并返回既有记录。
func findOrCreateMpUser(openid string) (models.User, bool, error) {
	var user models.User
	if err := database.DB.Where("open_id = ?", openid).First(&user).Error; err == nil {
		return user, false, nil
	} else if !strings.Contains(err.Error(), "record not found") {
		return models.User{}, false, fmt.Errorf("查询微信用户失败: %w", err)
	}

	// 不存在则创建：username 形如 mp_<openid 前 8 位>，密码随机（用户永不走 username/password 登录）
	suffix := openid
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}

	pwdBytes := make([]byte, 32)
	if _, err := rand.Read(pwdBytes); err != nil {
		return models.User{}, false, fmt.Errorf("生成随机密码失败: %w", err)
	}
	hash, err := utils.HashPassword(hex.EncodeToString(pwdBytes))
	if err != nil {
		return models.User{}, false, fmt.Errorf("密码加密失败: %w", err)
	}

	user = models.User{
		Username: "mp_" + suffix,
		Password: hash,
		RealName: "微信用户",
		Status:   1,
		IsAdmin:  false,
		OpenID:   openid,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		// 并发场景：另一请求已创建同 openid 用户——回查一次
		var existing models.User
		if retryErr := database.DB.Where("open_id = ?", openid).First(&existing).Error; retryErr == nil {
			return existing, false, nil
		}
		return models.User{}, false, fmt.Errorf("创建微信用户失败: %w", err)
	}
	return user, true, nil
}

// bindDefaultMpRole 把新微信用户绑定到 code="user" 的「普通用户」角色，
// 与 initBaseData 创建的默认账号对齐，避免越权。
func bindDefaultMpRole(userID uint) {
	var userRole models.Role
	if err := database.DB.Where("code = ?", "user").First(&userRole).Error; err == nil && userRole.ID > 0 {
		database.DB.Create(&models.UserRole{UserID: userID, RoleID: userRole.ID})
	}
}
