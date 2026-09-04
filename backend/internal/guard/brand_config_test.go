// Package guard 品牌配置护栏：把「品牌配置四处副本必须同步」与「后端返回字段必须
// 被两端 store 消费」两条软约束编译成会失败的静态检查。
//
// 背景：backend/config/config.go 里品牌配置的每个字段都有四份副本——AppConfig 结构体、
// init() 默认值、yamlFile 影子结构、以及逐字段手写的 if 非空覆盖链。漏改任何一处都
// 不会报错，只会静默产生「配了不生效」或「永远取默认值」。此外 GetSystemInfo 新增
// 返回字段而前端 store 未解构时，该字段会静默变成死配置（subtitle 就曾如此）。
//
// 实现说明：yamlFile 是 config 包的非导出类型，guard 包无法用 reflect 直接取它的
// 字段，故 G1 与 G2、G3 一样统一采用 go/ast 解析源码，不引入任何第三方依赖。
package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// configGoPath 返回品牌配置四处副本所在的文件路径。
func configGoPath() string {
	return filepath.Join(backendRoot(), "config", "config.go")
}

// systemCtrlPath 返回 GetSystemInfo 所在的文件路径。
func systemCtrlPath() string {
	return filepath.Join(backendRoot(), "controllers", "system.go")
}

// parseFileAST 解析单个 .go 文件并返回其 AST。
func parseFileAST(t *testing.T, path string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("解析 %s 失败: %v", path, err)
	}
	return f
}

// structField 描述结构体中一个字段的 Go 名与 yaml tag 名。
type structField struct {
	goName  string
	yamlTag string
}

// findStructType 在文件 AST 中查找名为 typeName 的结构体类型。
func findStructType(t *testing.T, f *ast.File, typeName string) *ast.StructType {
	t.Helper()
	var found *ast.StructType
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name.Name != typeName {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		found = st
		return false
	})
	if found == nil {
		t.Fatalf("在 %s 中未找到结构体 %s", f.Name.Name, typeName)
	}
	return found
}

// structFields 提取结构体中的命名字段（跳过匿名/嵌入字段）。
func structFields(st *ast.StructType) []structField {
	var fields []structField
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			continue
		}
		fields = append(fields, structField{
			goName:  f.Names[0].Name,
			yamlTag: yamlTagName(f.Tag),
		})
	}
	return fields
}

// nestedStructField 在结构体中查找名为 name 的字段，并返回其匿名结构体类型。
func nestedStructField(t *testing.T, st *ast.StructType, name string) *ast.StructType {
	t.Helper()
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 || f.Names[0].Name != name {
			continue
		}
		inner, ok := f.Type.(*ast.StructType)
		if !ok {
			t.Fatalf("字段 %s 不是匿名结构体类型，无法作为嵌套配置段", name)
		}
		return inner
	}
	t.Fatalf("结构体中未找到字段 %s", name)
	return nil
}

// yamlTagName 从结构体字段 tag 中取出 yaml tag 名（取逗号前的部分）。
//
// 注意：不能用 strings.Trim(tag.Value, "`\"") 去引号——Trim 按字符集裁剪，会把 tag
// 内容里有意义的闭合引号一并裁掉（如 `yaml:"name"` 会变成 yaml:"name），导致解析失败。
// 正确做法是先用 strconv.Unquote 解开字面量（同时支持反引号与双引号两种写法），
// 再交给 reflect.StructTag 按 Go 的 tag 约定解析。
func yamlTagName(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	content, err := strconv.Unquote(tag.Value)
	if err != nil {
		return ""
	}
	return strings.Split(reflect.StructTag(content).Get("yaml"), ",")[0]
}

// tagSet 把字段列表转成 yaml tag 的集合。
func tagSet(fields []structField) map[string]struct{} {
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if f.yamlTag == "" {
			continue
		}
		set[f.yamlTag] = struct{}{}
	}
	return set
}

// Test_BrandConfigStructsInSync G1：AppConfig 与 yamlFile.App 的 yaml tag 集合必须
// 完全一致。漏改任一方向都会产生「配了不生效」或僵尸字段。
func Test_BrandConfigStructsInSync(t *testing.T) {
	f := parseFileAST(t, configGoPath())

	appFields := structFields(findStructType(t, f, "AppConfig"))
	shadowFields := structFields(nestedStructField(t, findStructType(t, f, "yamlFile"), "App"))
	appTags, shadowTags := tagSet(appFields), tagSet(shadowFields)

	for _, fld := range appFields {
		if fld.yamlTag == "" {
			continue
		}
		if _, ok := shadowTags[fld.yamlTag]; !ok {
			t.Errorf("品牌配置字段不同步：AppConfig 有字段 %s（yaml tag %q），但 %s 的 yamlFile.App 影子结构中缺失——"+
				"请在该结构体的 App 段补上同名字段", fld.goName, fld.yamlTag, configGoPath())
		}
	}
	for _, fld := range shadowFields {
		if fld.yamlTag == "" {
			continue
		}
		if _, ok := appTags[fld.yamlTag]; !ok {
			t.Errorf("品牌配置字段不同步：yamlFile.App 有僵尸字段 %s（yaml tag %q），但 %s 的 AppConfig 中不存在——"+
				"请删除该字段或在 AppConfig 中补上", fld.goName, fld.yamlTag, configGoPath())
		}
	}
}

