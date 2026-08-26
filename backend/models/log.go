package models

// OperationLog 操作日志模型
type OperationLog struct {
	BaseModel
	UserID    uint   `json:"user_id"`
	Username  string `gorm:"size:50" json:"username"`
	RealName  string `gorm:"size:50" json:"real_name"`
	Module    string `gorm:"size:100" json:"module"`   // 操作模块
	Action    string `gorm:"size:100" json:"action"`   // 操作动作
	Method    string `gorm:"size:10" json:"method"`    // HTTP方法
	Path      string `gorm:"size:255" json:"path"`     // 请求路径
	IP        string `gorm:"size:50" json:"ip"`        // 客户端IP
	UserAgent string `gorm:"size:500" json:"user_agent"`
	ReqBody   string `gorm:"type:text" json:"req_body"` // 请求体
	RespCode  int    `json:"resp_code"`                 // 响应状态码
	Duration  int64  `json:"duration"`                  // 耗时(ms)
	Status    int    `gorm:"default:1" json:"status"`   // 1:成功 0:失败
	ErrorMsg  string `gorm:"size:1000" json:"error_msg"` // 错误信息
}

// LoginLog 登录日志模型
type LoginLog struct {
	BaseModel
	UserID    uint   `json:"user_id"`
	Username  string `gorm:"size:50" json:"username"`
	RealName  string `gorm:"size:50" json:"real_name"`
	IP        string `gorm:"size:50" json:"ip"`
	UserAgent string `gorm:"size:500" json:"user_agent"`
	Status    int    `gorm:"default:1" json:"status"` // 1:成功 0:失败
	FailMsg   string `gorm:"size:500" json:"fail_msg"`
	LoginTime string `gorm:"size:30" json:"login_time"`
}
