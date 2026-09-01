// Package guard 前端 store 成员护栏：把「所有 appStore.<X> 与 this.<X>() 引用都必须
// 能在 store 中解析」编译成会失败的静态检查。
//
// 背景：实现 login-brand-visual 时，frontend/src/stores/app.js 的 fetchSystemInfo 调用了
// 两个从未定义的 action（setLoginBg / setLoginBgMobile）。由于该函数外层是静默 try/catch，
// TypeError 被吞掉，表现为「背景图不加载 + favicon 不注入」且无任何报错；而 build、lint、
// Go 测试、接口验证四道关卡全部放行。同类缺陷还有模板里拼错成员名（appStore.logoAvailble）。
//
// 实现：不引入 JS 解析器，用 regexp 解析本项目高度稳定的 Pinia options store 写法
// （4 空格缩进）。若写法变更导致解析不到任何成员，护栏 Fatal 而非静默通过——
// 护栏必须能感知自己瞎了（brand-config-guard 第一版就栽在这上面）。
package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// jsEnd 描述一端前端应用（前端 Web 或移动端 H5）的源码根（相对于项目根）。
type jsEnd struct {
	name    string
	srcRoot string
}

// storeRef 描述一处 store 成员引用，用于失败信息定位。
type storeRef struct {
	file   string
	member string
}

// storeMemberSet 解析 Pinia store 源码，返回其对外可见的成员集合（state 键 + getters + actions）。
func storeMemberSet(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 store 文件失败 %s: %v", path, err)
	}
	src := string(data)

	members := make(map[string]bool)

	// actions：形如 `    setName(x) {` 或 `    async fetchSystemInfo() {`
	for _, m := range regexp.MustCompile(`(?m)^ {4}(?:async )?([A-Za-z_]\w*)\(`).FindAllStringSubmatch(src, -1) {
		members[m[1]] = true
	}
	// getters：形如 `    isAdmin: (state) => ...`（getters 区特有的 (state) 形参）
	for _, m := range regexp.MustCompile(`(?m)^ {4}([A-Za-z_]\w*): \(state\)`).FindAllStringSubmatch(src, -1) {
		members[m[1]] = true
	}
	// state：state: () => ({ ... }) 块内的顶层键
	if block := regexp.MustCompile(`(?s)state: \(\) => \(\{(.*?)\n {2}\}\),`).FindStringSubmatch(src); block != nil {
		for _, m := range regexp.MustCompile(`(?m)^ {4}([A-Za-z_]\w*):`).FindAllStringSubmatch(block[1], -1) {
			members[m[1]] = true
		}
	}

	// 解析不到成员说明 store 写法已变、正则失效。此时必须 Fatal：
	// 若当作「无成员」继续，后续所有引用都会误报；若当作「通过」，护栏就瞎了。
	if len(members) == 0 {
		t.Fatalf("未能从 %s 解析出任何 store 成员——store 写法可能已变更，"+
			"请同步更新 backend/internal/guard/frontend_store_test.go 的解析规则", path)
	}
	return members
}

// collectAppStoreRefs 遍历 srcRoot 下所有 .vue / .js，收集 appStore.<成员> 引用。
func collectAppStoreRefs(t *testing.T, srcRoot string) []storeRef {
	t.Helper()
	re := regexp.MustCompile(`appStore\.([A-Za-z_$]\w*)`)

	var refs []storeRef
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
			refs = append(refs, storeRef{file: path, member: m[1]})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("遍历 %s 失败: %v", srcRoot, err)
	}
	return refs
}

// collectThisCalls 收集 store 源码中 this.<成员>(...) 形式的调用。
func collectThisCalls(src string) []string {
	var out []string
	for _, m := range regexp.MustCompile(`this\.([A-Za-z_]\w*)\(`).FindAllStringSubmatch(src, -1) {
		out = append(out, m[1])
	}
	return out
}

// sortedMembers 把成员集合转成稳定排序的名字串，便于失败信息对照修正。
func sortedMembers(members map[string]bool) string {
	names := make([]string, 0, len(members))
	for n := range members {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Test_StoreMemberReferencesResolve 前端与移动端 store 的所有成员引用都必须可解析。
func Test_StoreMemberReferencesResolve(t *testing.T) {
	ends := []jsEnd{
		{"前端", filepath.Join("frontend", "src")},
		{"移动端", filepath.Join("mobile", "src")},
	}

	for _, end := range ends {
		srcRoot := filepath.Join(projectRoot(), end.srcRoot)
		storePath := filepath.Join(srcRoot, "stores", "app.js")

		data, err := os.ReadFile(storePath)
		if err != nil {
			t.Fatalf("读取%s store 失败 %s: %v", end.name, storePath, err)
		}
		members := storeMemberSet(t, storePath)

		// G4a：模板与脚本中的 appStore.<成员>
		for _, ref := range collectAppStoreRefs(t, srcRoot) {
			// Pinia 内置成员（$reset / $patch 等）由框架注入，不在 store 定义中
			if strings.HasPrefix(ref.member, "$") {
				continue
			}
			if !members[ref.member] {
				t.Errorf("%s store 成员引用不可解析：%s 中的 appStore.%s 在 %s 中不存在——"+
					"可能是拼写错误，或该成员已被重命名/删除。可用成员：%s",
					end.name, ref.file, ref.member, storePath, sortedMembers(members))
			}
		}

		// G4b：store 内部调用 this.<成员>(...)
		for _, name := range collectThisCalls(string(data)) {
			if !members[name] {
				t.Errorf("%s store 调用了未定义的成员：%s 中的 this.%s(...) 没有对应的 action 定义——"+
					"运行时会抛 TypeError；若调用方有 try/catch，错误会被静默吞掉而表现为局部失效。"+
					"可用成员：%s", end.name, storePath, name, sortedMembers(members))
			}
		}
	}
}