// Test_BrandConfigOverrideChainComplete G2：每个带 yaml tag 的品牌字段都必须在覆盖链
// 中有对应的 `if yf.App.<Field> != ""` 分支，否则 config.yaml 配了也不生效。
func Test_BrandConfigOverrideChainComplete(t *testing.T) {
	f := parseFileAST(t, configGoPath())
	covered := overrideCoveredFields(f)

	for _, fld := range structFields(findStructType(t, f, "AppConfig")) {
		if fld.yamlTag == "" {
			continue
		}
		if _, ok := covered[fld.goName]; !ok {
			t.Errorf("品牌配置覆盖链不完整：字段 %s（yaml tag %q）在 %s 中缺少对应的非空覆盖分支——"+
				"请补上 if yf.App.%s != \"\" { GlobalConfig.App.%s = yf.App.%s }",
				fld.goName, fld.yamlTag, configGoPath(), fld.goName, fld.goName, fld.goName)
		}
	}
}

// overrideCoveredFields 收集 config.go 中所有形如 `if yf.App.<Field> != ""` 的字段名。
func overrideCoveredFields(f *ast.File) map[string]struct{} {
	covered := make(map[string]struct{})
	ast.Inspect(f, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok || bin.Op != token.NEQ {
			return true
		}
		// 条件左侧形如 yf.App.<Field>
		sel, ok := bin.X.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil {
			return true
		}
		app, ok := sel.X.(*ast.SelectorExpr)
		if !ok || app.Sel == nil || app.Sel.Name != "App" {
			return true
		}
		root, ok := app.X.(*ast.Ident)
		if !ok || root.Name != "yf" {
			return true
		}
		// 条件右侧必须是空串字面量
		lit, ok := bin.Y.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || strings.Trim(lit.Value, `"`) != "" {
			return true
		}
		covered[sel.Sel.Name] = struct{}{}
		return true
	})
	return covered
}

// Test_SystemInfoFieldsConsumedByBothEnds G3：GetSystemInfo 返回的每个字段都必须被
// 前端、移动端与小程序端 store 的 fetchSystemInfo 解构消费，杜绝死配置复发。
//
// miniapp 特例：favicon 字段对小程序端「忽略不报错」——小程序无浏览器标签概念，
// favicon 在 miniapp 端可不解构消费（brand-config spec 的 miniapp 消费规则）。
func Test_SystemInfoFieldsConsumedByBothEnds(t *testing.T) {
	keys := systemInfoResponseKeys(t)

	ends := []struct {
		name           string
		path           string
		ignoredFields  map[string]bool
	}{
		{"前端", filepath.Join(projectRoot(), "frontend", "src", "stores", "app.js"), nil},
		{"移动端", filepath.Join(projectRoot(), "mobile", "src", "stores", "app.js"), nil},
		{"小程序端", filepath.Join(projectRoot(), "miniapp", "src", "stores", "app.js"), map[string]bool{"favicon": true}},
	}

	for _, end := range ends {
		consumed := storeConsumedFields(t, end.path)
		for _, key := range keys {
			if end.ignoredFields != nil && end.ignoredFields[key] {
				continue
			}
			if _, ok := consumed[key]; !ok {
				t.Errorf("死配置：GetSystemInfo 返回字段 %q，但%s store（%s）的 fetchSystemInfo 未解构消费——"+
					"请在该文件的 const { ... } = res.data 中补上 %s", key, end.name, end.path, key)
			}
		}
	}
}

// systemInfoResponseKeys 解析 GetSystemInfo 函数体内 gin.H{...} 的所有 key。
func systemInfoResponseKeys(t *testing.T) []string {
	t.Helper()
	f := parseFileAST(t, systemCtrlPath())

	var fn *ast.FuncDecl
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.FuncDecl); ok && d.Name.Name == "GetSystemInfo" {
			fn = d
			break
		}
	}
	if fn == nil {
		t.Fatalf("在 %s 中未找到函数 GetSystemInfo", systemCtrlPath())
	}

	var keys []string
	ast.Inspect(fn, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil || sel.Sel.Name != "H" {
			return true
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if k, ok := kv.Key.(*ast.BasicLit); ok && k.Kind == token.STRING {
				keys = append(keys, strings.Trim(k.Value, `"`))
			}
		}
		return true
	})
	if len(keys) == 0 {
		t.Fatalf("未能从 %s 的 GetSystemInfo 中解析出 gin.H 返回字段——若写法已变更，请同步更新本护栏", systemCtrlPath())
	}
	return keys
}

// storeConsumedFields 扫描 store 文件中 `const { ... } = res.data` 的解构字段名。
func storeConsumedFields(t *testing.T, path string) map[string]struct{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	m := regexp.MustCompile(`const\s*\{([^}]*)\}\s*=\s*res\.data`).FindStringSubmatch(string(data))
	if m == nil {
		t.Fatalf("在 %s 中未找到 `const { ... } = res.data` 解构——若写法已变更，请同步更新本护栏的扫描规则", path)
	}
	consumed := make(map[string]struct{})
	for _, part := range strings.Split(m[1], ",") {
		if name := strings.TrimSpace(part); name != "" {
			consumed[name] = struct{}{}
		}
	}
	return consumed
}
