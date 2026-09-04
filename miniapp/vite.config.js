import { defineConfig, loadEnv } from 'vite'
import uniPluginImport from '@dcloudio/vite-plugin-uni'
import { resolve } from 'path'

// @dcloudio/vite-plugin-uni 是 CJS 包，ESM import 时 default 未必拿到（取决于 Node
// 与 esbuild 的 interop 行为）。兼容写法：取 .default 或本体。
const uni = uniPluginImport.default || uniPluginImport

// uniapp + vite 配置
//
// 多端铁律（AGENTS.md §1 多端统一）：
//   - @ 别名指向 src/（与 frontend/mobile 一致）
//   - stores/ 复数目录（强制，由 eslint flat config 护栏）
//   - 后端字段 JSON tag 沿用 snake_case 透传
//
// 页面目录用 pages/ 而非 views/：uniapp 的 pages.json 是硬约定，文件位置必须在
// pages/ 下，强扭反而不地道（见 miniapp-wechat-end spec 的「多端统一铁律」要求）。
//
// 小程序请求不走 vite dev proxy（受限于微信小程序的网络请求白名单机制），
// 业务代码 MUST 通过 `import.meta.env.VITE_API_BASE` 读取后端地址，
// dev 阶段填本地 http://localhost:8080，生产部署填真实域名（须在小程序后台
// 配置 request 合法域名）。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBase = env.VITE_API_BASE || 'http://localhost:8080'

  return {
    plugins: [uni()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    define: {
      // 注入到运行时：业务代码读 import.meta.env.VITE_API_BASE 拿后端地址
      __API_BASE__: JSON.stringify(apiBase),
    },
  }
})
