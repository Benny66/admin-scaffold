package models

// User 用户模型
type User struct {
	BaseModel
	Username    string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string `gorm:"size:255;not null" json:"-"`
	RealName    string `gorm:"size:50" json:"real_name"`
	Email       string `gorm:"size:100" json:"email"`
	Phone       string `gorm:"size:20" json:"phone"`
	Avatar      string `gorm:"size:255" json:"avatar"`
	Status      int    `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	IsAdmin     bool   `gorm:"default:false" json:"is_admin"`
	Roles       []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	LastLoginAt string `gorm:"size:30" json:"last_login_at"`
	LastLoginIP string `gorm:"size:50" json:"last_login_ip"`
	Remark      string `gorm:"size:500" json:"remark"`
}

// Role 角色模型
type Role struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Code        string `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string `gorm:"size:255" json:"description"`
	Status      int    `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	Sort        int    `gorm:"default:0" json:"sort"`
	Users       []User `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID uint `gorm:"primaryKey" json:"user_id"`
	RoleID uint `gorm:"primaryKey" json:"role_id"`
}

// Permission 权限模型
type Permission struct {
	BaseModel
	Name     string `gorm:"size:100;not null" json:"name"`
	Code     string `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Type     string `gorm:"size:20" json:"type"` // menu:菜单 button:按钮 api:接口
	ParentID uint   `json:"parent_id"`
	Path     string `gorm:"size:255" json:"path"`
	Icon     string `gorm:"size:100" json:"icon"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey" json:"role_id"`
	PermissionID uint `gorm:"primaryKey" json:"permission_id"`
}
