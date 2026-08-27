// 根目录统一 ESLint 配置（flat config）—— 前端与移动端共享的「架构护栏」。
// 把 frontend/CLAUDE.md 里的域宪法编译成会失败的静态检查：
//   1. 禁止直接 import axios（必须走 @/utils/request 封装）
//   2. 禁止 @/store/ 单数（状态管理目录统一 @/stores/）
// 与后端 internal/guard/ 的 guard 测试对称：违反即 CI 变红，而非靠 code review。
import js from '@eslint/js'
import vue from 'eslint-plugin-vue'

export default [
  // 忽略构建产物与依赖目录
  {
    ignores: ['**/dist/**', '**/node_modules/**'],
  },

  // JS 基础推荐规则（catch 参数豁免：catch(e) 未用 e 是常见风格，不算违规）
  {
    ...js.configs.recommended,
    rules: {
      ...js.configs.recommended.rules,
      'no-unused-vars': ['error', { caughtErrors: 'none' }],
    },
  },

  // Vue3 essential 规则（关闭不适用于本项目的规则）
  ...vue.configs['flat/essential'],

  // 架构护栏规则
  {
    rules: {
      // 单文件组件命名：本项目用路由懒加载，views/xxx/index.vue、Login.vue 等
      // 单词组件是标准做法，multi-word 规则不适用。
      'vue/multi-word-component-names': 'off',

      // 禁止直接 import axios，必须走 @/utils/request 封装
      'no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: 'axios',
              message: '禁止直接导入 axios，请使用 @/utils/request 的封装实例（见 frontend/CLAUDE.md 第 3 节）',
            },
          ],
          patterns: [
            {
              group: ['@/store', '@/store/*'],
              message: '状态管理目录必须统一为 @/stores/（复数），禁止 @/store/ 单数回潮',
            },
          ],
        },
      ],
    },
  },

  // 豁免：utils/request.js 是请求封装层本身，可合法 import axios
  {
    files: ['**/utils/request.js'],
    rules: {
      'no-restricted-imports': 'off',
    },
  },

  // 豁免：vite.config.js 是 Node 环境，__dirname 由 vite 注入，不算未定义
  {
    files: ['**/vite.config.js'],
    languageOptions: {
      globals: {
        __dirname: 'readonly',
      },
    },
  },
]
