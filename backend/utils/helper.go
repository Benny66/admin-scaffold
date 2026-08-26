package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行bcrypt加密
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetPageParams 获取分页参数
func GetPageParams(pageStr, pageSizeStr string) (int, int) {
	page := 1
	pageSize := 10
	if p := parseInt(pageStr); p > 0 {
		page = p
	}
	if ps := parseInt(pageSizeStr); ps > 0 && ps <= 100 {
		pageSize = ps
	}
	return page, pageSize
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
