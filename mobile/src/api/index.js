import request from '@/utils/request'

// 认证
export const login = (data) => request.post('/auth/login', data)
export const logout = () => request.post('/auth/logout')
export const getUserInfo = () => request.get('/auth/info')
export const changePassword = (data) => request.put('/auth/password', data)

// 系统信息
export const getSystemInfo = () => request.get('/system/info')

// 字典
export const getDictItemsByCode = (code) => request.get(`/dict/code/${code}`)
