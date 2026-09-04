// Package guard 品牌色提取护栏：把「桌面端与移动端的 colorExtract 模块必须提供
// 完全相同的导出接口」编译成会失败的静态检查。
//
// 背景：brand-color-extract spec 要求 frontend 与 mobile 共享同一份提取逻辑
// （同接口、同算法）。两端是两份物理文件，改了一端忘改另一端不会有任何报错，
// 只表现为「桌面端跟随 logo 变色、移动端仍是蓝色」这类局部失效。
//
// 实现：不引入 JS 解析器，用 regexp 解析本项目稳定的 ES module 具名导出写法
// （`export function` / `export async function` / `export const`）。
package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var (
	exportFuncRe  = regexp.MustCompile(`(?m)^export\s+(?:async\s+)?function\s+([A-Za-z_]\w*)`)
	exportConstRe = regexp.MustCompile(`(?m)^export\s+const\s+([A-Za-z_]\w*)`)
)

// jsExportedNames 解析一个 ES module 源文件的具名导出，返回排序后的名字列表。
func jsExportedNames(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	src := string(data)

	matches := append(exportFuncRe.FindAllStringSubmatch(src, -1), exportConstRe.FindAllStringSubmatch(src, -1)...)
	seen := make(map[string]bool, len(matches))
	for _, m := range matches {
		seen[m[1]] = true
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// exportHas 判断导出列表中是否包含 name。
func exportHas(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// Test_ColorExtractExportsMatchAcrossEnds 桌面端与移动端 colorExtract 的导出接口必须一致。
func Test_ColorExtractExportsMatchAcrossEnds(t *testing.T) {
	ends := []struct {
		name string
		path string
	}{
		{"桌面端", filepath.Join(projectRoot(), "frontend", "src", "utils", "colorExtract.js")},
		{"移动端", filepath.Join(projectRoot(), "mobile", "src", "utils", "colorExtract.js")},
	}

	// spec 点名的核心能力，缺一个即视为该端未接入
	required := []string{"extractDominantColor", "deriveThemeVars", "applyThemeVars", "DEFAULT_PRIMARY", "DEFAULT_BRAND_TO"}

	exports := make([][]string, len(ends))
	for i, end := range ends {
		if _, err := os.Stat(end.path); err != nil {
			t.Fatalf("%s缺少 %s（brand-color-extract spec 要求两端共享同一份提取逻辑）: %v", end.name, end.path, err)
		}

		names := jsExportedNames(t, end.path)
		// 解析不到任何导出说明写法已变、正则失效。此时必须 Fatal：
		// 若当作「无导出」继续，缺失断言会全部误报；若当作「通过」，护栏就瞎了。
		if len(names) == 0 {
			t.Fatalf("未能从%s的 colorExtract.js 解析出任何导出——若写法已变更，"+
				"请同步更新 backend/internal/guard/color_extract_test.go 的解析规则", end.name)
		}
		for _, want := range required {
			if !exportHas(names, want) {
				t.Errorf("%s的 colorExtract.js 缺少导出 %s（brand-color-extract spec 点名要求）", end.name, want)
			}
		}
		exports[i] = names
	}

	// 4) 两端文件必须逐字节相同：spec 要求「同接口、同算法」，只比对导出集合
	// 挡不住「改了一端的提取算法却忘了同步另一端」这类漂移。
	desktopRaw, err := os.ReadFile(ends[0].path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", ends[0].path, err)
	}
	mobileRaw, err := os.ReadFile(ends[1].path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", ends[1].path, err)
	}
	if string(desktopRaw) != string(mobileRaw) {
		t.Errorf("colorExtract 实现已漂移：%s 与 %s 内容不一致——"+
			"brand-color-extract spec 要求两端共享同一份实现，请同步修改另一端", ends[0].path, ends[1].path)
	}

	desktop, mobile := exports[0], exports[1]
	for _, name := range desktop {
		if !exportHas(mobile, name) {
			t.Errorf("colorExtract 导出不一致：桌面端有 %s，移动端缺失——两端必须提供相同接口。"+
				"桌面端导出：%s；移动端导出：%s", name, strings.Join(desktop, ", "), strings.Join(mobile, ", "))
		}
	}
	for _, name := range mobile {
		if !exportHas(desktop, name) {
			t.Errorf("colorExtract 导出不一致：移动端有 %s，桌面端缺失——两端必须提供相同接口。"+
				"桌面端导出：%s；移动端导出：%s", name, strings.Join(desktop, ", "), strings.Join(mobile, ", "))
		}
	}
}
