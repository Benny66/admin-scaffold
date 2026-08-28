# Design: guard-read-module-name

## Context

`guard_test.go` 的 `Test_ControllerLayerMustNotTouchGORM` 用字面量 `"base-backend/database"` 识别 database 包。模块名存在两处（`go.mod` + guard 字面量），复制基座改名后二者失配，护栏失效。

## Goals / Non-Goals

**Goals:**
- 模块名单一真相化：guard 从 `go.mod` 读，不再硬编码 `base-backend`。
- 行为等价：`base-backend` 场景下判断结果不变。
- 零新依赖：沿用 guard「标准库 + 既有 yaml.v3」约定。

**Non-Goals:**
- 不写 init.sh（改名脚本），那是后续 change。
- 不改 `gen-module.sh`、`_example/` 的 import（那些改动属于 init 脚本职责，本 change 不越界）。

## Decisions

### D1：新增 `readModuleName()`，从 go.mod 解析 module 名

```go
// readModuleName 从 backend/go.mod 首行 `module <name>` 读取模块名。
func readModuleName(t *testing.T) string {
    data, err := os.ReadFile(filepath.Join(backendRoot(), "go.mod"))
    if err != nil { t.Fatalf(...) }
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "module ") {
            return strings.TrimSpace(strings.TrimPrefix(line, "module "))
        }
    }
    t.Fatal("go.mod 缺少 module 声明")
    return ""
}
```

**Why：** `go.mod` 是 Go 模块名的权威来源，`module` 声明必在文件内。解析逻辑极简，无需第三方依赖。

### D2：判断改为 `imp == moduleName+"/database"`

```go
moduleName := readModuleName(t)
...
if imp == moduleName+"/database" {
    t.Errorf("%s: controllers 层不得 import %s/database（数据库访问只允许在 service 层）", path, moduleName)
}
```

**Why：** 语义不变（database 包的判断），只是把「模块前缀」从字面量换成运行时读取。错误信息用动态名，改名后提示依然准确。

### D3：保留对 `gorm.io/gorm` 的判断不动

controller 直连 GORM 有两类：`import gorm.io/gorm` 与 `import <module>/database`。前者是第三方库路径，永远叫 `gorm.io/gorm`，无需动态化；后者是本项目路径，需动态化。

**Why：** `gorm.io/gorm` 是外部依赖的固定路径，不随项目改名，保留字面量正确。

## Risks / Trade-offs

- [go.mod 解析脆弱性] → `module` 声明是 go.mod 固定结构，解析「首行 `module <name>`」稳定可靠；若未来 go.mod 格式变化（如 module 声明前出现注释/空白），`TrimSpace` + `HasPrefix` 已覆盖前导空白，注释行不会以 `module ` 开头，安全。
- [guard 改动引入回归] → guard 自身是测试，改完 `make test` 即自证；`base-backend` 场景下 `readModuleName()` 返回原值，判断结果与改前逐字节一致，无回归风险。

## Open Questions

（无。本 change 小而独立，无未决项。）
