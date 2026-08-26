<template>
  <div class="home-page">
    <van-nav-bar title="企业管理系统" />

    <div class="welcome">
      <van-image round width="60" height="60" src="https://img.yzcdn.cn/vant/cat.jpeg" />
      <div class="welcome-text">
        <div class="name">{{ userInfo?.real_name || userInfo?.username || '未登录' }}</div>
        <div class="desc">欢迎使用企业管理系统移动端</div>
      </div>
    </div>

    <van-cell-group inset title="常用功能">
      <van-cell title="我的信息" icon="user-o" is-link to="/mine" />
      <van-cell title="关于系统" icon="info-o" :value="'v1.0.0'" />
    </van-cell-group>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '@/store/app'

const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)
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
