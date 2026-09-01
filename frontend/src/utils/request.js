import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import router from '@/router'

// 创建 axios 实例
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
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 统一处理响应和错误
//
// skipGlobalError：调用方可在请求 config 上带此标记，声明「本请求的错误由我自己
// 呈现」。拦截器据此跳过一切全局错误呈现（toast、「登录已过期」弹窗、清 token），
// 只 reject 一个携带 message 的 Error 交给调用方。
// 典型用途是登录页：错误要在表单内联展示，否则同一条消息会既弹 toast 又显示在表单里。
let is401DialogShowing = false
request.interceptors.response.use(
  (response) => {
    // blob 类型直接返回，不做 code 检查
    if (response.config.responseType === 'blob') {
      return response
    }
    const res = response.data
    if (res.code === 200) {
      return res
    }
    // 业务错误
    const msg = res.message || '请求失败'
    if (!response.config?.skipGlobalError) {
      ElMessage.error(msg)
    }
    return Promise.reject(new Error(msg))
  },
  (error) => {
    // 调用方声明自行呈现错误时跳过全局呈现（见文件头 skipGlobalError 说明）
    const skip = !!error.config?.skipGlobalError

    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        // 登录页豁免：「密码错误」与「会话过期」是两个语义，全局拦截器无法区分，
        // 只能由调用方声明。登录页在表单内联展示，既不弹「登录已过期」也不弹 toast。
        if (skip) {
          return Promise.reject(new Error(data?.message || '用户名或密码错误'))
        }
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
        localStorage.removeItem('permissions')
        if (!is401DialogShowing) {
          is401DialogShowing = true
          ElMessageBox.confirm('登录已过期，请重新登录', '提示', {
            confirmButtonText: '重新登录',
            cancelButtonText: '取消',
            type: 'warning',
          })
            .then(() => {
              router.push('/login')
            })
            .catch(() => {
              router.push('/login')
            })
            .finally(() => {
              is401DialogShowing = false
            })
        }
        return Promise.reject(error)
      }

      const msg = status === 403 ? '没有权限执行此操作' : data?.message || `请求错误 (${status})`
      if (!skip) {
        ElMessage.error(msg)
      }
      return Promise.reject(skip ? new Error(msg) : error)
    }

    const msg = error.request ? '网络连接失败，请检查网络' : '请求配置错误'
    if (!skip) {
      ElMessage.error(msg)
    }
    return Promise.reject(skip ? new Error(msg) : error)
  }
)

export default request
