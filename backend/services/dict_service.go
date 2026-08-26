package services

import (
	"base-backend/database"
	"base-backend/models"
	"errors"
)

// GetDictTypeList 分页查询字典类型列表
func GetDictTypeList(page, pageSize int, keyword string) ([]models.DictType, int64, error) {
	var types []models.DictType
	var total int64

	query := database.DB.Model(&models.DictType{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&types).Error; err != nil {
		return nil, 0, err
	}
	return types, total, nil
}

// GetDictType 获取单个字典类型
func GetDictType(id uint) (*models.DictType, error) {
	var t models.DictType
	if err := database.DB.First(&t, id).Error; err != nil {
		return nil, errors.New("字典类型不存在")
	}
	return &t, nil
}

// CreateDictType 创建字典类型
func CreateDictType(t *models.DictType) error {
	var count int64
	database.DB.Model(&models.DictType{}).Where("code = ?", t.Code).Count(&count)
	if count > 0 {
		return errors.New("字典编码已存在")
	}
	return database.DB.Create(t).Error
}

// UpdateDictType 更新字典类型
func UpdateDictType(id uint, updates map[string]interface{}) error {
	var t models.DictType
	if err := database.DB.First(&t, id).Error; err != nil {
		return errors.New("字典类型不存在")
	}
	return database.DB.Model(&t).Updates(updates).Error
}

// DeleteDictType 删除字典类型（级联删除字典项）
func DeleteDictType(id uint) error {
	var t models.DictType
	if err := database.DB.First(&t, id).Error; err != nil {
		return errors.New("字典类型不存在")
	}
	database.DB.Where("dict_type_id = ?", id).Delete(&models.DictItem{})
	return database.DB.Delete(&t).Error
}

// GetDictItemList 分页查询字典项列表
func GetDictItemList(page, pageSize int, dictTypeID uint) ([]models.DictItem, int64, error) {
	var items []models.DictItem
	var total int64

	query := database.DB.Model(&models.DictItem{})
	if dictTypeID > 0 {
		query = query.Where("dict_type_id = ?", dictTypeID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetDictItemsByCode 根据字典编码获取启用中的字典项（供下拉选项使用）
func GetDictItemsByCode(code string) ([]models.DictItem, error) {
	var t models.DictType
	if err := database.DB.Where("code = ?", code).First(&t).Error; err != nil {
		return nil, errors.New("字典类型不存在")
	}
	var items []models.DictItem
	if err := database.DB.Where("dict_type_id = ? AND status = 1", t.ID).
		Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CreateDictItem 创建字典项
func CreateDictItem(item *models.DictItem) error {
	return database.DB.Create(item).Error
}

// UpdateDictItem 更新字典项
func UpdateDictItem(id uint, updates map[string]interface{}) error {
	var item models.DictItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return errors.New("字典项不存在")
	}
	return database.DB.Model(&item).Updates(updates).Error
}

// DeleteDictItem 删除字典项
func DeleteDictItem(id uint) error {
	var item models.DictItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return errors.New("字典项不存在")
	}
	return database.DB.Delete(&item).Error
}
