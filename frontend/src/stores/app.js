import { defineStore } from 'pinia'
import { getSystemInfo } from '@/api'

export const useAppStore = defineStore('app', {
  state: () => ({
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
    permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
    systemName: localStorage.getItem('systemName') || '企业管理系统',
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
    hasPermission(code) {
      if (this.isAdmin) return true
      return this.permissions.includes(code)
    },
    async fetchSystemInfo() {
      try {
        const res = await getSystemInfo()
        if (res.code === 200 && res.data) {
          this.setSystemName(res.data.name || '企业管理系统')
        }
      } catch (e) {
        // 静默失败，使用默认名称
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
