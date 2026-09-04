import request from '@/utils/request'

// ==================== 认证 ====================
// config 可传 { skipGlobalError: true }：错误由调用方自行呈现，豁免全局 toast 与
// 「登录已过期」弹窗，避免登录页同一条消息既弹 toast 又显示在表单内联错误区
export const login = (data, config = {}) => request.post('/auth/login', data, config)
export const logout = () => request.post('/auth/logout')
export const getUserInfo = () => request.get('/auth/info')
export const changePassword = (data) => request.put('/auth/password', data)

// ==================== 用户管理 ====================
export const getUserList = (params) => request.get('/users', { params })
export const getUser = (id) => request.get(`/users/${id}`)
export const createUser = (data) => request.post('/users', data)
export const updateUser = (id, data) => request.put(`/users/${id}`, data)
export const deleteUser = (id) => request.delete(`/users/${id}`)
export const resetPassword = (id, data) => request.put(`/users/${id}/password`, data)
export const toggleUserStatus = (id, data) => request.put(`/users/${id}/status`, data)

// ==================== 角色管理 ====================
export const getRoleList = (params) => request.get('/roles', { params })
export const getAllRoles = () => request.get('/roles/all')
export const getRole = (id) => request.get(`/roles/${id}`)
export const createRole = (data) => request.post('/roles', data)
export const updateRole = (id, data) => request.put(`/roles/${id}`, data)
export const deleteRole = (id) => request.delete(`/roles/${id}`)

// ==================== 权限管理 ====================
export const getPermissionList = (params) => request.get('/permissions', { params })
export const getAllPermissions = () => request.get('/permissions/all')
export const getPermission = (id) => request.get(`/permissions/${id}`)
export const createPermission = (data) => request.post('/permissions', data)
export const updatePermission = (id, data) => request.put(`/permissions/${id}`, data)
export const deletePermission = (id) => request.delete(`/permissions/${id}`)
export const getRolePermissions = (roleID) => request.get(`/permissions/role/${roleID}`)
export const assignPermissionsToRole = (roleID, data) => request.post(`/permissions/role/${roleID}/assign`, data)

// ==================== 字典管理 ====================
export const getDictTypeList = (params) => request.get('/dict/types', { params })
export const getDictType = (id) => request.get(`/dict/types/${id}`)
export const createDictType = (data) => request.post('/dict/types', data)
export const updateDictType = (id, data) => request.put(`/dict/types/${id}`, data)
export const deleteDictType = (id) => request.delete(`/dict/types/${id}`)
export const getDictItemList = (params) => request.get('/dict/items', { params })
export const createDictItem = (data) => request.post('/dict/items', data)
export const updateDictItem = (id, data) => request.put(`/dict/items/${id}`, data)
export const deleteDictItem = (id) => request.delete(`/dict/items/${id}`)
export const getDictItemsByCode = (code) => request.get(`/dict/code/${code}`)
// 单类型导出 / 下载导入模板（响应为 blob 文件）
export const exportDictItems = (id, params) => request.get(`/dict/types/${id}/items/export`, { params, responseType: 'blob' })
// 全量导出（多 sheet，响应为 blob 文件）
export const exportAllDictTypes = (params) => request.get('/dict/types/export', { params, responseType: 'blob' })
// 上传 xlsx 按 value 覆盖合并导入
export const importDictItems = (id, formData) => request.post(`/dict/types/${id}/items/import`, formData)

// ==================== 日志管理 ====================
export const getOperationLogList = (params) => request.get('/logs/operation', { params })
export const getLoginLogList = (params) => request.get('/logs/login', { params })
export const clearOperationLog = () => request.delete('/logs/operation')

// ==================== 系统信息 ====================
export const getSystemInfo = () => request.get('/system/info')
