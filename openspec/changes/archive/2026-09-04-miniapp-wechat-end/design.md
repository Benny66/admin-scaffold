## Context

基座是「backend + frontend（Web）+ mobile（Vant H5）」三端 monorepo。三端统一铁律（`stores/`、`@→src/`、snake_case）由 `AGENTS.md §1` 与 `eslint.config.js` 共同强制；依赖登记制（`deps.yaml`）+ guard 测试双向校验；`scripts/init.sh` 一键改名；`scripts/package.sh` 多平台打包只覆盖后端 + 前端 + 移动端产物。

现在新增第四端：uniapp 微信小程序。这是一次跨端扩展，涉及：
- 新增 `miniapp/` 目录骨架（uniapp + vue3 + pinia + uni-mp-weixin）。
- 后端新增微信登录链路（`mp-login` 接口 + `User.openid` 字段 + `wechat` 配置段）。
- 跨端铁律、依赖登记、guard 护栏、init 脚本、lint 聚合命令、文档叙事的同步扩展。

约束：
- 基座定位是「与业务无关的可复用横切能力」，miniapp 端只做骨架 + 最小范例，不做业务页。
- in-progress 的 `auth-token-lifecycle` 正在改 `frontend/mobile` 的 `request.js` / `Login.vue` / `stores/app.js`，本 change 必须等它归档后再动这些文件，避免合并冲突。
- AGENTS.md「优先选可跨项目复用的通用依赖」判据：后端不引微信 SDK（`net/http` 自包即可），前端 uniapp 系列依赖是 uniapp 生态的官方包，登记到 deps.yaml 即可。

## Goals / Non-Goals

**Goals:**
- 让基座真正「多端」：第四端 miniapp/ 平级落地，三端铁律（stores/、@→src/、snake_case）顺势推广到多端语境。
- miniapp 端的登录链路与基座 JWT/RBAC 无缝衔接：`mp-login` 签发的 token 与 `username/password` 登录签发的是同一种 JWT，所有中间件无差异对待。
- 工程基础设施（deps.yaml / Makefile / eslint / init.sh / .gitignore / README / AGENTS）按四端扩展。
- 最小范例（`pages/login` + `pages/index`）跑通"登录 → 拉系统信息"完整链路，作为新增业务页的黄金路径。

**Non-Goals:**
- 不做 uniapp 编译到 H5 的目标（未来可单独开 change 用 uniapp 一统移动端、废 mobile）。
- 不做 App 平台编译（iOS/Android）。
- 不做 unionid 跨主体去重、不做手机号授权、不做订阅消息、不做小程序码生成。
- 不扩展 `scripts/package.sh` 把 miniapp 纳入 deploy/——小程序发布走微信开发者工具上传，与 deploy/ tarball 是两条独立链路。
- 不引入任何微信第三方 SDK；后端用 `net/http` 直调 `jscode2session`。
- 本 change 不动 `auth-token-lifecycle` 已规划的 `request.js` 续期逻辑——miniapp 的 `request.js` 只做登录 + 基础请求封装，续期留给后续 change 协同。

## Decisions

### D1: 目录命名为 `miniapp/`（而非 `mp/` / `mp-weixin/` / `mobile/mp/`）

| 候选 | 评 |
|---|---|
| `miniapp/` ✅ | 平级最直观；语义清楚；与 frontend/mobile 风格一致（小写单词） |
| `mp/` | 太短；语义模糊（mp = ?） |
| `mp-weixin/` | 太具体，未来想覆盖支付宝/抖音小程序要改名 |
| `mobile/mp/` | 把 mobile/ 重构成聚合目录，要挪现有 Vant H5 文件，破坏性大 |
| 替换 mobile/ | 废 Vant H5，破坏向后兼容，基座定位变了 |

`miniapp/` 一词兼顾未来扩展（支付宝/抖音/百度小程序都可以编译到这个目录），且不暗示具体平台。

### D2: 不强求 `views/`，保留 `pages/` 作为 miniapp 目录约定

基座三端铁律（`AGENTS.md §1`）只约束 `stores/`（复数）、`@` → `src/`、snake_case JSON tag，**未约束 `views/`**——这正好留了口子。uniapp 的 `pages.json` 是硬约定（页面路径必须在 `pages.json` 注册，文件位置约定在 `pages/` 下），强扭成 `views/` 反而让 uniapp 工具链难用。所以 miniapp 端**显式采用 `pages/`**，与基座铁律不冲突。

### D3: HTTP 层用 `uni.request` 自包，不引 axios

| 候选 | 评 |
|---|---|
| `uni.request` 自包 ✅ | 地道，零适配成本；uni-app 文档主推；包一层即可对齐基座 `Authorization` + 401 处理约定 |
| axios + uni-app adapter | 要引 `axios` + 适配器，依赖变重；uni-app 多端编译时 adapter 行为不一 |
| axios + polyfill XMLHttpRequest | 仅 H5 编译可用，小程序环境没有 XHR |

