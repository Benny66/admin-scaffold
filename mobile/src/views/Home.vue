<template>
  <div class="home-page">
    <van-nav-bar :title="systemName">
      <template #left>
        <van-image v-if="appStore.logo" :src="appStore.logo" width="28" height="28" fit="contain" />
      </template>
    </van-nav-bar>

    <div class="welcome">
      <van-image
        v-if="appStore.logoAvailable"
        round
        width="60"
        height="60"
        :src="appStore.logo"
        @error="appStore.markLogoFailed()"
      />
      <div v-else class="avatar-fallback">
        {{ (userInfo?.real_name || userInfo?.username || 'U').charAt(0) }}
      </div>
      <div class="welcome-text">
        <div class="name">{{ userInfo?.real_name || userInfo?.username || '未登录' }}</div>
        <div class="desc">欢迎使用{{ systemName }}移动端</div>
      </div>
    </div>

    <van-cell-group inset title="常用功能">
      <van-cell title="我的信息" icon="user-o" is-link to="/mine" />
      <van-cell title="关于系统" icon="info-o" :value="'v1.0.0'" />
    </van-cell-group>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)
const systemName = computed(() => appStore.systemName || 'Base Admin')

onMounted(() => {
  appStore.fetchSystemInfo()
})
</script>

<style scoped>
.home-page {
  min-height: 100%;
}
.welcome {
  display: flex;
  align-items: center;
  padding: 24px 16px;
  background: #fff;
  margin-bottom: 16px;
}
.avatar-fallback {
  width: 60px;
  height: 60px;
  line-height: 60px;
  border-radius: 50%;
  background: #1989fa;
  color: #fff;
  font-size: 24px;
  text-align: center;
}
.welcome-text {
  margin-left: 16px;
}
.name {
  font-size: 18px;
  font-weight: bold;
}
.desc {
  color: #999;
  font-size: 13px;
  margin-top: 4px;
}
</style>
