import request from '@/utils/request'

// miniapp 端 API 定义（与 frontend/mobile 的 api/index.js 同构）
// 所有请求走 @/utils/request 封装（eslint flat config 禁止直接调 uni.request）

// 微信小程序登录：提交 wx.login 拿到的 code，后端换 openid 后签发 JWT
export const mpLogin = (code) => request.post('/auth/mp-login', { code })

// 系统信息（品牌配置：系统名、logo 等）
export const getSystemInfo = () => request.get('/system/info')
