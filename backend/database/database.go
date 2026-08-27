package database

import (
	"base-backend/config"
	"base-backend/models"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库连接并自动迁移建表
func Init() {
	var err error

	dbType := config.GlobalConfig.Database.Type

	switch dbType {
	case "mysql":
		dsn := config.GlobalConfig.Database.MySQL.DSN()
		log.Printf("连接 MySQL 数据库: %s@%s/%s", config.GlobalConfig.Database.MySQL.Username,
			config.GlobalConfig.Database.MySQL.Host, config.GlobalConfig.Database.MySQL.Database)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatalf("MySQL 连接失败: %v\n请检查 config.yaml 中的 mysql 配置", err)
		}
		log.Println("MySQL 数据库连接成功")

	default: // sqlite
		dsn := config.GlobalConfig.Database.SQLite.DSN
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("无法创建数据库目录 %s: %v", dir, err)
		}

		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatalf("SQLite 连接失败: %v", err)
		}

		// 启用WAL模式提升SQLite并发性能
		if err := DB.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			log.Printf("警告: 无法启用WAL模式（目录可能只读），使用默认journal模式: %v", err)
		}
		DB.Exec("PRAGMA foreign_keys=ON;")
		log.Println("SQLite 数据库连接成功")
	}

	// 临时关闭外键约束（仅 SQLite 需要），避免 AutoMigrate 重建表时外键校验失败
	if dbType == "sqlite" {
		DB.Exec("PRAGMA foreign_keys=OFF;")
	}

	// 自动迁移建表（系统管理五件套相关表）
	err = DB.AutoMigrate(
		// 【gen:migrate】
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Permission{},
		&models.RolePermission{},
		&models.DictType{},
		&models.DictItem{},
		&models.OperationLog{},
		&models.LoginLog{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 迁移完成后重新开启外键约束（仅 SQLite）
	if dbType == "sqlite" {
		DB.Exec("PRAGMA foreign_keys=ON;")
	}

	log.Printf("数据库初始化成功，所有表已自动创建（类型: %s）", dbType)

	// 初始化基础数据
	initBaseData()
}

// initBaseData 初始化基础数据（管理员账号、默认角色、权限、字典）
func initBaseData() {
	// 创建默认角色
	var roleCount int64
	DB.Model(&models.Role{}).Count(&roleCount)
	if roleCount == 0 {
		roles := []models.Role{
			{Name: "超级管理员", Code: "admin", Description: "系统超级管理员，拥有所有权限", Status: 1},
			{Name: "普通用户", Code: "user", Description: "普通用户，只读权限", Status: 1},
		}
		DB.Create(&roles)
	}

	// 创建默认管理员用户
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser := models.User{
			Username: "admin",
			Password: string(hash),
			RealName: "系统管理员",
			Email:    "admin@company.com",
			Phone:    "13800138000",
			Status:   1,
			IsAdmin:  true,
		}
		DB.Create(&adminUser)

		// 绑定管理员角色
		var adminRole models.Role
		DB.Where("code = ?", "admin").First(&adminRole)
		if adminRole.ID > 0 {
			DB.Create(&models.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID})
		}
	}

	// 创建默认权限（仅系统管理相关）
	var permCount int64
	DB.Model(&models.Permission{}).Count(&permCount)
	if permCount == 0 {
		permissions := []models.Permission{
			// 用户管理权限
			{Name: "查看用户", Code: "users:view", Type: "api", Sort: 1, Status: 1},
			{Name: "创建用户", Code: "users:create", Type: "api", Sort: 2, Status: 1},
			{Name: "编辑用户", Code: "users:edit", Type: "api", Sort: 3, Status: 1},
			{Name: "删除用户", Code: "users:delete", Type: "api", Sort: 4, Status: 1},
			// 角色管理权限
			{Name: "查看角色", Code: "roles:view", Type: "api", Sort: 5, Status: 1},
			{Name: "创建角色", Code: "roles:create", Type: "api", Sort: 6, Status: 1},
			{Name: "编辑角色", Code: "roles:edit", Type: "api", Sort: 7, Status: 1},
			{Name: "删除角色", Code: "roles:delete", Type: "api", Sort: 8, Status: 1},
			// 权限管理权限
			{Name: "查看权限", Code: "permissions:view", Type: "api", Sort: 9, Status: 1},
			{Name: "创建权限", Code: "permissions:create", Type: "api", Sort: 10, Status: 1},
			{Name: "编辑权限", Code: "permissions:edit", Type: "api", Sort: 11, Status: 1},
			{Name: "删除权限", Code: "permissions:delete", Type: "api", Sort: 12, Status: 1},
			// 字典权限
			{Name: "查看字典", Code: "dict:view", Type: "api", Sort: 13, Status: 1},
			{Name: "创建字典", Code: "dict:create", Type: "api", Sort: 14, Status: 1},
			{Name: "编辑字典", Code: "dict:edit", Type: "api", Sort: 15, Status: 1},
			{Name: "删除字典", Code: "dict:delete", Type: "api", Sort: 16, Status: 1},
			// 日志权限
			{Name: "查看日志", Code: "logs:view", Type: "api", Sort: 17, Status: 1},
		}
		DB.Create(&permissions)
	}

	// 创建默认字典类型（通用示例，非业务绑定）
	var dictCount int64
	DB.Model(&models.DictType{}).Count(&dictCount)
	if dictCount == 0 {
		dictTypes := []models.DictType{
			{Name: "用户状态", Code: "user_status", Description: "用户启用状态", Status: 1},
			{Name: "通用状态", Code: "common_status", Description: "通用启用禁用状态", Status: 1},
		}
		DB.Create(&dictTypes)

		dictItems := []models.DictItem{
			{DictTypeID: 1, Label: "启用", Value: "1", Sort: 1, Status: 1},
			{DictTypeID: 1, Label: "禁用", Value: "0", Sort: 2, Status: 1},
			{DictTypeID: 2, Label: "是", Value: "1", Sort: 1, Status: 1},
			{DictTypeID: 2, Label: "否", Value: "0", Sort: 2, Status: 1},
		}
		DB.Create(&dictItems)
	}

	// 为超级管理员角色分配所有权限
	var adminRole models.Role
	DB.Where("code = ?", "admin").First(&adminRole)
	if adminRole.ID > 0 {
		var rolePermCount int64
		DB.Model(&models.RolePermission{}).Where("role_id = ?", adminRole.ID).Count(&rolePermCount)
		if rolePermCount == 0 {
			var permissions []models.Permission
			DB.Find(&permissions)
			for _, perm := range permissions {
				DB.Create(&models.RolePermission{RoleID: adminRole.ID, PermissionID: perm.ID})
			}
		}
	}
}
