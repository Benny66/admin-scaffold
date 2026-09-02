// Package guard 权限码注册护栏：把「router.go 用到的权限码必须在 initBaseData 注册过」
// 编译成会失败的静态检查。
//
// 背景：make gen 会自动往 router.go 注入 `PermissionRequired("<资源>:<动作>")`，但
// initBaseData 里的 Permission 记录若漏注册，会产生一条全程静默的失败链——
// guard（只校验「前端码 ⊆ router.go」）、smoke（用 admin 登录，PermissionRequired
// 对 isAdmin 直通放行）、CI 全部变绿，只有真实非管理员用户会撞上 403；而且权限管理
// 界面的数据源是 permissions 表，表里没有这些码时管理员在界面上也无法授权自救。
//
// 本护栏是既有「模型必须注册进 AutoMigrate」（防漏建表）的同构对偶：
//   AutoMigrate 漏注册 → 运行时表不存在
//   权限码漏注册       → 非管理员用户静默 403
//
// 实现：沿用 guard 包既有手法（标准库 regexp + 文件读取），不引入新依赖。
package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// permissionAnchor 权限声明块锚点。它一身二用：
//   1. gen-module.sh 注入新模块权限码的位置
//   2. 本护栏识别「已注册权限码」提取范围的起点
// 因此删除或改名会同时破坏生成与校验，两边必须一致。
const permissionAnchor = "【gen:permissions】"

// routerPermissionCodes 从 router/router.go 提取 PermissionRequired("...") 的权限码集合。
func routerPermissionCodes(t *testing.T) map[string]bool {
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

// declaredPermissionCodes 从 database/database.go 中 【gen:permissions】 锚点起始的
// 权限声明块内提取 Code: "..." 的权限码集合。
//
// 提取范围 MUST 限定在锚点块内：initBaseData 里 Role{Code: "admin"}、
// DictType{Code: "user_status"} 等其他模型同样带有 Code: 字段，无脑全文件扫描会把
// 它们误收进「已注册集合」，从而掩盖真实的权限码缺失。
func declaredPermissionCodes(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(backendRoot(), "database", "database.go"))
	if err != nil {
		t.Fatalf("读取 database.go 失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	start := -1
	for i, ln := range lines {
		if strings.Contains(ln, permissionAnchor) {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("database.go 缺少 %s 锚点——权限声明块无法定位。该锚点同时是 make gen 的注入位置，"+
			"请在 permissions 字面量块的首行补回：%s", permissionAnchor,
			filepath.Join(backendRoot(), "database", "database.go"))
	}

	// 从锚点扫到权限字面量块的闭合 `}`（独占一行，且该行不含 `{`）
	var block []string
	for _, ln := range lines[start:] {
		trimmed := strings.TrimSpace(ln)
		if strings.HasSuffix(trimmed, "}") && !strings.Contains(trimmed, "{") {
			break
		}
		block = append(block, ln)
	}

	re := regexp.MustCompile(`Code:\s*"([^"]+)"`)
	codes := make(map[string]bool)
	for _, ln := range block {
		// 必须先剔除行注释：否则被 // 注释掉的权限码仍会被当作已注册，
		// 护栏产生假阴性——而「注释掉一行」恰恰是开发者最常见的临时操作。
		if idx := strings.Index(ln, "//"); idx >= 0 {
			ln = ln[:idx]
		}
		for _, m := range re.FindAllStringSubmatch(ln, -1) {
			codes[m[1]] = true
		}
	}
	return codes
}

// Test_PermissionCodesRegisteredInBaseData：router.go 用到的权限码必须已在 initBaseData 注册。
// 只做单向 router ⊆ initBaseData，不做反向（initBaseData 允许存在尚未接线的权限码）。
func Test_PermissionCodesRegisteredInBaseData(t *testing.T) {
	used := routerPermissionCodes(t)
	declared := declaredPermissionCodes(t)

	// 沿用 frontend_rbac_test.go 的 G5c 原则：解析结果为空必须 Fatal，
	// 否则正则或锚点失效时护栏会「瞎掉」并静默放行。
	if len(used) == 0 {
		t.Fatalf("未能从 router.go 解析出任何 PermissionRequired 权限码——写法可能已变更，" +
			"请同步更新 backend/internal/guard/permission_registry_test.go 的解析规则")
	}
	if len(declared) == 0 {
		t.Fatalf("未能从 database.go 的 %s 块解析出任何权限码——锚点位置或 Code: 写法可能已变更，"+
			"请同步更新 backend/internal/guard/permission_registry_test.go 的解析规则",
			permissionAnchor)
	}

	var missing []string
	for code := range used {
		if !declared[code] {
			missing = append(missing, code)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		var declaredList []string
		for code := range declared {
			declaredList = append(declaredList, code)
		}
		sort.Strings(declaredList)

		t.Errorf("以下权限码在 router.go 的 PermissionRequired 中被使用，但未在 database.go 的 "+
			"initBaseData 权限声明块（%s 锚点内）注册：\n  %v\n"+
			"后果：非管理员用户请求这些接口会 403，且权限管理界面看不到这些码，管理员无法在界面上授权。\n"+
			"修法：用 make gen 生成模块（会自动注入），或在 initBaseData 手工补注册。\n"+
			"已注册码：%s",
			permissionAnchor, missing, strings.Join(declaredList, ", "))
	}
}

// Test_ExampleTemplateMatchesPluralizeRule：范例模板展示的路由路径必须与 pluralize.sh
// 的规则产出一致，防止「注释写一套、生成器产另一套」成为新的漂移源。
//
// pluralize.sh 是复数规则的单一真相，gen-module.sh 与护栏共同调用它；而 _example/
// 是 AI 生成模块时模仿的标准答案，两者一旦不同步，AI 会照着过时的注释写出不一致的代码。
func Test_ExampleTemplateMatchesPluralizeRule(t *testing.T) {
	script := filepath.Join(backendRoot(), "scripts", "pluralize.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("复数化脚本缺失：%s——它是 gen-module.sh 与本护栏的共同单一真相，不得删除", script)
	}

	out, err := exec.Command("bash", script, "example").Output()
	if err != nil {
		t.Fatalf("调用 pluralize.sh example 失败: %v", err)
	}
	want := strings.TrimSpace(string(out))
	if want == "" {
		t.Fatalf("pluralize.sh 对 example 返回空结果")
	}

	tmpl := filepath.Join(backendRoot(), "_example", "router", "example.go")
	data, err := os.ReadFile(tmpl)
	if err != nil {
		t.Fatalf("读取 _example/router/example.go 失败: %v", err)
	}

	re := regexp.MustCompile(`Group\("/([a-z0-9_]+)"\)`)
	m := re.FindStringSubmatch(string(data))
	if m == nil {
		t.Fatalf("未能从 _example/router/example.go 解析出示例路由路径（形如 Group(\"/examples\")）——" +
			"模板写法可能已变更，请同步更新本护栏的解析规则")
	}

	if got := m[1]; got != want {
		t.Errorf("_example/router/example.go 注释中的路由路径 /%s 与 pluralize.sh 的规则产出 %q 不一致。"+
			"范例模板是 AI 生成模块的标准答案，与生成器实际产出不同步就会成为新的漂移源。"+
			"请将模板中的路径改为 /%s", got, want, want)
	}
}
