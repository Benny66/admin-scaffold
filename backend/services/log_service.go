package services

import (
	"base-backend/database"
	"base-backend/models"
	"time"
)

// GetOperationLogList 分页查询操作日志
func GetOperationLogList(page, pageSize int, keyword string) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	query := database.DB.Model(&models.OperationLog{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR path LIKE ? OR module LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// GetLoginLogList 分页查询登录日志
func GetLoginLogList(page, pageSize int, keyword string) ([]models.LoginLog, int64, error) {
	var logs []models.LoginLog
	var total int64

	query := database.DB.Model(&models.LoginLog{})
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ClearOperationLog 清空操作日志
func ClearOperationLog() error {
	return database.DB.Where("1 = 1").Delete(&models.OperationLog{}).Error
}

// RecordLoginLog 记录登录日志（异步）
func RecordLoginLog(userID uint, username, realName, ip, userAgent string, status int, failMsg string) {
	log := models.LoginLog{
		UserID:    userID,
		Username:  username,
		RealName:  realName,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
		FailMsg:   failMsg,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	go database.DB.Create(&log)
}
