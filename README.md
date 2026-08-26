# Base 脚手架基座

> 企业管理系统三端脚手架基座：Go + Gin + GORM + Vue3 + Element Plus + Vant

这是一个**可运行的三端 monorepo 骨架**，从企业固定资产管理系统中提炼出的、与业务无关的可复用横切能力与工程约定。clone 后改名即可开始开发新项目。

---

## 定位

```
┌──────────────────────────────────────────────────────────────┐
│                    base/ 脚手架基座                            │
├──────────────┬──────────────┬────────────────────────────────┤
│   backend/   │  frontend/   │         mobile/                 │
│  Go + Gin    │  Vue3 +      │       Vue3 + Vant               │
│  GORM        │  Element Plus│       移动端骨架                 │
│  三层分层    │  Pinia       │                                 │
└──────────────┴──────────────┴────────────────────────────────┘
              │                    │
              └──────────┬─────────┘
                         ▼
              统一响应协议 + 统一鉴权 + RBAC
```

基座内置「系统管理五件套」作为开箱即用能力与最佳实践示范：

- 用户管理（CRUD、启用/禁用、密码重置、角色分配）
- 角色管理（CRUD、权限配置）
- 权限管理（CRUD、角色授权）
- 字典管理（字典类型、字典项维护）
- 操作日志（操作日志、登录日志查询）

---

## 目录结构

```
base/
├── backend/                 # Go 后端（controller → service → model 三层）
│   ├── config/              # 配置（默认值 → YAML → 环境变量）
│   ├── database/            # SQLite/MySQL 双驱动 + 自动迁移 + 基础数据
│   ├── models/              # 数据模型
│   ├── middleware/          # JWT/CORS/日志/RBAC 中间件
│   ├── services/            # 业务逻辑层
│   ├── controllers/         # 控制器层（参数校验 + 调 service + 响应）
│   ├── router/              # 路由注册
│   ├── utils/               # 工具（响应/JWT/加密等）
│   └── main.go
├── frontend/                # Web 前端（Vue3 + Element Plus）
│   └── src/
│       ├── api/             # API 定义
│       ├── layout/          # 布局
│       ├── router/          # 路由 + 守卫
│       ├── stores/          # Pinia 状态（统一 stores/）
│       ├── utils/           # 请求封装等
│       └── views/           # 页面（Login + system/ 五件套）
├── mobile/                  # 移动端（Vue3 + Vant）
├── docs/                    # 规范文档
│   ├── 代码规范.md
│   ├── 目录结构约定.md
│   ├── API协议.md
│   ├── 鉴权与权限.md
│   ├── 后端分层规范.md
│   └── 配置体系.md
├── AGENTS.md                # AI 开发规范（宪法）
└── README.md
```

---

## 快速启动

### 环境要求

| 工具 | 版本要求 |
|------|---------|
| Go | >= 1.21 |
| Node.js | >= 18.0 |
| npm | >= 9.0 |

### 第一步：启动后端

```bash
cd backend
go mod tidy
go run main.go
```

启动成功后输出默认管理员账号 `admin / admin123`。

### 第二步：启动前端

```bash
cd frontend
npm install
npm run dev
```

访问 http://localhost:5173，使用 `admin / admin123` 登录。

### 第三步（可选）：启动移动端

```bash
cd mobile
npm install
npm run dev
```

---

## 如何基于基座新建项目

1. **复制目录**：将整个 `base/` 复制为新项目目录，如 `my-system/`。
2. **改名**：
   - 后端：修改 `backend/go.mod` 的 module 名（当前 `base-backend`），并全局替换所有 `import "base-backend/..."` 为新模块名。
   - 前端：修改 `frontend/package.json` 与 `mobile/package.json` 的 `name`。
3. **清理初始化数据**：删除运行时生成的 `backend/*.db` 与 `backend/config.yaml`（基座默认不携带）。
4. **配置**：按需修改 `backend/config.example.yaml` 为 `config.yaml`，设置端口、JWT 密钥、数据库类型。
5. **开发**：在 `services/` + `controllers/` + `router/` 中按现有范式新增业务模块，前端在 `views/` 下新增页面。

> 详细约定见 `docs/` 与 `AGENTS.md`。

---

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.21+ / Gin / GORM / SQLite(纯Go) + MySQL 双驱动 / JWT / bcrypt / yaml |
| 前端 | Vue3 / Vite / Element Plus / Pinia / Vue Router / axios |
| 移动端 | Vue3 / Vite / Vant / Pinia / Vue Router / axios |

---

## 特性说明

- **三层分层**：`controller → service → model`，职责边界清晰，业务可测试。
- **RBAC 权限**：用户-角色-权限三级模型，`PermissionRequired` / `AdminRequired` 两级中间件。
- **统一响应**：`{code, message, data}` + 分页 `{list, total, page, page_size}`。
- **三层配置**：默认值 → YAML → 环境变量，逐层覆盖。
- **双数据库**：SQLite（开箱即用）+ MySQL（生产），自动迁移建表 + 初始化基础数据。
