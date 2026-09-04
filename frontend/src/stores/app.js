import { defineStore } from 'pinia'
import { getSystemInfo } from '@/api'
import {
  applyThemeVars,
  clearThemeVars,
  deriveThemeVars,
  extractDominantColor,
} from '@/utils/colorExtract'

// BRAND_CACHE_VERSION 缓存结构版本。deriveThemeVars 的输出结构或算法变化时 +1，
// 让老缓存整体失效——design D5 只覆盖了「换 logo 自动失效」，管不到基座升级后
// 旧缓存把旧算法算出的色值钉死在本地。
const BRAND_CACHE_VERSION = 'v1'

// brandCacheKey 把 logo URL 编进缓存 key：换 logo 文件名即自动失效重算（design D5）。
function brandCacheKey(logoUrl) {
  return `brand_theme_${BRAND_CACHE_VERSION}_${logoUrl}`
}

// readBrandCache / writeBrandCache：无痕模式下 localStorage 读写会直接抛错，
// 故统一 try-catch 静默降级——缓存不可用只影响「少算一次」，不影响主题注入。
function readBrandCache(key) {
  try {
    const raw = localStorage.getItem(key)
    return raw ? JSON.parse(raw) : null
  } catch (e) {
    return null
  }
}

function writeBrandCache(key, vars) {
  try {
    localStorage.setItem(key, JSON.stringify(vars))
  } catch (e) {
    // 配额满或存储被禁用：放弃缓存，下次重新提取
  }
}

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
    // logo 加载失败标记（如 config 配了但 backend/static 下文件不存在）。
    // 刻意不持久化：补上文件后刷新即可恢复，无需清缓存。
    logoFailed: false,
  }),
  getters: {
    isLoggedIn: (state) => !!state.userInfo && !!localStorage.getItem('token'),
    isAdmin: (state) => !!state.userInfo?.is_admin,
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
    hasPermission(code) {
      if (this.isAdmin) return true
      return this.permissions.includes(code)
    },
    // applyBrandTheme 从 logo 提取主色并注入全站主题变量（brand-color-extract spec）。
    // 任一环节失败（无 logo / 加载失败 / 纯灰或透明）都回退默认蓝色，
    // 保证基座零配置开箱即用、视觉与现状一致。
    async applyBrandTheme(logoUrl) {
      // 整条链路兜底：本 action 被 fetchSystemInfo 以「即发即忘」方式调用（不 await，
      // 免得图片加载拖慢品牌信息渲染），故必须自己吞掉所有异常——否则会产生
      // unhandled rejection，且主题变量停在半注入状态。
      try {
        // 回退 = 不注入：清掉可能残留的变量，交回各端 CSS 的 var() fallback。
        // 不注入默认蓝色是因为两端渐变起点本就不同色，统一注入会让移动端变色。
        if (!logoUrl) {
          clearThemeVars()
          return
        }

        // 命中缓存：直接注入，不重复做 Canvas 提取
        const cacheKey = brandCacheKey(logoUrl)
        const cached = readBrandCache(cacheKey)
        if (cached) {
          applyThemeVars(cached)
          return
        }

        const primary = await extractDominantColor(logoUrl)
        // 提取失败（纯灰 / 全透明 / 加载失败）与无 logo 同路径回退
        if (!primary) {
          clearThemeVars()
          return
        }

        const vars = deriveThemeVars(primary)
        applyThemeVars(vars)
        writeBrandCache(cacheKey, vars)
      } catch (e) {
        clearThemeVars()
      }
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
          // 品牌色派生必须在拿到 logo 之后：无 logo / 提取失败都回退默认蓝色
          this.applyBrandTheme(logo || '')
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
