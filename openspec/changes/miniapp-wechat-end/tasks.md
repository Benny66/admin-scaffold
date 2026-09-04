## 1. 前置检查与协调

- [x] 1.1 检查 `auth-token-lifecycle` 状态：若已归档可继续；若仍在 in-progress，本 change 的 miniapp/src/utils/request.js 只做基础封装（Authorization + 401 跳登录），不接 refresh
- [x] 1.2 确认 `login-brand-visual` 是否已归档（它是 `auth-token-lifecycle` 的前置）
- [x] 1.3 在新分支 `feature/miniapp-wechat-end` 上开工

## 2. 后端：User 模型 + 配置段

- [x] 2.1 `backend/models/user.go` 新增 `OpenID string` 字段（gorm tag：`uniqueIndex`，可空）
- [x] 2.2 `backend/config/config.go` 在 `AppConfig` 中新增 `Wechat { AppID, Secret string }` 段；默认值空串
- [x] 2.3 `backend/config.example.yaml` 新增 `wechat:` 段示例（appid/secret 留占位 + 注释说明"由具体项目填入"）
- [x] 2.4 `backend/database/migrations.go` 或对应 AutoMigrate 调用点确认：User 表已通过 GORM AutoMigrate 加列，无需额外迁移脚本
- [x] 2.5 跑 `make test` 确认后端编译通过、guard 测试通过（含 dependency-registry 校验此时还未涉及 miniapp 段）

## 3. 后端：jscode2session 调用封装

- [x] 3.1 新建 `backend/utils/wechat.go`，定义 `JsCode2Session(appid, secret, code string) (openid, sessionKey string, err error)`
- [x] 3.2 用 `net/http.Get` + `encoding/json` 调 `https://api.weixin.qq.com/sns/jscode2session?...`
- [x] 3.3 解析响应：`errcode != 0` 时返回包含 errcode + errmsg 的错误
- [x] 3.4 写 `backend/utils/wechat_test.go` 单元测试，mock HTTP 响应覆盖成功 / errcode / 网络错误三种
- [x] 3.5 跑 `make test` 确认通过

## 4. 后端：mp-login service + controller + router

- [x] 4.1 `backend/services/auth_service.go`（或新建 `backend/services/mp_auth_service.go`）新增 `MpLogin(code string) (*LoginResult, error)`：
  - 取 `config.Wechat.AppID/Secret`，空则返回明确错误
  - 调 `JsCode2Session` 拿 openid
  - 按 openid 查 User：未找到则创建新 User（username = `mp_<openid 前 8 位>`，随机密码，`OpenID = openid`，绑定默认角色）
  - 找到则直接走原 login 签发路径
  - 签发与 `/api/auth/login` 完全相同的 JWT（复用现有 `GenerateToken` 工具）
  - 返回 `{ token, user, permissions }` 与原 Login 结构一致
- [x] 4.2 `backend/controllers/auth.go`（或新建 `backend/controllers/mp_auth.go`）新增 `MpLogin(c *gin.Context)` handler：参数校验（code 非空）→ 调 service → `utils.Success` 或 `utils.Fail`
- [x] 4.3 `backend/router/router.go` 在公开路由组（与 `/api/auth/login` 同组）注册 `POST /api/auth/mp-login`
- [x] 4.4 跑 `make test` 确认编译 + 既有用例不被破坏
- [x] 4.5 手测：未配置 wechat 段时调 mp-login 返回 500 + 明确指引；冒烟用例见 §7

## 5. 后端：smoke 测试扩展

- [x] 5.1 `backend/scripts/smoke.sh` 在现有 admin 登录用例之后，新增 `curl -sf -X POST $BASE/api/auth/mp-login -d '{"code":"smoke-test"}'` 一行
- [x] 5.2 断言响应 `code === 500`（未配置 wechat 段时，与 wechat-mp-login spec 一致）+ 包含"未配置"指引，证明接口已注册可达
- [x] 5.3 跑 `make smoke` 确认全绿

## 6. miniapp 端骨架（前端侧）

