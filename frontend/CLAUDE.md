# frontend/CLAUDE.md — 前端域宪法

AI 在 `frontend/`（Web 端）下编写 Vue3 代码时 MUST 遵守本文件。跨端通用规则见根 [`AGENTS.md`](../AGENTS.md)。

> **护栏强制**：下列标注 ⚙️ 的规则已由根 `eslint.config.js` 编译成会失败的 ESLint 检查（`make lint`），违反即 CI 变红，而非靠人工 code review。

---

## 1. 命名与目录约定

- ⚙️ 状态管理目录 MUST 统一为 `stores/`（禁止 `store/` 单复数混用）—— 由 `no-restricted-imports` 强制。
- 页面按业务域分目录：`views/<domain>/`，公共组件放 `components/`。
- API 定义集中在 `api/index.js`，请求封装统一走 `utils/request.js`。
- 路径别名 `@` 指向 `src/`。

## 2. 组件风格

- 使用 `<script setup>` 语法。
- 模板里复杂表达式抽成计算属性或方法。
- 每个页面组件：`<template>` → `<script setup>` → `<style scoped>` 三段式。
- 组件文件 PascalCase（如 `UserList.vue`）；JS 模块 camelCase；目录 kebab-case。

## 3. 请求封装规范

- 所有 HTTP 请求 MUST 走 `utils/request.js` 的 axios 实例。
- MUST 自动附带 `Authorization: Bearer <token>`。
- 响应拦截器 MUST 统一处理：`code === 200` 返回 `res`；401 清 token 跳登录；403 提示无权限；blob 直接返回。
- ⚙️ 禁止绕过封装直接 `import axios` —— 由 `no-restricted-imports` 强制（`utils/request.js` 自身是唯一豁免）。

## 4. 权限控制

- 登录后将 `permissions`（权限码数组）与用户信息存入 Pinia 与 localStorage。
- `appStore.hasPermission(code)` 用于按钮级/菜单级控制。
- 路由守卫：未登录访问受保护页 → 跳 `/login`。

## 5. 系统名称

- 系统名称 MUST 来自 `appStore.systemName`（由 `/system/info` 拉取），禁止硬编码「企业管理系统」字符串。
- `index.html` 的 `<title>` 是中性占位（`Base Admin`），运行时由 `fetchSystemInfo` + `document.title` 覆盖为真实品牌名。
