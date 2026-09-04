# multi-platform-packaging delta — fix-mobile-deploy-runtime

## MODIFIED Requirements

### Requirement: 部署包组装

打包脚本 MUST 组装 `deploy/` 目录，包含：后端二进制、`dist/`（前端）、`dist-mobile/`（移动端）、`static/`（品牌静态资源，拷贝自 `backend/static/`）、`config.yaml`（从 example 复制并强制 release）、启动脚本（start.sh / start.bat）。

#### Scenario: 部署目录结构完整

- **WHEN** 打包完成
- **THEN** `deploy/` 下存在后端二进制、dist、dist-mobile、static、config.yaml、启动脚本，缺一不可

#### Scenario: 配置为 release 模式

- **WHEN** 生成的 `deploy/config.yaml`
- **THEN** 其 `server.mode` 为 `release`，而非开发默认的 `debug`

#### Scenario: 品牌资源随包部署可访问

- **WHEN** 打包后在 `deploy/` 下启动后端，请求 `/static/logo.png`
- **THEN** 返回 `deploy/static/logo.png`（200），品牌资源不因部署而 404

## ADDED Requirements

### Requirement: 移动端产物按 /m/ 子路径构建

`dist-mobile/` 移动端构建产物 MUST 以 `/m/` 作为资源 base（`vite build` 的 `base: '/m/'`），使产出的 `index.html` 引用的 JS/CSS 等静态资源以 `/m/assets/*` 前缀呈现，能被后端 `/m/` 子路径托管完整加载。

#### Scenario: 产物资源路径带 /m/ 前缀

- **WHEN** 执行移动端生产构建
- **THEN** `mobile/dist/index.html` 引用的 `script`/`link` 资源 URL 以 `/m/assets/` 开头（而非 `/assets/`）

#### Scenario: 打包部署后移动端可完整渲染

- **WHEN** 使用 `make package` 打包并在 `deploy/` 下启动后端，浏览器访问 `/m/`
- **THEN** 页面返回 200 且其引用的 JS/CSS 资源均能正常加载（非桌面 HTML 回退、非 404），移动端页面可渲染并完成登录
