// Package guard 前端 RBAC 护栏：把「前端用到的权限码必须在后端注册过」与
// 「菜单不得硬编码」编译成会失败的静态检查。
//
// 背景：前端 RBAC 曾断在最后一米——后端 17 个权限码在 router.go 逐条接线，登录也把
// permissions 发到了浏览器，但前端的 hasPermission 从未被调用、菜单硬编码、按钮不隐藏。
// 本护栏守住修复后的闭环，防止新业务模块「前端挂了码、后端忘了注册」再次静默漂移。
//
// 实现：沿用 guard 包既有手法（标准库 regexp + 文件遍历），不引入 JS 解析器。
package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// backendPermissionCodes 从 router/router.go 提取 PermissionRequired("...") 的权限码集合。
func backendPermissionCodes(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(backendRoot(), "router", "router.go"))
	if err != nil {
		t.Fatalf("读取 router.go 失败: %v", err)
	}
	re := regexp.MustCompile(`PermissionRequired\("([^"]+)"\)`)
	codes := make(map[string]bool)
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		codes[m[1]] = true
	}
	return codes
}

// frontendPermissionCodes 从 frontend/src 下 .vue/.js 提取权限码，返回 码 -> 首次出现文件。
//
// 权限码形态为 `<资源>:<动作>`（如 users:view），以三种引号之一包裹：
//   - v-permission="'users:create'"        单个值
//   - permission: 'users:view'              meta 声明
//   - v-permission="['users:edit','users:create']"  数组（元素仍各自带引号）
//
// 正则 `['"`](word:word)['"`]` 匹配引号包裹的权限码，同时覆盖单/双/反引号；数组
// v-permission="['users:edit','users:create']" 的元素各自带引号，也能被捕获。
// 若解析结果为空，调用方会 Fatal（护栏必须能感知自己瞎了）。
func frontendPermissionCodes(t *testing.T) map[string]string {
	t.Helper()
	re := regexp.MustCompile("['\"`]([a-z_][a-z0-9_]*:[a-z_][a-z0-9_]*)[\"'`]")
	found := make(map[string]string)

	srcRoot := filepath.Join(projectRoot(), "frontend", "src")
	err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".vue" && ext != ".js" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, m := range re.FindAllStringSubmatch(string(data), -1) {
			code := m[1]
			if _, ok := found[code]; !ok {
				found[code] = path
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("遍历 frontend/src 失败: %v", err)
	}
	return found
}

// sortedCodes 把码集合转成稳定排序的名字串，便于失败信息对照修正。
func sortedCodes(codes map[string]bool) string {
	names := make([]string, 0, len(codes))
	for n := range codes {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Test_FrontendPermissionCodesRegisteredInBackend G5a：前端用到的权限码必须能在后端注册过。
// 只做单向 FE ⊆ BE，不做反向（理由见 design D6：BE 含大量接口级码，前端无对应菜单/按钮）。
func Test_FrontendPermissionCodesRegisteredInBackend(t *testing.T) {
	be := backendPermissionCodes(t)
	fe := frontendPermissionCodes(t)

	// G5c：解析结果为空必须 Fatal，而非当作「无引用」静默通过——否则正则失效时护栏会瞎掉。
	if len(be) == 0 {
		t.Fatalf("未能从 router.go 解析出任何权限码——PermissionRequired 写法可能已变更，" +
			"请同步更新 backend/internal/guard/frontend_rbac_test.go 的解析规则")
	}
	if len(fe) == 0 {
		t.Fatalf("未能从 frontend/src 解析出任何权限码——v-permission / meta.permission 写法可能已变更，" +
			"请同步更新 backend/internal/guard/frontend_rbac_test.go 的解析规则")
	}

	for code, file := range fe {
		if !be[code] {
			t.Errorf("前端权限码 %q（出现于 %s）未在后端 router.go 的 PermissionRequired 中注册——"+
				"请在后端补注册该码，或修正前端拼写。已注册码：%s", code, file, sortedCodes(be))
		}
	}
}

// Test_LayoutMustNotHardcodeMenu G5b：侧边栏菜单必须从路由声明派生，禁止硬编码路径。
func Test_LayoutMustNotHardcodeMenu(t *testing.T) {
	path := filepath.Join(projectRoot(), "frontend", "src", "layout", "Layout.vue")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 Layout.vue 失败: %v", err)
	}
	if strings.Contains(string(data), "'/system/") {
		t.Errorf("Layout.vue 中出现了 '/system/ 路径字面量——菜单必须从路由声明派生（单一数据源），" +
			"新增模块只需在 router/index.js 注册路由，禁止再手写一份菜单（见 frontend-rbac design D1/D2）")
	}
}