- [x] 6.1 创建 `miniapp/` 目录，初始化 uniapp vue3 + vite + pinia + uni-mp-weixin 项目（参考官方 cli 模板）
- [x] 6.2 `miniapp/package.json`：name=`base-backend-miniapp`；scripts 含 `dev:mp-weixin`、`build:mp-weixin`、`lint`（透传根 `eslint.config.js`）；dependencies 锁定 `@dcloudio/uni-app`、`@dcloudio/uni-mp-weixin`、`@dcloudio/vite-plugin-uni`、`vue`、`pinia`
- [x] 6.3 `miniapp/vite.config.js`：alias `@ → src`；plugins 包含 `@dcloudio/vite-plugin-uni`
- [x] 6.4 `miniapp/src/manifest.json`：`name=base-backend-miniapp`、`mp-weixin.appid` 占位（如 `touristappid`）、`mp-weixin.setting.esBuild=true`
- [x] 6.5 `miniapp/src/pages.json`：注册 `pages/login/index`（首项，作为入口）、`pages/index/index`
- [x] 6.6 `miniapp/src/App.vue` + `main.js`：初始化 pinia、挂载 App
- [x] 6.7 在微信开发者工具中打开 `miniapp/dist/dev/mp-weixin/` 确认空骨架能编译

## 7. miniapp 端：utils/request.js 封装层

- [x] 7.1 创建 `miniapp/src/utils/request.js`，封装 `uni.request`：
  - 导出 `request.get/post/put/delete` 或单一 `request(options)` 函数
  - 自动从 storage 读取 token 设置 `Authorization: Bearer <token>`
  - 响应拦截：`code === 200` 返回 `res.data`；401 清 token 跳 `pages/login/index`；403 提示无权限；其他错误 toast 提示
  - 基础 URL 从 `import.meta.env` 或 manifest 配置读取（dev 用占位，生产填真实域名）
- [x] 7.2 跑 `cd miniapp && npm run lint` 确认封装层自身不触发 uni.request 禁令（豁免生效）

## 8. miniapp 端：stores + 最小范例页面

- [x] 8.1 创建 `miniapp/src/stores/app.js`：state `{ systemName, logo, token, userInfo, permissions, isAdmin }` + actions `fetchSystemInfo`、`setToken`、`logout`
- [x] 8.2 创建 `miniapp/src/api/index.js`：导出 `mpLogin(code)`、`getSystemInfo()`，全部走 `@/utils/request`
- [x] 8.3 创建 `miniapp/src/pages/login/index.vue`：onLoad 调 `uni.login` 拿 code → 调 `mpLogin(code)` → 成功后存 token 跳 `pages/index/index`
- [x] 8.4 创建 `miniapp/src/pages/index/index.vue`：onShow 调 `fetchSystemInfo` → navbar 显示 `systemName`（无 logo 回退文字）
- [ ] 8.5 在微信开发者工具中手测：登录成功跳首页、首页显示从后端拉取的 systemName

## 9. eslint 护栏扩展

- [x] 9.1 `eslint.config.js` flat config：保留全局 axios 禁令 + `@/store/` 单数禁令不变
- [x] 9.2 新增 `files: ['miniapp/src/**']` 段：关闭 axios 禁令，保留 `@/store/` 单数禁令
- [x] 9.3 新增 `files: ['miniapp/src/**', '!miniapp/src/utils/request.js']` 段：开启 `no-restricted-syntax` 抓 `CallExpression[callee.object.name='uni'][callee.property.name='request']`，message 提示走 `@/utils/request`
- [x] 9.4 新增豁免 `files: ['miniapp/src/utils/request.js']`：`no-restricted-syntax: 'off'`
- [x] 9.5 跑 `make lint` 确认三端（frontend / mobile / miniapp）都跑通

## 10. deps.yaml + Makefile + .gitignore

