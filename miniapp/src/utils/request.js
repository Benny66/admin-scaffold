// miniapp 端 HTTP 请求封装层（基于 uni.request，不引 axios）
//
// 多端铁律落地（miniapp-wechat-end spec 的「HTTP 请求必须走封装层」要求）：
//   - 所有业务请求 MUST 走本封装，eslint flat config 用 no-restricted-syntax
//     抓 CallExpression[callee.object.name='uni'][callee.property.name='request']
//     强制禁止直接调用 uni.request（本文件是唯一豁免）。
//   - 与 frontend/mobile 的 utils/request.js 行为对齐：
//     · 自动附带 Authorization: Bearer <token>
//     · 401 清 token 跳登录页
//     · 403 提示无权限
//     · code === 200 返回 res.data；其他业务码 toast 提示
//
// 不接 refresh 链路（design D10）：等 in-progress 的 auth-token-lifecycle 归档后
// 再开 follow-up change 把 miniapp 接入令牌版本吊销 + 刷新令牌 + 静默续期。

const BASE_URL = (import.meta.env.VITE_API_BASE || 'http://localhost:8080') + '/api'

// isLoggingOut 防止 401 多次并发触发重复跳登录
let isLoggingOut = false

/**
 * 发起请求，返回 res.data（业务码 200）；非 200 reject Error。
 *
 * @param {Object} options 与 uni.request 同形参，外加：
 *   - skipGlobalError: 调用方声明自行呈现错误（如登录页内联展示），跳过 toast 与跳登录
 * @returns {Promise<any>}
 */
function request(options = {}) {
  const {
    url,
    method = 'GET',
    data,
    header = {},
    skipGlobalError = false,
  } = options

  // 自动附带 JWT（从 storage 读，与 stores/app.js 的 setToken 写入位置一致）
  const token = uni.getStorageSync('token')
  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }
  if (!header['Content-Type']) {
    header['Content-Type'] = 'application/json'
  }

  // 注意：本行是 uni.request 的唯一合法调用点（eslint 豁免）
  return new Promise((resolve, reject) => {
    uni.request({
      url: /^https?:\/\//.test(url) ? url : BASE_URL + url,
      method,
      data,
      header,
      success: (res) => {
        // HTTP 401：清 token 跳登录（除非调用方声明自行处理）
        if (res.statusCode === 401) {
          if (skipGlobalError) {
            reject(new Error(res.data?.message || '用户名或密码错误'))
            return
          }
          uni.removeStorageSync('token')
          uni.removeStorageSync('userInfo')
          if (!isLoggingOut) {
            isLoggingOut = true
            uni.showModal({
              title: '提示',
              content: '登录已过期，请重新登录',
              showCancel: false,
              complete: () => {
                isLoggingOut = false
                uni.reLaunch({ url: '/pages/login/index' })
              },
            })
          }
          reject(new Error('登录已过期'))
          return
        }

        // HTTP 403：提示无权限
        if (res.statusCode === 403) {
          if (!skipGlobalError) {
            uni.showToast({ title: '没有权限执行此操作', icon: 'none' })
          }
          reject(new Error('没有权限执行此操作'))
          return
        }

        // 其他 HTTP 错误
        if (res.statusCode < 200 || res.statusCode >= 300) {
          const msg = res.data?.message || `请求错误 (${res.statusCode})`
          if (!skipGlobalError) {
            uni.showToast({ title: msg, icon: 'none' })
          }
          reject(new Error(msg))
          return
        }

        // 业务码 200：返回 res（与 frontend/mobile 对齐，调用方读 .data）
        if (res.data?.code === 200) {
          resolve(res.data)
          return
        }

        // 业务码非 200：toast + reject
        const msg = res.data?.message || '请求失败'
        if (!skipGlobalError) {
          uni.showToast({ title: msg, icon: 'none' })
        }
        reject(new Error(msg))
      },
      fail: () => {
        const msg = '网络连接失败，请检查网络'
        if (!skipGlobalError) {
          uni.showToast({ title: msg, icon: 'none' })
        }
        reject(new Error(msg))
      },
    })
  })
}

// 便捷方法，与 axios 风格对齐（方便业务侧 request.get/post/... 调用）
request.get = (url, options = {}) => request({ ...options, url, method: 'GET' })
request.post = (url, data, options = {}) => request({ ...options, url, method: 'POST', data })
request.put = (url, data, options = {}) => request({ ...options, url, method: 'PUT', data })
request.delete = (url, options = {}) => request({ ...options, url, method: 'DELETE' })

export default request
