// Package guard 依赖登记制护栏：把「新增依赖 MUST 登记到 deps.yaml」这条软约束
// 编译成会失败的静态检查。与既有 guard 测试同构——双向校验，正向拦「漏登记」，
// 反向拦「僵尸登记项」，防止 AI 静默引入依赖或登记了不存在的包。
//
// 校验对象只含「直接依赖」：go.mod 主 require 块（跳过 // indirect）、
// package.json 的 dependencies（跳过 devDependencies）。
package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// projectRoot 返回项目根目录（deps.yaml 与 frontend/mobile 所在），
// 即 backend 根目录的上一级。
func projectRoot() string {
	return filepath.Join(backendRoot(), "..")
}

// depEntry 是 deps.yaml 中单条依赖登记。
type depEntry struct {
	Package string `yaml:"package"`
	Reason  string `yaml:"reason"`
}

// depsFile 是 deps.yaml 的顶层结构。
type depsFile struct {
	Backend  []depEntry `yaml:"backend"`
	Frontend []depEntry `yaml:"frontend"`
	Mobile   []depEntry `yaml:"mobile"`
}

// pkgJSON 用于解析 package.json 的 dependencies 字段。
type pkgJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

// parseGoModDirectDeps 解析 go.mod，返回主 require 块中的直接依赖集合（跳过 // indirect）。
func parseGoModDirectDeps(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}

	deps := make(map[string]bool)
	inRequire := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if !inRequire {
			continue
		}
		// 跳过 indirect 依赖；只登记直接依赖
		if strings.Contains(line, "// indirect") {
			continue
		}
		// 形如 "github.com/gin-gonic/gin v1.9.1"，取首字段即模块路径
		if fields := strings.Fields(line); len(fields) >= 1 {
			deps[fields[0]] = true
		}
	}
	return deps
}

// parsePkgJSONDeps 解析 package.json，返回 dependencies 字段的键集合。
func parsePkgJSONDeps(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	var pkg pkgJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatalf("解析 %s 失败: %v", path, err)
	}
	deps := make(map[string]bool, len(pkg.Dependencies))
	for name := range pkg.Dependencies {
		deps[name] = true
	}
	return deps
}

// loadDepsRegistry 解析 deps.yaml，返回三端「已登记依赖」集合。
func loadDepsRegistry(t *testing.T) (backend, frontend, mobile map[string]bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot(), "deps.yaml"))
	if err != nil {
		t.Fatalf("读取 deps.yaml 失败: %v", err)
	}
	var reg depsFile
	if err := yaml.Unmarshal(data, &reg); err != nil {
		t.Fatalf("解析 deps.yaml 失败: %v", err)
	}

	toSet := func(entries []depEntry) map[string]bool {
		s := make(map[string]bool, len(entries))
		for _, e := range entries {
			s[e.Package] = true
		}
		return s
	}
	return toSet(reg.Backend), toSet(reg.Frontend), toSet(reg.Mobile)
}

// assertBidirectional 双向校验：清单依赖 ⊆ 登记项，且登记项 ⊆ 清单依赖。
func assertBidirectional(t *testing.T, end string, actual, registered map[string]bool) {
	t.Helper()
	// 正向：清单里的每个直接依赖都必须登记
	for pkg := range actual {
		if !registered[pkg] {
			t.Errorf("依赖 %q 未登记：请在根 deps.yaml 的 %s 段追加 {package, reason}（见 AGENTS.md 第 2 条）", pkg, end)
		}
	}
	// 反向：登记项必须真实存在于清单，防 typo / 僵尸条目
	for pkg := range registered {
		if !actual[pkg] {
			t.Errorf("deps.yaml 的 %s 段登记了 %q，但对应依赖清单中不存在（僵尸条目或拼写错误）", end, pkg)
		}
	}
}

// Test_UnregisteredDependency 依赖登记制：三端清单与 deps.yaml 双向一致。
func Test_UnregisteredDependency(t *testing.T) {
	backend, frontend, mobile := loadDepsRegistry(t)

	assertBidirectional(t, "backend",
		parseGoModDirectDeps(t, filepath.Join(backendRoot(), "go.mod")), backend)
	assertBidirectional(t, "frontend",
		parsePkgJSONDeps(t, filepath.Join(projectRoot(), "frontend", "package.json")), frontend)
	assertBidirectional(t, "mobile",
		parsePkgJSONDeps(t, filepath.Join(projectRoot(), "mobile", "package.json")), mobile)
}
