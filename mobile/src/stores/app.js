import { defineStore } from 'pinia'
import { getSystemInfo } from '@/api'

export const useAppStore = defineStore('app', {
  state: () => ({
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
    permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
    systemName: localStorage.getItem('systemName') || '',
    subtitle: localStorage.getItem('subtitle') || '',
    logo: localStorage.getItem('logo') || '',
    footer: localStorage.getItem('footer') || '',
    loginBg: localStorage.getItem('loginBg') || '',
    loginBgMobile: localStorage.getItem('loginBgMobile') || '',
    // logo 加载失败标记（如 config 配了但文件不存在）。不持久化：补上文件后刷新即恢复。
    logoFailed: false,
  }),
  getters: {
    isLoggedIn: (state) => !!state.userInfo && !!localStorage.getItem('token'),
    // 所有 logo 展示点共用同一个判断，避免每个组件各存一份失败状态
    logoAvailable: (state) => !!state.logo && !state.logoFailed,
  },
  actions: {
    setUserInfo(userInfo) {
      this.userInfo = userInfo
      localStorage.setItem('userInfo', JSON.stringify(userInfo))
    },
    setPermissions(permissions) {
      this.permissions = permissions
      localStorage.setItem('permissions', JSON.stringify(permissions))
    },
    setSystemName(name) {
      this.systemName = name
      localStorage.setItem('systemName', name)
    },
    setSubtitle(subtitle) {
      this.subtitle = subtitle
      localStorage.setItem('subtitle', subtitle)
    },
    setLogo(logo) {
      this.logo = logo
      localStorage.setItem('logo', logo)
    },
    setFooter(footer) {
      this.footer = footer
      localStorage.setItem('footer', footer)
    },
    setLoginBg(loginBg) {
      this.loginBg = loginBg
      localStorage.setItem('loginBg', loginBg)
    },
    setLoginBgMobile(loginBgMobile) {
      this.loginBgMobile = loginBgMobile
      localStorage.setItem('loginBgMobile', loginBgMobile)
    },
    // 任一展示点加载失败即标记，其余展示点随之回退到文字，不再产生破图
    markLogoFailed() {
      this.logoFailed = true
    },
    async fetchSystemInfo() {
      try {
        const res = await getSystemInfo()
        if (res.code === 200 && res.data) {
          const { name, subtitle, logo, favicon, footer, login_bg, login_bg_mobile } = res.data
          this.setSystemName(name || 'Base Admin')
          this.setSubtitle(subtitle || '')
          this.setLogo(logo || '')
          this.setFooter(footer || '')
          this.setLoginBg(login_bg || '')
          this.setLoginBgMobile(login_bg_mobile || '')
          // 运行时覆盖浏览器标签页标题（index.html 的 title 是中性占位）
          document.title = this.systemName
          // 运行时设置浏览器标签图标（与前端对齐，favicon 跟随 config）
          if (favicon) {
            let link = document.querySelector('link[rel="icon"]')
            if (!link) {
              link = document.createElement('link')
              link.rel = 'icon'
              document.head.appendChild(link)
            }
            link.href = favicon
          }
        }
      } catch (e) {
        // 静默失败，使用默认值
      }
    },
    logout() {
      this.userInfo = null
      this.permissions = []
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      localStorage.removeItem('permissions')
    },
  },
})
