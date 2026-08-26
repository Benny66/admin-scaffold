package middleware

import (
	"base-backend/database"
	"base-backend/models"
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter 自定义响应写入器，用于捕获响应状态码
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// OperationLogger 操作日志中间件
func OperationLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要记录的路径
		path := c.Request.URL.Path
		if path == "/api/auth/login" {
			c.Next()
			return
		}

		start := time.Now()

		// 读取请求体
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 自定义响应写入器
		rw := &responseWriter{ResponseWriter: c.Writer, statusCode: 200}
		c.Writer = rw

		c.Next()

		duration := time.Since(start).Milliseconds()

		// 获取用户信息
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")

		uid := uint(0)
		uname := ""
		if userID != nil {
			uid = userID.(uint)
		}
		if username != nil {
			uname = username.(string)
		}

		// 只记录写操作
		method := c.Request.Method
		if method == "GET" {
			return
		}

		// 异步写入日志
		go func() {
			log := models.OperationLog{
				UserID:    uid,
				Username:  uname,
				Method:    method,
				Path:      path,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				ReqBody:   string(reqBody),
				RespCode:  rw.statusCode,
				Duration:  duration,
				Status:    1,
			}
			if rw.statusCode >= 400 {
				log.Status = 0
			}
			database.DB.Create(&log)
		}()
	}
}
