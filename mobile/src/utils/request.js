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
//
// skipGlobalError：调用方可在请求 config 上带此标记，声明「本请求的错误由我自己
// 呈现」。拦截器据此跳过一切全局错误呈现（toast、「登录已过期」弹窗、清 token），
// 只 reject 一个携带 message 的 Error 交给调用方。
// 典型用途是登录页：错误要在表单内联展示，否则同一条消息会既弹 toast 又显示在表单里。
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
    const msg = res.message || '请求失败'
    if (!response.config?.skipGlobalError) {
      showToast(msg)
    }
    return Promise.reject(new Error(msg))
  },
  (error) => {
    // 调用方声明自行呈现错误时跳过全局呈现（见文件头 skipGlobalError 说明）
    const skip = !!error.config?.skipGlobalError

    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        // 登录页豁免（与前端同构）：「密码错误」≠「会话过期」，由调用方内联展示
        if (skip) {
          return Promise.reject(new Error(data?.message || '用户名或密码错误'))
        }
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

      const msg = status === 403 ? '没有权限执行此操作' : data?.message || `请求错误 (${status})`
      if (!skip) {
        showToast(msg)
      }
      return Promise.reject(skip ? new Error(msg) : error)
    }

    const msg = error.request ? '网络连接失败，请检查网络' : '请求配置错误'
    if (!skip) {
      showToast(msg)
    }
    return Promise.reject(skip ? new Error(msg) : error)
  }
)

export default request
