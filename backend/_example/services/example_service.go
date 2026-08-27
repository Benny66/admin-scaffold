// 这是「黄金路径」唯一范例的 service 层模板。
// 生成器 gen-module.sh 将本文件复制为 services/<name>_service.go，替换占位符。
// 分层铁律：service 承载业务逻辑，通过 GORM 访问 DB，不触碰 gin.Context。
package services

import (
	"base-backend/database"
	"base-backend/models"
	"errors"
)

// GetExampleList 分页查询示例列表
func GetExampleList(page, pageSize int, keyword string) ([]models.Example, int64, error) {
	var list []models.Example
	var total int64

	query := database.DB.Model(&models.Example{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetExample 获取单个示例
func GetExample(id uint) (*models.Example, error) {
	var item models.Example
	if err := database.DB.First(&item, id).Error; err != nil {
		return nil, errors.New("示例不存在")
	}
	return &item, nil
}

// CreateExample 创建示例
func CreateExample(item *models.Example) error {
	// TODO: 业务逻辑 —— 唯一性校验、默认值、关联处理等
	return database.DB.Create(item).Error
}

// UpdateExample 更新示例
func UpdateExample(id uint, updates map[string]interface{}) error {
	var item models.Example
	if err := database.DB.First(&item, id).Error; err != nil {
		return errors.New("示例不存在")
	}
	// TODO: 业务逻辑 —— 字段校验、关联更新等
	return database.DB.Model(&item).Updates(updates).Error
}

// DeleteExample 删除示例
func DeleteExample(id uint) error {
	var item models.Example
	if err := database.DB.First(&item, id).Error; err != nil {
		return errors.New("示例不存在")
	}
	// TODO: 业务逻辑 —— 级联清理、保护性校验等
	return database.DB.Delete(&item).Error
}
