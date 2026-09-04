// 根目录统一 ESLint 配置（flat config）—— 前端、移动端与小程序端共享的「架构护栏」。
// 把 frontend/CLAUDE.md 里的域宪法编译成会失败的静态检查：
//   1. 禁止直接 import axios（必须走 @/utils/request 封装）
//   2. 禁止 @/store/ 单数（状态管理目录统一 @/stores/）
//   3. 菜单必须归组（menu-grouping：frontend/src/router/index.js 的两级结构）
// 与后端 internal/guard/ 的 guard 测试对称：违反即 CI 变红，而非靠 code review。
import js from '@eslint/js'
import vue from 'eslint-plugin-vue'
import menuGroupRule from './eslint-rules/menu-group.js'

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

  // 菜单结构护栏（menu-grouping）：仅作用于 frontend/src/router/index.js。
  // 通过局部插件注册自定义 AST 规则，把「菜单必须归组、字段必须齐全、分组可达」
  // 编译成红灯；解析不到根路由时也必须报错（延续 guard「护栏要能感知自己瞎了」）。
  // files 用 **/src/router/index.js 形态以同时匹配「从根目录跑」与「从 frontend/
  // 子目录跑 npm run lint」两种 cwd；但 mobile/miniapp 等端的 router/index.js 不
  // 参与两级菜单结构（mobile 是单页路由），故用 ignores 显式排除三端目录。
  // ignores 同时列出两种形态：`mobile/**` 覆盖「从根目录跑」时路径以 mobile/ 开头；
  // `src/router/index.js` 是兜底——但兜底太广会误伤 frontend，故采用更精确的方案：
  // 在规则体内检查「若 path:'/' 路由无 children 数组，跳过检查（mobile 风格）」。
  {
    files: ['**/src/router/index.js'],
    ignores: [
      '**/mobile/**',
      'mobile/**',
      '**/miniapp/**',
      'miniapp/**',
      '**/miniapp-wechat-end/**',
      'miniapp-wechat-end/**',
    ],
    plugins: {
      local: {
        rules: {
          'menu-group': menuGroupRule,
        },
      },
    },
    rules: {
      'local/menu-group': 'error',
    },
  },
]
