# Proposal: frontend-store-guard

## Why

实现 `login-brand-visual` 时，`frontend/src/stores/app.js` 出现了这样一个缺陷：`fetchSystemInfo`
调用了 `this.setLoginBg(...)` 与 `this.setLoginBgMobile(...)`，但这两个 action **从未在 store 中定义**。

它的破坏方式特别隐蔽：

```
setSystemName   ✓  抛错点之前的都正常执行
setSubtitle     ✓
setLogo         ✓
setFooter       ✓
setLoginBg      ✗ TypeError: this.setLoginBg is not a function
  ├─ 背景图永远不加载
  └─ favicon 注入代码永远不执行
catch (e) { /* 静默失败 */ }   ← 错误被吞掉，无任何报错
```

结果是**无报错的局部失效**。而以下四道关卡全部放行：

- `npm run build` —— Vite 不检查运行时属性是否存在
- `make lint` —— ESLint 不认识 Pinia 的 action 集合
- `make test` —— 仓库只有 Go guard 测试，前端 JS 无任何静态检查
- 接口验证 —— 后端返回字段正确，但前端根本没存下来

同类缺陷还有「模板里写 `appStore.logoAvailble`（拼错）」这类，同样静默失效。

根因是：**仓库的 guard 体系只覆盖 Go 侧（`backend/internal/guard/`），前端与移动端的 JS 层完全没有静态检查。** 本 change 补上这一类护栏。

## What Changes

- 新增 `backend/internal/guard/frontend_store_test.go`，对前端与移动端各做一组断言：
  - **G4a 模板引用可解析**：`src/` 下所有 `.vue` / `.js` 中出现的 `appStore.<成员>`，必须能在对应端 `stores/app.js` 中解析到（state 键 / getter / action 之一）。
  - **G4b store 内部调用可解析**：store 内出现的 `this.<成员>(...)`，必须能在该 store 自身解析到（action 或 getter）。
- 沿用 guard 包既有手法（标准库 `regexp` + 文件遍历），不引入 JS 解析器依赖。
- 失败信息必须指明「哪个文件的哪个成员解析不到」，并列出可用成员，便于直接修。

## Capabilities

### New Capabilities

- `frontend-store-guard`: 把「前端/移动端 store 的成员引用必须可解析」编译成会失败的 guard 测试，堵住 JS 层静默失效。

### Modified Capabilities

（无。）

## Impact

- 新增文件：`backend/internal/guard/frontend_store_test.go`、`openspec/changes/frontend-store-guard/` 下四个制品。
- 无源码改动：本 change 只加检查，不改业务代码。
- 无新第三方依赖（纯标准库）。
- 无破坏性：当前两端 store 的引用均已完整，护栏落地即为绿色。
- 不并入 `brand-config-guard`：后者 scope 是「品牌配置四处副本一致性」，与「store 成员引用完整性」无关，合并会稀释两边的能力定义。
