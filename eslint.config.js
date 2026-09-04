// 根目录统一 ESLint 配置（flat config）—— 前端、移动端与小程序端共享的「架构护栏」。
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

  // miniapp 段差异化规则（miniapp-wechat-end spec / frontend-guardrails spec）：
  //   - 关闭 axios 禁令（miniapp 不引 axios，地道写法是 uni.request）
  //   - 保留 @/store/ 单数禁令（多端统一铁律）
  //   - uni 是 uniapp 全局对象，声明为 readonly 全局变量
  {
    files: ['miniapp/src/**'],
    languageOptions: {
      globals: {
        uni: 'readonly',
      },
    },
    rules: {
      'no-restricted-imports': [
        'error',
        {
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

  // miniapp 禁止直接调 uni.request，强制走 @/utils/request 封装
  // （封装层自身在下面单独豁免）
  {
    files: ['miniapp/src/**', '!miniapp/src/utils/request.js'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: "CallExpression[callee.object.name='uni'][callee.property.name='request']",
          message: '禁止直接调用 uni.request，请使用 @/utils/request 的封装实例',
        },
      ],
    },
  },

  // 豁免：miniapp/src/utils/request.js 是封装层本身，可合法调用 uni.request
  {
    files: ['miniapp/src/utils/request.js'],
    rules: {
      'no-restricted-syntax': 'off',
    },
  },

  // 豁免：vite.config.js 运行在 Node 环境，不参与浏览器打包，
  // 故 __dirname 与 process 都不是未定义变量（src/ 下仍视为未定义）。
  {
    files: ['**/vite.config.js'],
    languageOptions: {
      globals: {
        __dirname: 'readonly',
        process: 'readonly',
      },
    },
  },
]
