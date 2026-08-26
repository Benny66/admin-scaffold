package models

// DictType 字典类型模型
type DictType struct {
	BaseModel
	Name        string     `gorm:"size:100;not null" json:"name"`
	Code        string     `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string     `gorm:"size:500" json:"description"`
	Status      int        `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	Sort        int        `gorm:"default:0" json:"sort"`
	Items       []DictItem `gorm:"foreignKey:DictTypeID" json:"items,omitempty"`
}

// DictItem 字典项模型
type DictItem struct {
	BaseModel
	DictTypeID uint   `gorm:"not null" json:"dict_type_id"`
	Label      string `gorm:"size:100;not null" json:"label"`
	Value      string `gorm:"size:100;not null" json:"value"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Status     int    `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	Remark     string `gorm:"size:500" json:"remark"`
}