走 `uni.request` 后，eslint 的"禁止直接 `import axios`"规则在 miniapp 端**不生效**（miniapp 不引 axios，规则自然不会触发）；同时新增一条"禁止直接调 `uni.request`"规则，强制走 `@/utils/request` 封装，与基座"请求必须走封装层"的语义对齐。这两条规则在 flat config 里按 files 段差异化配置。

### D4: `User.openid` 字段而非独立 `mp_users` 表

| 候选 | 评 |
|---|---|
| `User.OpenID` 字段 ✅ | 单一用户实体，与 username/password 登录共存；JWT claims 与 RBAC 链路零改动；GORM AutoMigrate 加列即可 |
| 独立 `mp_users` 表 | 用户实体分裂，权限/角色要二次映射，复杂度陡增；微信用户也要走 RBAC，没必要单独建表 |

`OpenID` 字段可空（老用户为空）、唯一索引（防止同 openid 重复建号）。空值表示该用户从未用过小程序登录；非空值表示该用户除原登录方式外，也可由 mp-login 命中。

### D5: 后端不引微信 SDK，用 `net/http` 自包 `jscode2session`

`jscode2session` 是一个 GET `https://api.weixin.qq.com/sns/jscode2session?appid=X&secret=Y&js_code=Z&grant_type=authorization_code`，返回 JSON `{openid, session_key, unionid?, errcode?, errmsg?}`。标准库即可，无须 `github.com/silenceper/wechat` 之类 SDK。

封装为 `backend/utils/wechat.go::JsCode2Session(appid, secret, code) (openid, sessionKey string, err error)`，便于测试 mock。

### D6: `wechat` 配置段不支持环境变量覆盖（与 `app` 品牌段一致）

按 `brand-config` 的判据——"环境变量通道仅适用于 `server` / `database` / `jwt` 三段"——`wechat` 段也不走 env。理由：
- 小程序 `secret` 是高敏感凭证，env 容易通过子进程 / `ps aux` / 容器 inspect 泄露。
- `config.yaml` 由部署侧管控权限，比 env 更稳。
- 与品牌段的判据对齐，不引入新的特例。

未配置时（`appid == "" || secret == ""`）`mp-login` 返回 500 + 明确指引，**不**让后端启动失败（基座要能开箱运行，不强制要求配 wechat 才能起服务）。

### D7: eslint flat config 用 files 段差异化规则

uniapp 用 `uni.request` 而非 axios，axios 禁令对 miniapp 不适用。flat config 可以按 `files: ['miniapp/src/**']` 段关闭 axios 规则、开启 `uni.request` 禁令。具体形式：

```js
// 现有：全局禁 axios + 禁 @/store/ 单数
{
  rules: {
    'no-restricted-imports': ['error', { paths: [axios], patterns: [@/store] }],
  },
},
// 新增：miniapp 段差异化
{
  files: ['miniapp/src/**'],
  rules: {
    // 关闭 axios 禁令（miniapp 不引 axios）
    'no-restricted-imports': ['error', {
      patterns: [
        { group: ['@/store', '@/store/*'], message: '...' },
        // 新增：禁止直接调 uni.request，但需要用 no-restricted-syntax 而非 no-restricted-imports
      ],
    }],
  },
},
// 新增：miniapp 的 uni.request 禁令（用 no-restricted-syntax 抓 CallExpression）
{
  files: ['miniapp/src/**', '!miniapp/src/utils/request.js'],
  rules: {
    'no-restricted-syntax': ['error', {
      selector: "CallExpression[callee.object.name='uni'][callee.property.name='request']",
      message: '禁止直接调用 uni.request，请使用 @/utils/request 的封装实例',
    }],
  },
},
// 豁免：miniapp/src/utils/request.js 自身
{
  files: ['miniapp/src/utils/request.js'],
  rules: { 'no-restricted-syntax': 'off' },
},
```

这样三端规则在同一份 flat config 里差异化呈现，仍保留"单一真相"特性。

### D8: `scripts/package.sh` 不扩展，小程序发布走独立链路

小程序发布流程：
1. `npm run build:mp-weixin` 产出 `miniapp/dist/build/mp-weixin/`。
2. 微信开发者工具打开该目录，点"上传"，提审，发布。

这条链路与 `deploy/` tarball + 启动脚本是**两条独立流程**——后者是后端 + 前端静态资源的服务端部署，前者是小程序包上传到微信平台。强行把 miniapp 纳入 `package.sh` 反而让人误以为 `deploy/` 里能跑小程序。所以本 change 明确不动 `package.sh`，只新增 `Makefile.build-mp` 作为构建入口。

### D9: `make dev` 不并行起 miniapp dev

