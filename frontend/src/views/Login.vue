<template>
  <div class="login-container">
    <div class="login-card">
      <h2 class="login-title">{{ systemName }}</h2>
      <p class="login-subtitle">系统管理基座</p>
      <el-form :model="form" @submit.prevent="handleLogin">
        <el-form-item>
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="handleLogin">
          登录
        </el-button>
      </el-form>
      <p class="login-tip">默认账号：admin / admin123</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const appStore = useAppStore()

const systemName = computed(() => appStore.systemName || 'Base Admin')
const form = ref({ username: '', password: '' })
const loading = ref(false)

onMounted(() => {
  appStore.fetchSystemInfo()
})

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const res = await login(form.value)
    if (res.code === 200) {
      const { token, user, permissions } = res.data
      localStorage.setItem('token', token)
      appStore.setUserInfo(user)
      appStore.setPermissions(permissions || [])
      ElMessage.success('登录成功')
      router.push('/')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f6feb 0%, #0d3b8c 100%);
}
.login-card {
  width: 380px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}
.login-title {
  text-align: center;
  font-size: 24px;
  color: #1f6feb;
  margin-bottom: 8px;
}
.login-subtitle {
  text-align: center;
  color: #999;
  margin-bottom: 28px;
}
.login-btn {
  width: 100%;
}
.login-tip {
  margin-top: 16px;
  text-align: center;
  color: #bbb;
  font-size: 12px;
}
</style>
