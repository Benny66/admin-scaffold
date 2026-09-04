package router

import (
	"os"
	"path/filepath"
	"strings"

	"base-backend/utils"

	"github.com/gin-gonic/gin"
)

// setupStaticServing 在 dist/dist-mobile 存在时托管前端产物，并提供 SPA 回退。
// 产物不存在时优雅降级为纯 API，保持「go run 零配置开发」体验不变。
//
// 注意：桌面端不能用 r.Static("/", ...) —— 它注册的是 catch-all 通配 /*filepath，
// 会与 /api 前缀在 httprouter 中冲突并 panic。故改用 r.NoRoute（仅在无匹配路由时
// 触发，天然不与 /api、/static、/m 冲突）手动服务文件。
func setupStaticServing(r *gin.Engine) {
	hasDist := dirExists("./dist")
	hasMobile := dirExists("./dist-mobile")

	// 移动端 H5 产物托管（/m/ 前缀不占根级，不与 /api 冲突）。
	// 不用 r.Static：Static 对 /m/<history 子路由>（文件不存在）直接 404，
	// 移动端需要与桌面端对称的 SPA 回退，故走自定义 handler（见 serveMobile）。
	if hasMobile {
		m := r.Group("/m")
		m.GET("/*filepath", serveMobile)
	}

	// 桌面端产物托管 + SPA 回退（仅当 dist 存在时启用）
	if hasDist {
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			// API、品牌静态资源、移动端前缀不参与桌面端 SPA 回退
			if strings.HasPrefix(p, "/api/") ||
				strings.HasPrefix(p, "/static/") ||
				strings.HasPrefix(p, "/m/") {
				utils.Fail(c, 404, "资源不存在")
				return
			}

			// 请求的是 dist 下的具体文件（如 /assets/xxx.js），直接返回
			fsPath := filepath.Join("dist", filepath.Clean(p))
			if stat, err := os.Stat(fsPath); err == nil && !stat.IsDir() {
				c.File(fsPath)
				return
			}

			// 其余（前端 history 路由）回退到 index.html
			c.File("dist/index.html")
		})
	}
}

// dirExists 判断目录是否存在（用于条件托管，避免无产物时误启用静态路由）。
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// serveMobile 服务移动端 H5 单页：dist-mobile/ 中存在目标文件则原样返回
//（含 /m/assets/* 静态资源，构建产物以 /m/ 为 base 引用）；否则回退返回
// dist-mobile/index.html，使移动端 history 子路由（如 /m/login）直达/刷新不 404，
// 与桌面 NoRoute 的 SPA 回退对称。
func serveMobile(c *gin.Context) {
	// 资源不存在时返回 JSON 404（复用统一响应，勿给资源请求回 HTML）
	notFound := func() {
		utils.Fail(c, 404, "资源不存在")
	}

	// filepath 形如 "/"、"/assets/x.js"；去除前导斜杠后 Clean。
	// 防目录穿越：Clean 后以 .. 开头或为绝对路径视为非法（替代 r.Static 的内置防护）。
	req := filepath.Clean(strings.TrimPrefix(c.Param("filepath"), "/"))
	if req == ".." || strings.HasPrefix(req, ".."+string(filepath.Separator)) || filepath.IsAbs(req) {
		notFound()
		return
	}

	index := filepath.Join("dist-mobile", "index.html")

	// 根入口 /m/（或 /m）→ index.html
	if req == "" || req == "." {
		c.File(index)
		return
	}

	fsPath := filepath.Join("dist-mobile", req)

	// 目标为目录 → index.html（如 /m/ 带尾斜杠的等价形式）
	if dirExists(fsPath) {
		c.File(filepath.Join(fsPath, "index.html"))
		return
	}

	// 真实文件存在（静态资源）→ 原样返回
	if info, err := os.Stat(fsPath); err == nil && !info.IsDir() {
		c.File(fsPath)
		return
	}

	// 其余（前端 history 子路由）→ 回退 index.html
	c.File(index)
}
