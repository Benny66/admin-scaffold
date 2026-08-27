// 这是「黄金路径」唯一范例的路由注册模板。
// 生成器 gen-module.sh 将本文件内容注入 router/router.go 的 【gen:routes】 锚点前。
// 权限码约定：<资源>:<动作>（view/create/edit/delete），与 database.go 的 initBaseData 对齐。
//
// 示例路由（生成后自动替换 Example/example）：
//
//   exampleGroup := protected.Group("/examples")
//   {
//       exampleGroup.GET("", middleware.PermissionRequired("example:view"), controllers.GetExampleList)
//       exampleGroup.GET("/:id", middleware.PermissionRequired("example:view"), controllers.GetExample)
//       exampleGroup.POST("", middleware.PermissionRequired("example:create"), controllers.CreateExample)
//       exampleGroup.PUT("/:id", middleware.PermissionRequired("example:edit"), controllers.UpdateExample)
//       exampleGroup.DELETE("/:id", middleware.PermissionRequired("example:delete"), controllers.DeleteExample)
//   }