`make dev` 现在并行起 backend + frontend（`$(MAKE) -j2 dev-backend dev-frontend`）。小程序 dev 需要微信开发者工具配合（产出目录 + 工具打开），不像 H5 那样 dev server 直接打开浏览器。强行并行起 miniapp dev 会让人误以为有 H5 风格的"打开浏览器即用"体验，反而误导。所以：
- `make dev` 不变（backend + frontend 并行）。
- 新增独立 `make dev-mp`（仅起 miniapp dev），按需手动起。
- 新增独立 `make build-mp`（构建 mp-weixin 产物）。

### D10: 前置依赖 `auth-token-lifecycle`

`auth-token-lifecycle` 正在改 `frontend/mobile` 的 `request.js`（加 401 静默刷新 + 单飞）、`stores/app.js`、`Login.vue`，并新增 `User.TokenVersion` 字段 + `POST /api/auth/refresh`。

本 change 的 mp-login 签发的 JWT 也要：
- 携带 `Ver` 字段（受版本吊销约束）。
- 能被 `/api/auth/refresh` 续期。
- miniapp 端的 `utils/request.js` 也要做 401 静默刷新 + 单飞。

两条路径高度重叠。所以本 change 在 `auth-token-lifecycle` 归档前**只做最小骨架**：
- miniapp 目录 + 最小范例页面（不接入 refresh 逻辑）。
- 后端 mp-login 接口（签发的 token 与 `/api/auth/login` 同结构，但 `Ver` 字段等 `auth-token-lifecycle` 落地后自动带上）。
- `request.js` 只做 `Authorization` 头 + 401 跳登录，不做静默刷新。

`auth-token-lifecycle` 归档后，再单独开 follow-up change 把 miniapp 接入 refresh 链路。

## Risks / Trade-offs

- **[Risk] `User.openid` 字段加唯一索引在 SQLite 上可能影响老库迁移性能** → 老 `User` 表行数小（基座典型 < 100 行），加列 + 唯一索引开销可忽略；迁移失败时 GORM 会回滚报错。
- **[Risk] eslint `no-restricted-syntax` 用 CSS-selector 抓 `uni.request` 调用，对 `const r = uni.request; r(...)` 这种间接调用漏判** → 接受；这是规则的天花板，AI 写间接调用属于有意绕过，由 code review 兜底。基座现有 axios 禁令也有同样的天花板（`import { default as ax } from 'axios'` 之类）。
- **[Risk] uniapp vite 编译产物的版本兼容性** → 锁定 `@dcloudio/vite-plugin-uni` 与 uni-app 主版本一致（参考官方版本对照表），登记到 deps.yaml 后由 lockfile 锁死。
- **[Trade-off] miniapp 不消费 `favicon` 字段** → 小程序无浏览器标签概念，`brand-config-guard` 扩展到 miniapp 时对 `favicon` 字段做"忽略不报错"处理，不强制消费。
- **[Trade-off] miniapp 不做菜单级 RBAC** → 小程序页面通常无侧边栏菜单，权限校验靠后端 `PermissionRequired` 兜底；前端 `v-permission` 指令可在业务页按需引入（不在本 change 范围）。`frontend-rbac` spec 的"扫描范围只 frontend/src"保持不动，不在 miniapp 引入 RBAC 扫描。
- **[Trade-off] mp-login 接口未配置时返回 500 而非 503** → 500 + 明确指引比 503 更易诊断；503 语义偏"暂时不可用"，而未配置是部署侧问题，500 更准确。

## Migration Plan

部署侧：
1. 升级后端到含本 change 的版本。
2. 在 `backend/config.yaml` 加 `wechat.appid` + `wechat.secret`（可暂留空，开发态不影响后端启动）。
3. 后端首次启动 GORM AutoMigrate 给 `users` 表加 `openid` 列（NULL 默认）。
4. 小程序侧：clone 后 `cd miniapp && npm install`，在 `miniapp/src/manifest.json` 填 `mp-weixin.appid`，`npm run dev:mp-weixin` 用微信开发者工具打开 `miniapp/dist/dev/mp-weixin/`。

回滚：
- 后端：回退到本 change 之前的 commit，`users.openid` 列保留无害（GORM 不会主动删列），老逻辑不受影响。
- 前端：删除 `miniapp/` 目录、回退 `deps.yaml` / `Makefile` / `eslint.config.js` / `scripts/init.sh` / `.gitignore` / `README.md` / `AGENTS.md` 即可。

## Open Questions

1. **miniapp 端是否要支持 `v-permission` 指令？** 当前 `frontend-rbac` spec 明确"不扫描 mobile/src"，miniapp 也类似——但小程序业务页若要做按钮级权限，需要 miniapp 版的 `v-permission`。本 change 不做，留给后续按需开 change。
2. **`mp-login` 是否要支持 unionid 跨主体去重？** 当前 spec 只匹配 openid；若未来要支持"同一微信开放平台主体下，不同小程序共享用户"，需要加 unionid 字段。本 change 明确排除。
3. **miniapp 是否要接入 `auth-token-lifecycle` 的 refresh 链路？** 当前 spec 只做基础请求封装；refresh 接入留给 `auth-token-lifecycle` 归档后的 follow-up change。
