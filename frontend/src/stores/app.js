import { defineStore } from 'pinia'
import { getSystemInfo } from '@/api'

export const useAppStore = defineStore('app', {
  state: () => ({
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
    permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
    systemName: localStorage.getItem('systemName') || '',
    logo: localStorage.getItem('logo') || '',
    footer: localStorage.getItem('footer') || '',
  }),
  getters: {
    isLoggedIn: (state) => !!state.userInfo && !!localStorage.getItem('token'),
    isAdmin: (state) => !!state.userInfo?.is_admin,
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
    setLogo(logo) {
      this.logo = logo
      localStorage.setItem('logo', logo)
    },
    setFooter(footer) {
      this.footer = footer
      localStorage.setItem('footer', footer)
    },
    hasPermission(code) {
      if (this.isAdmin) return true
      return this.permissions.includes(code)
    },
    async fetchSystemInfo() {
      try {
        const res = await getSystemInfo()
        if (res.code === 200 && res.data) {
          const { name, logo, favicon, footer } = res.data
          this.setSystemName(name || 'Base Admin')
          this.setLogo(logo || '')
          this.setFooter(footer || '')
          // 运行时动态设置浏览器标签图标（favicon 跟随 config，而非编译期写死）
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
