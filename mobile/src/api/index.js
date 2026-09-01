import request from '@/utils/request'

// 认证
// config 可传 { skipGlobalError: true }：错误由调用方自行呈现，豁免全局 toast 与
// 「登录已过期」弹窗，避免登录页同一条消息既弹 toast 又显示在表单内联错误区
export const login = (data, config = {}) => request.post('/auth/login', data, config)
export const logout = () => request.post('/auth/logout')
export const getUserInfo = () => request.get('/auth/info')
export const changePassword = (data) => request.put('/auth/password', data)

// 系统信息
export const getSystemInfo = () => request.get('/system/info')

// 字典
export const getDictItemsByCode = (code) => request.get(`/dict/code/${code}`)
