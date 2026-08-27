// 这是「黄金路径」唯一范例的 model 层模板。
// 生成器 gen-module.sh 将本文件复制为 models/<name>.go，并把 Example→<Pascal>、example→<name>。
// 模型约定：内嵌 BaseModel，字段带 snake_case 的 JSON tag（见 AGENTS.md 第 6/7 条）。
package models

// Example 示例模型
type Example struct {
	BaseModel
	Name   string `gorm:"size:100;not null" json:"name"`
	Status int    `gorm:"default:1" json:"status"` // 1:启用 0:禁用
}
