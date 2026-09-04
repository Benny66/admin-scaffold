<template>
  <view class="login-page">
    <view class="login-brand">
      <image
        v-if="appStore.logoAvailable"
        :src="appStore.logo"
        class="brand-logo"
        @error="appStore.markLogoFailed()"
      />
      <view v-else class="brand-logo-text">{{ (appStore.systemName || 'B').charAt(0) }}</view>
      <text class="brand-name">{{ appStore.systemName || 'Base Admin' }}</text>
    </view>

    <view class="login-action">
      <view v-if="loading" class="loading-tip">登录中...</view>
      <view v-else-if="errorMsg" class="error-tip">{{ errorMsg }}</view>
      <button class="login-btn" type="primary" :loading="loading" @click="handleLogin">
        微信一键登录
      </button>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { mpLogin } from '@/api'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const loading = ref(false)
const errorMsg = ref('')

// 小程序登录流程（wechat-mp-login spec）：
//   uni.login 拿 code → POST /api/auth/mp-login → 后端换 openid → 签发 JWT
//   → 存 token + userInfo + permissions → 跳首页
async function handleLogin() {
  errorMsg.value = ''
  loading.value = true
  try {
    // 1. 调 uni.login 拿微信登录 code
    const loginRes = await new Promise((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: resolve,
        fail: reject,
      })
    })

    if (!loginRes.code) {
      errorMsg.value = '获取微信登录凭证失败'
      return
    }

    // 2. 调后端 mp-login 接口（skipGlobalError：错误由本页内联展示）
    const res = await mpLogin(loginRes.code)
    if (res.code === 200 && res.data) {
      const { token, user, permissions } = res.data
      appStore.setToken(token)
      appStore.setUserInfo(user)
      appStore.setPermissions(permissions || [])
      uni.reLaunch({ url: '/pages/index/index' })
    }
  } catch (e) {
    errorMsg.value = e?.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

// 页面加载时自动触发登录（小程序首屏即登录是常见交互）
onLoad(() => {
  handleLogin()
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1989fa 0%, #0d3b8c 100%);
  padding: 40rpx;
}

.login-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 80rpx;
}

.brand-logo {
  width: 120rpx;
  height: 120rpx;
  margin-bottom: 24rpx;
  border-radius: 24rpx;
}

.brand-logo-text {
  width: 120rpx;
  height: 120rpx;
  line-height: 120rpx;
  text-align: center;
  margin-bottom: 24rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.2);
  font-size: 48rpx;
  font-weight: bold;
  color: #fff;
}

.brand-name {
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
}

.login-action {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.loading-tip {
  color: #fff;
  font-size: 28rpx;
  margin-bottom: 24rpx;
}

.error-tip {
  color: #ffebee;
  font-size: 28rpx;
  margin-bottom: 24rpx;
  text-align: center;
}

.login-btn {
  width: 80%;
  border-radius: 48rpx;
}
</style>
