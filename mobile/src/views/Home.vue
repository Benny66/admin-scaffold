<template>
  <div class="home-page">
    <van-nav-bar :title="systemName">
      <template #left>
        <van-image v-if="appStore.logo" :src="appStore.logo" width="28" height="28" fit="contain" />
      </template>
    </van-nav-bar>

    <div class="welcome">
      <van-image round width="60" height="60" src="https://img.yzcdn.cn/vant/cat.jpeg" />
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
