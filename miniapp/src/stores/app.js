import { defineStore } from 'pinia'
import { getSystemInfo } from '@/api'

// miniapp 端全局 store（与 frontend/mobile 的 stores/app.js 同构）
//
// 多端铁律（AGENTS.md §1 多端统一）：
//   - stores/ 目录（复数，由 eslint flat config 护栏）
//   - @ 别名指向 src/（vite.config.js 已配置）
//   - 后端字段 JSON tag 沿用 snake_case 透传
//
// 持久化用 uni.getStorageSync / uni.setStorageSync（小程序无 localStorage）。
// token 由 stores/app.js 统一管理（setToken 写入），utils/request.js 从 storage 读取。
export const useAppStore = defineStore('app', {
  state: () => ({
    systemName: uni.getStorageSync('systemName') || '',
    logo: uni.getStorageSync('logo') || '',
    token: uni.getStorageSync('token') || '',
    userInfo: uni.getStorageSync('userInfo') ? JSON.parse(uni.getStorageSync('userInfo')) : null,
    permissions: uni.getStorageSync('permissions') ? JSON.parse(uni.getStorageSync('permissions')) : [],
    // logo 加载失败标记（如 config 配了但后端 static 下文件不存在）。不持久化。
    logoFailed: false,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token && !!state.userInfo,
    isAdmin: (state) => !!state.userInfo?.is_admin,
    // logo 展示点共用判断，失败后回退文字
    logoAvailable: (state) => !!state.logo && !state.logoFailed,
  },
  actions: {
    setToken(token) {
      this.token = token
      uni.setStorageSync('token', token)
    },
    setUserInfo(userInfo) {
      this.userInfo = userInfo
      uni.setStorageSync('userInfo', JSON.stringify(userInfo))
    },
    setPermissions(permissions) {
      this.permissions = permissions
      uni.setStorageSync('permissions', JSON.stringify(permissions))
    },
    setSystemName(name) {
      this.systemName = name
      uni.setStorageSync('systemName', name)
    },
    setLogo(logo) {
      this.logo = logo
      uni.setStorageSync('logo', logo)
    },
    markLogoFailed() {
      this.logoFailed = true
    },
    async fetchSystemInfo() {
      try {
        const res = await getSystemInfo()
        if (res.code === 200 && res.data) {
          // 解构消费 GetSystemInfo 返回的全部字段（brand-config guard 强制）。
          // miniapp 实际只使用 name（→ systemName）与 logo；favicon / login_bg 等
          // 字段对 miniapp 端「忽略不报错」（小程序无浏览器标签概念）。
          // eslint-disable-next-line no-unused-vars
          const { name, subtitle, logo, favicon, footer, login_bg, login_bg_mobile } = res.data
          this.setSystemName(name || 'Base Admin')
          this.setLogo(logo || '')
        }
      } catch (e) {
        // 静默失败，使用默认值
      }
    },
    logout() {
      this.token = ''
      this.userInfo = null
      this.permissions = []
      uni.removeStorageSync('token')
      uni.removeStorageSync('userInfo')
      uni.removeStorageSync('permissions')
    },
  },
})
