// Package guard 提供「架构护栏」测试：把 AGENTS.md 中的分层铁律、响应协议、
// 模型完整性等硬性约束编译成可失败的静态检查。这些测试运行在 CI 中，AI 违反
// 约束时构建会变红，而非等到人工 code review 才发现。
//
// 实现方式：通过 go/parser + go/ast 解析 backend 下各层源码，做结构化断言，
// 不引入任何第三方依赖。
package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// backendRoot 返回 backend 根目录（guard 位于 internal/guard/ 下，向上两级）。
// go test 的工作目录是本包目录 backend/internal/guard/，故 "../.." 即 backend 根。
func backendRoot() string {
	return filepath.Join("..", "..")
}

// parseGoFiles 解析目录下所有 .go 文件（含 _test.go），返回 文件名 -> AST。
func parseGoFiles(t *testing.T, dir string) map[string]*ast.File {
	t.Helper()

	files := make(map[string]*ast.File)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("无法读取目录 %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("解析 %s 失败: %v", path, err)
		}
		files[path] = f
	}
	return files
}

// fileImports 返回一个文件 AST 中所有 import 的包路径。
func fileImports(f *ast.File) []string {
	var imps []string
	for _, imp := range f.Imports {
		// import path 的值带引号，去掉即可
		imps = append(imps, strings.Trim(imp.Path.Value, `"`))
	}
	return imps
}

// containsJSONCall 判断 AST 中是否存在对 `c.JSON(` 的直接调用。
// 用于禁止 controller 手写响应。
func containsJSONCall(f *ast.File) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "JSON" {
			found = true
			return false
		}
		return true
	})
	return found
}

// hasEmbeddedBaseModel 判断结构体是否内嵌 BaseModel（即「需要建表的模型」）。
func hasEmbeddedBaseModel(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		// 匿名字段（Names == nil）且类型为 BaseModel
		if len(field.Names) == 0 {
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "BaseModel" {
				return true
			}
		}
	}
	return false
}

// modelStructs 解析 models 包，返回：
//   - withBase: 内嵌 BaseModel 的结构体名（需要建表的模型）
//   - all:      所有顶层结构体名（含关联表）
func modelStructs(t *testing.T) (withBase map[string]bool, all map[string]bool) {
	t.Helper()

	withBase = make(map[string]bool)
	all = make(map[string]bool)
	dir := filepath.Join(backendRoot(), "models")

	for _, f := range parseGoFiles(t, dir) {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name
				all[name] = true
				if hasEmbeddedBaseModel(st) {
					withBase[name] = true
				}
			}
		}
	}
	return withBase, all
}

// autoMigrateModels 解析 database.go，提取 AutoMigrate 调用中传入的模型名。
// AutoMigrate 的参数形如 `&models.User{}`，即 &CompositeLit{X: SelectorExpr{models.User}}。
func autoMigrateModels(t *testing.T) map[string]bool {
	t.Helper()

	result := make(map[string]bool)
	path := filepath.Join(backendRoot(), "database", "database.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("解析 %s 失败: %v", path, err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "AutoMigrate" {
			return true
		}
		for _, arg := range call.Args {
			// 形如 &models.User{}
			unary, ok := arg.(*ast.UnaryExpr)
			if !ok || unary.Op != token.AND {
				continue
			}
			comp, ok := unary.X.(*ast.CompositeLit)
			if !ok {
				continue
			}
			sel, ok := comp.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			result[sel.Sel.Name] = true
		}
		return true
	})
	return result
}

// Test_ServiceLayerMustNotImportGin 分层铁律：services/ 不得触碰 gin。
func Test_ServiceLayerMustNotImportGin(t *testing.T) {
	dir := filepath.Join(backendRoot(), "services")
	for path, f := range parseGoFiles(t, dir) {
		for _, imp := range fileImports(f) {
			if strings.Contains(imp, "gin-gonic/gin") {
				t.Errorf("%s: services 层不得 import %q（服务层不触碰 HTTP 语义，见 AGENTS.md 第 1 条）", path, imp)
			}
		}
	}
}

