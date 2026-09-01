# Tasks: frontend-store-guard

## 1. 实现 store 成员解析

- [x] 1.1 新建 `backend/internal/guard/frontend_store_test.go`，package guard，复用既有 `projectRoot()`
- [x] 1.2 实现成员解析：三组正则分别提取 actions（`^ {4}(?:async )?(\w+)\(`）、getters（`^ {4}(\w+): \(state\)`）、state（`state: () => ({ ... })` 块内的 `^ {4}(\w+):`）
- [x] 1.3 解析结果为空时 `t.Fatalf`（D4：护栏必须能感知自己瞎了，不能静默通过）

## 2. 实现两类引用点断言

- [x] 2.1 G4a：遍历 `<端>/src` 下所有 `.vue` / `.js`，扫 `appStore\.(\w+)`，断言每个成员在集合内
- [x] 2.2 G4b：扫 store 自身的 `this\.(\w+)\(`，断言每个成员在集合内
- [x] 2.3 豁免 `$` 前缀成员（Pinia 内置 API，见 D3）
- [x] 2.4 失败信息 MUST 包含文件路径、成员名、端名，并列出可用成员（去重排序）

## 3. 两端同等覆盖

- [x] 3.1 对 `frontend/src/stores/app.js` 与 `mobile/src/stores/app.js` 各执行一组断言

## 4. 验证护栏真的会红

- [x] 4.1 临时在 `frontend/src/views/Login.vue` 写 `appStore.loginBgUrl`（不存在的成员），跑 `go test ./internal/guard/`，确认报红并指明文件与成员，然后回滚
      > 实测报红：「前端 store 成员引用不可解析：frontend/src/views/Login.vue 中的 appStore.loginBgUrl
      > 在 frontend/src/stores/app.js 中不存在……可用成员：……」
- [x] 4.2 临时删掉 desktop store 的 `setLoginBg` action（复现本次缺陷），确认 G4b 报红，然后回滚
      > 实测报红：「前端 store 调用了未定义的成员：frontend/src/stores/app.js 中的
      > this.setLoginBg(...) 没有对应的 action 定义……」，并列出可用成员。
- [x] 4.3 回滚后跑 `make test` 确认全绿，`make lint` 无新增告警
