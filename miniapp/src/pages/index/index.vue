<template>
  <view class="index-page">
    <view class="index-header">
      <image
        v-if="appStore.logoAvailable"
        :src="appStore.logo"
        class="header-logo"
        @error="appStore.markLogoFailed()"
      />
      <text class="header-title">{{ appStore.systemName || 'Base Admin' }}</text>
    </view>

    <view class="index-body">
      <view class="welcome-card">
        <text class="welcome-text">欢迎来到{{ appStore.systemName || 'Base Admin' }}</text>
        <text class="welcome-sub" v-if="appStore.isLoggedIn">
          当前用户：{{ appStore.userInfo?.real_name || appStore.userInfo?.username }}
        </text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

// onShow 调 fetchSystemInfo 拉取系统名与 logo（brand-config spec）
// 每次页面显示都刷新，保证品牌信息实时
onShow(() => {
  appStore.fetchSystemInfo()
})
</script>

<style scoped>
.index-page {
  min-height: 100vh;
  background-color: #f8f8f8;
}

.index-header {
  display: flex;
  align-items: center;
  padding: 24rpx 32rpx;
  background-color: #fff;
  border-bottom: 1rpx solid #e8e8e8;
}

.header-logo {
  width: 64rpx;
  height: 64rpx;
  margin-right: 16rpx;
  border-radius: 12rpx;
}

.header-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #333;
}

.index-body {
  padding: 32rpx;
}

.welcome-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 48rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.welcome-text {
  font-size: 36rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 16rpx;
}

.welcome-sub {
  font-size: 28rpx;
  color: #999;
}
</style>