// Test_ControllerLayerMustNotTouchGORM 分层铁律：controllers/ 不得直接操作 GORM。
func Test_ControllerLayerMustNotTouchGORM(t *testing.T) {
	dir := filepath.Join(backendRoot(), "controllers")
	for path, f := range parseGoFiles(t, dir) {
		for _, imp := range fileImports(f) {
			if strings.Contains(imp, "gorm.io/gorm") {
				t.Errorf("%s: controllers 层不得 import %q（控制器不直接操作 GORM）", path, imp)
			}
			if imp == "base-backend/database" {
				t.Errorf("%s: controllers 层不得 import base-backend/database（数据库访问只允许在 service 层）", path)
			}
		}
	}
}

// Test_ControllersMustNotWriteJSONDirectly 响应协议：controllers/ 禁止手写 c.JSON。
func Test_ControllersMustNotWriteJSONDirectly(t *testing.T) {
	dir := filepath.Join(backendRoot(), "controllers")
	for path, f := range parseGoFiles(t, dir) {
		if containsJSONCall(f) {
			t.Errorf("%s: controllers 层禁止手写 c.JSON，必须使用 utils 的 Success/Fail/SuccessPage 等方法（见 AGENTS.md 第 2 条）", path)
		}
	}
}

// Test_AllModelsRegisteredInAutoMigrate 模型完整性：models/ 中每个内嵌 BaseModel 的
// 结构体都必须出现在 database.go 的 AutoMigrate 列表；反向亦然（防 typo）。
func Test_AllModelsRegisteredInAutoMigrate(t *testing.T) {
	withBase, all := modelStructs(t)
	migrated := autoMigrateModels(t)

	// 正向：需要建表的模型（内嵌 BaseModel）必须注册进 AutoMigrate
	for name := range withBase {
		if !migrated[name] {
			t.Errorf("模型 %s 内嵌 BaseModel 但未出现在 database.go 的 AutoMigrate 列表，请注册（见 AGENTS.md 第 7 条）", name)
		}
	}

	// 反向：AutoMigrate 引用的结构体必须真实存在于 models 包（防 typo / 漏删）
	for name := range migrated {
		if !all[name] {
			t.Errorf("AutoMigrate 引用了 models 包中不存在的结构体 %s，请检查是否拼写错误", name)
		}
	}
}

// Test_GoldenPathExampleIntact 黄金路径一致性：_example/ 范例目录必须保持完整，
// 且其模板文件的 package 声明与真实三层包一致（范例是 gen-module.sh 的模板源，
// 腐化会导致生成器产出错误骨架）。
func Test_GoldenPathExampleIntact(t *testing.T) {
	root := backendRoot()
	exampleDir := filepath.Join(root, "_example")

	// 1) 三个关键模板文件必须存在
	required := map[string]string{
		filepath.Join("models", "example.go"):           "models",
		filepath.Join("services", "example_service.go"): "services",
		filepath.Join("controllers", "example.go"):      "controllers",
	}
	for rel, wantPkg := range required {
		path := filepath.Join(exampleDir, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("黄金路径范例缺失 %s（gen-module.sh 依赖该模板）: %v", path, err)
			continue
		}
		// 2) package 声明必须与真实层包名一致
		if !strings.Contains(string(data), "package "+wantPkg) {
			t.Errorf("%s: 范例 package 应为 %q（与真实分层包一致），否则生成器产出错误骨架", path, wantPkg)
		}
	}
}

// Test_RouterAndMigrateGenAnchors 生成器锚点存在性：router.go 与 database.go 中
// 的 gen 锚点注释必须保留，否则 gen-module.sh 无法注入。
func Test_RouterAndMigrateGenAnchors(t *testing.T) {
	routerPath := filepath.Join(backendRoot(), "router", "router.go")
	if data, err := os.ReadFile(routerPath); err != nil || !strings.Contains(string(data), "【gen:routes】") {
		t.Errorf("router.go 缺少 【gen:routes】 锚点（gen-module.sh 依赖它注入路由）")
	}

	dbPath := filepath.Join(backendRoot(), "database", "database.go")
	if data, err := os.ReadFile(dbPath); err != nil || !strings.Contains(string(data), "【gen:migrate】") {
		t.Errorf("database.go 缺少 【gen:migrate】 锚点（gen-module.sh 依赖它注入 AutoMigrate）")
	}
}