- [x] 10.1 `deps.yaml` 新增 `miniapp:` 段，登记 `@dcloudio/uni-app` / `@dcloudio/uni-mp-weixin` / `@dcloudio/vite-plugin-uni` / `vue` / `pinia`，每条附 reason
- [x] 10.2 确认 guard 测试 dependency-registry 段扩展到 miniapp（更新 `internal/guard/dependency_registry_test.go` 把 miniapp/package.json 纳入校验）
- [x] 10.3 `Makefile` 新增 `dev-mp: cd miniapp && npm run dev:mp-weixin`
- [x] 10.4 `Makefile` 新增 `build-mp: cd miniapp && npm run build:mp-weixin`
- [x] 10.5 `Makefile` 扩展 `lint` target，新增 `cd miniapp && npm run lint`
- [x] 10.6 `.gitignore` 新增 `miniapp/dist/` 与 `miniapp/unpackage/`
- [x] 10.7 跑 `make test` + `make lint` 确认 guard 校验全绿

## 11. guard 测试扩展（frontend-store-guard + brand-config-guard）

- [x] 11.1 扩展 `internal/guard/frontend_store_guard_test.go`：扫描范围增加 `miniapp/src/`，断言 `miniapp/src/stores/app.js` 的 `appStore.<成员>` 都可解析
- [x] 11.2 扩展 `internal/guard/brand_config_guard_test.go`：消费点扩展到 `miniapp/src/stores/app.js` 的 `fetchSystemInfo`，`favicon` 字段对 miniapp 端"忽略不报错"
- [x] 11.3 跑 `make test` 确认扩展后的 guard 全绿

## 12. scripts/init.sh 扩展

- [x] 12.1 `scripts/init.sh` 新增改名步骤：改 `miniapp/package.json` 的 `name` 为 `<新项目名>-miniapp`
- [x] 12.2 `scripts/init.sh` 新增改名步骤：改 `miniapp/src/manifest.json` 的 `name` 字段为 `<新项目名>`，`mp-weixin.appid` 不动
- [x] 12.3 在临时目录复制基座跑 `make init name=test-project`，确认 miniapp 包名与 manifest 都被正确改写
- [x] 12.4 验证 `make init` 后 `cd miniapp && npm install && npm run dev:mp-weixin` 仍可启动

## 13. 文档叙事：三端 → 多端

- [x] 13.1 `README.md` 标题与目录结构图：从「三端脚手架」改为「多端脚手架」；目录结构补 `miniapp/` 行及定位说明
- [x] 13.2 `README.md` 启动说明新增「第四步（可选）：启动 miniapp」一节
- [x] 13.3 `README.md` 技术栈表新增 miniapp 行（uniapp + vue3 + pinia + uni-mp-weixin）
- [x] 13.4 `AGENTS.md §1` 标题「三端统一」改「多端统一」，正文 backend/frontend/mobile/miniapp 四端措辞统一
- [x] 13.5 `docs/目录结构约定.md` 补 miniapp 一节，说明 `pages/` 约定与三端铁律的边界
- [x] 13.6 `docs/鉴权与权限.md` 补 miniapp 登录流程：`wx.login → code → mp-login → JWT`，与现有 username/password 流程并列
- [x] 13.7 在两份文档中确认"新增业务模块参考 `backend/_example/`"的措辞与 `frontend-rbac` spec 自洽（不引入矛盾）

## 14. 整体验收

- [x] 14.1 `make test` 全绿（含扩展后的 guard 测试：dependency-registry / frontend-store-guard / brand-config-guard / frontend-rbac）
- [x] 14.2 `make lint` 全绿（backend vet + 三端 ESLint）
- [x] 14.3 `make smoke` 全绿（含 mp-login 可达性冒烟）
- [ ] 14.4 手测：在微信开发者工具中打开 miniapp，完成登录 → 跳首页 → 显示 systemName 完整链路
- [x] 14.5 手测：跑 `make init name=test-project` 在临时目录，确认 miniapp 包名 + manifest 被正确改写、`mp-weixin.appid` 保留
- [ ] 14.6 PR 评审通过，准备 archive change

## 15. 归档

- [ ] 15.1 提交 PR、合并到 main
- [ ] 15.2 跑 `openspec archive miniapp-wechat-end` 把 spec 并入 `openspec/specs/`
- [ ] 15.3 在 follow-up change（待 auth-token-lifecycle 归档后开）规划 miniapp 接入 refresh 链路
