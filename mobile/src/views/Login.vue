<template>
  <div class="login-page">
    <div class="login-header">
      <h1>{{ systemName }}</h1>
      <p>移动端</p>
    </div>
    <van-form @submit="handleLogin">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="用户名"
          placeholder="请输入用户名"
          :rules="[{ required: true, message: '请输入用户名' }]"
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="请输入密码"
          :rules="[{ required: true, message: '请输入密码' }]"
        />
      </van-cell-group>
      <div class="login-btn">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          登录
        </van-button>
      </div>
    </van-form>
    <p class="login-tip">默认账号：admin / admin123</p>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { login } from '@/api'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const appStore = useAppStore()

const systemName = computed(() => appStore.systemName || 'Base Admin')
const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const res = await login(form)
    if (res.code === 200) {
      const { token, user, permissions } = res.data
      localStorage.setItem('token', token)
      appStore.setUserInfo(user)
      appStore.setPermissions(permissions || [])
      showToast('登录成功')
      router.push('/')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100%;
  background: linear-gradient(135deg, #1989fa 0%, #0d3b8c 100%);
  padding-top: 80px;
}
.login-header {
  text-align: center;
  color: #fff;
  margin-bottom: 40px;
}
.login-header h1 {
  font-size: 26px;
  margin-bottom: 8px;
}
.login-header p {
  opacity: 0.8;
}
.login-btn {
  margin: 24px 16px;
}
.login-tip {
  text-align: center;
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
}
</style>
