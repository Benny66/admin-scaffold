# Tasks: fix-brand-hardcode

## 1. index.html 去硬编码

- [x] 1.1 `frontend/index.html` 的 `<title>` 改为 `Base Admin`
- [x] 1.2 `mobile/index.html` 的 `<title>` 改为 `Base Admin`

## 2. init.sh 支持中文品牌替换

- [x] 2.1 `scripts/init.sh` 新增 `--app-name` 参数解析与提示
- [x] 2.2 在替换逻辑中加入「企业管理系统」→ app-name 的替换（缺省跳过）
- [x] 2.3 更新 init.sh 头部注释与 README 的「如何基于基座新建项目」说明

## 3. 运行时 title 覆盖 + 文档一致性

- [x] 3.1 确认前端 `fetchSystemInfo` 后 `document.title` 覆盖逻辑存在（router.beforeEach 已设）
- [x] 3.2 确认移动端是否设 `document.title`，若缺失则补上（fetchSystemInfo 后设品牌名）
- [x] 3.3 更新 `frontend/CLAUDE.md` 说明 index.html title 为中性占位、运行时覆盖

## 4. 验证

- [x] 4.1 全局搜「企业管理系统」，确认代码文件无残留（仅 init.sh 替换逻辑 + 文档说明）
- [x] 4.2 前端/移动端 build 通过，启动后标签页标题为运行时品牌名
