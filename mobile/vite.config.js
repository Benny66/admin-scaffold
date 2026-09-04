import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBase = env.VITE_API_BASE || 'http://localhost:8080'

  return {
    // 生产构建按 /m/ 子路径发布（后端在 /m/ 托管 dist-mobile，见 static-serving spec）。
    // dev 保持 '/'，不影响 http://localhost:5174/ 直接访问。
    // import.meta.env.BASE_URL 会注入该值，mobile/src/router 的 history base 与之同源，
    // 因此 base 是唯一配置点，不要单独在 router 里写死 base。
    base: command === 'build' ? '/m/' : '/',
    plugins: [
      vue(),
      Components({
        resolvers: [VantResolver()],
      }),
    ],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      port: 5174,
      proxy: {
        '/api': {
          target: apiBase,
          changeOrigin: true,
        },
        '/static': {
          target: apiBase,
          changeOrigin: true,
        },
      },
    },
  }
})
