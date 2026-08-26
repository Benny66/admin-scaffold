import axios from 'axios'
import { showToast, showConfirmDialog } from 'vant'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 自动附加 JWT 令牌
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器 - 统一处理响应和错误
let is401Showing = false
request.interceptors.response.use(
  (response) => {
    if (response.config.responseType === 'blob') {
      return response
    }
    const res = response.data
    if (res.code === 200) {
      return res
    }
    showToast(res.message || '请求失败')
    return Promise.reject(new Error(res.message || '请求失败'))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
        if (!is401Showing) {
          is401Showing = true
          showConfirmDialog({
            title: '提示',
            message: '登录已过期，请重新登录',
          })
            .then(() => router.push('/login'))
            .catch(() => router.push('/login'))
            .finally(() => {
              is401Showing = false
            })
        }
        return Promise.reject(error)
      }
      if (status === 403) {
        showToast('没有权限执行此操作')
        return Promise.reject(error)
      }
      showToast(data?.message || `请求错误 (${status})`)
    } else if (error.request) {
      showToast('网络连接失败，请检查网络')
    } else {
      showToast('请求配置错误')
    }
    return Promise.reject(error)
  }
)

export default request
