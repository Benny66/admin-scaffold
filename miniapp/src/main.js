import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'

// uniapp vue3 入口（与 frontend/mobile 的 createApp 写法不同，须用 createSSRApp）
//
// 多端铁律（AGENTS.md §1 多端统一）：
//   - stores/ 目录（复数，由 eslint 护栏）
//   - @ 别名指向 src/（vite.config.js 已配置）
export function createApp() {
  const app = createSSRApp(App)
  app.use(createPinia())
  return {
    app,
  }
}
