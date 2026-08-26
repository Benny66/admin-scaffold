package main

import (
	"base-backend/config"
	"base-backend/database"
	"base-backend/middleware"
	"base-backend/router"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin运行模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 初始化数据库
	database.Init()

	// 创建Gin引擎
	r := gin.New()

	// 注册全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// 注册路由
	router.SetupRouter(r)

	// 启动服务
	addr := ":" + config.GlobalConfig.Server.Port
	log.Printf("Base Backend 启动成功，监听端口: %s", addr)
	log.Printf("默认管理员账号: admin / admin123")

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
