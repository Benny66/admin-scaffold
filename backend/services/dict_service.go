package services

import (
	"base-backend/database"
	"base-backend/models"
	"errors"
	"strings"
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

// MaxImportRows 导入行数上限（字典项规模小，同步处理；超过即拒绝防止阻塞请求）
const MaxImportRows = 500

// GetAllDictTypesWithItems 一次取出全量字典类型及其字典项（供全量导出）
func GetAllDictTypesWithItems() ([]models.DictType, error) {
	var types []models.DictType
	if err := database.DB.Order("sort ASC, id ASC").Find(&types).Error; err != nil {
		return nil, err
	}
	// 逐类型加载字典项，避免 gorm Preload 对同属主记录关联匹配的隐式全表扫描
	for i := range types {
		if err := database.DB.Where("dict_type_id = ?", types[i].ID).
			Order("sort ASC, id ASC").Find(&types[i].Items).Error; err != nil {
			return nil, err
		}
	}
	return types, nil
}

// GetDictItemsAll 取单个类型的全部字典项（供单类型导出/模板）
func GetDictItemsAll(typeID uint) ([]models.DictItem, error) {
	var items []models.DictItem
	if err := database.DB.Where("dict_type_id = ?", typeID).
		Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// DictItemImportRow 导入行结构，与 xlsx 列头一一对应：label | value | sort | status | remark
type DictItemImportRow struct {
	Label  string
	Value  string
	Sort   int
	Status int
	Remark string
}

// ImportDictItems 按 value 覆盖合并导入字典项：
// value 已存在（同属当前 dict_type_id）→ 更新 label/sort/status/remark；不存在 → 新增。
// 文件未涉及的既有项保留（非全量替换）。返回新增/更新条数。
// 行数超上限或出现非法 status 时整体拒绝，不做部分写入。
func ImportDictItems(typeID uint, rows []DictItemImportRow) (newCount, updatedCount int, err error) {
	if len(rows) > MaxImportRows {
		return 0, 0, errors.New("导入行数超过上限")
	}

	// 一次性取出当前类型的既有项，建立 value → 记录的索引
	var existing []models.DictItem
	if err := database.DB.Where("dict_type_id = ?", typeID).Find(&existing).Error; err != nil {
		return 0, 0, err
	}
	byValue := make(map[string]*models.DictItem, len(existing))
	for i := range existing {
		byValue[existing[i].Value] = &existing[i]
	}

	for _, r := range rows {
		// 必填字段校验：label/value 任一缺失则跳过该行
		if strings.TrimSpace(r.Label) == "" || strings.TrimSpace(r.Value) == "" {
			continue
		}
		// 状态校验：仅允许 0/1（与模型 status 语义一致）
		if r.Status != 0 && r.Status != 1 {
			return 0, 0, errors.New("状态字段仅允许 0 或 1")
		}
		if item, ok := byValue[r.Value]; ok {
			if err := database.DB.Model(item).Updates(map[string]interface{}{
				"label":  r.Label,
				"sort":   r.Sort,
				"status": r.Status,
				"remark": r.Remark,
			}).Error; err != nil {
				return 0, 0, err
			}
			updatedCount++
		} else {
			item := &models.DictItem{
				DictTypeID: typeID,
				Label:      r.Label,
				Value:      r.Value,
				Sort:       r.Sort,
				Status:     r.Status,
				Remark:     r.Remark,
			}
			if err := database.DB.Create(item).Error; err != nil {
				return 0, 0, err
			}
			byValue[r.Value] = item
			newCount++
		}
	}
	return newCount, updatedCount, nil
}
