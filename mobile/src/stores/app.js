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
    async fetchSystemInfo() {
      try {
        const res = await getSystemInfo()
        if (res.code === 200 && res.data) {
          const { name, logo, footer } = res.data
          this.setSystemName(name || 'Base Admin')
          this.setLogo(logo || '')
          this.setFooter(footer || '')
          // 运行时覆盖浏览器标签页标题（index.html 的 title 是中性占位）
          document.title = this.systemName
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
