## MODIFIED Requirements

### Requirement: 运行时标识与密钥重置

初始化 MUST 替换运行时标识（env var 前缀 `BASE_BACKEND_*`、`package.sh` 二进制名/压缩包名、JWT Issuer），MUST 将默认密钥替换为随机串，并 MUST 改写 `miniapp/package.json` 的 `name` 与 `miniapp/src/manifest.json` 的 `name` 字段为新项目名。`miniapp/src/manifest.json` 的 `mp-weixin.appid` MUST 保留基座占位，由具体项目开发者填入。

#### Scenario: 环境变量前缀随项目名

- **WHEN** 项目名含连字符（如 `my-system`）
- **THEN** env var 前缀规范化为大写去连字符（`MY_SYSTEM_`），如 `MY_SYSTEM_SERVER_PORT`

#### Scenario: 默认密钥被替换

- **WHEN** 初始化执行完成
- **THEN** `config.go` 默认值与 `config.example.yaml` 中的 `base-backend-secret-key-change-me` 被替换为随机生成的密钥

#### Scenario: 数据库名默认不动

- **WHEN** 未传 `--db-name`
- **THEN** 数据库默认名保持 `base_backend.db` / `base_backend` 不变；仅当显式传 `--db-name` 时替换

#### Scenario: miniapp 包名与 manifest 改写

- **WHEN** 执行 `make init name=my-system`
- **THEN** `miniapp/package.json` 的 `name` 改为 `my-system-miniapp`，`miniapp/src/manifest.json` 的 `name` 改为 `my-system`，`mp-weixin.appid` 不被改写

#### Scenario: appid 不被脚本改写

- **WHEN** 初始化执行完成
- **THEN** `miniapp/src/manifest.json` 的 `mp-weixin.appid` 保留为基座占位（如 `touristappid` 或空串），由具体项目开发者填入
