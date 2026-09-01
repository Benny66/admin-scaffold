<template>
  <div class="login-page">
    <div class="login-body">
      <!-- 左栏：品牌展示区（≥900px 时与表单并排，窄屏时变为顶部品牌带） -->
      <aside class="login-brand">
        <div
          class="brand-bg"
          :class="{ 'brand-bg--visible': bgLoaded }"
          :style="bgStyle"
        ></div>
        <div v-if="bgLoaded" class="brand-scrim"></div>
        <div class="brand-content">
          <img
            v-if="appStore.logoAvailable"
            :src="appStore.logo"
            class="brand-logo"
            alt="logo"
            @error="appStore.markLogoFailed()"
          />
          <div v-else class="brand-logo-text">{{ (systemName || 'B').charAt(0) }}</div>
          <h1 class="brand-name">{{ systemName }}</h1>
          <p v-if="subtitle" class="brand-subtitle">{{ subtitle }}</p>
        </div>
      </aside>

      <!-- 右栏：表单区 -->
      <main class="login-form-area">
        <div class="login-card">
          <h2 class="card-title">登录</h2>
          <p class="card-desc">欢迎使用{{ systemName }}</p>

          <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                placeholder="用户名"
                size="large"
                :prefix-icon="User"
                autocomplete="username"
                autofocus
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="密码"
                size="large"
                :prefix-icon="Lock"
                show-password
                autocomplete="current-password"
                @keyup.enter="handleLogin"
                @keyup="checkCapsLock"
              />
            </el-form-item>

            <p v-if="capsLockOn" class="inline-tip">大写锁定已开启</p>
            <p v-if="errorMsg" class="inline-tip inline-tip--error">{{ errorMsg }}</p>

            <el-button
              type="primary"
              size="large"
              class="login-btn"
              :loading="loading"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form>

          <p v-if="devTip" class="login-tip">{{ devTip }}</p>
        </div>
      </main>
    </div>

    <footer class="login-footer">
      <AppFooter />
    </footer>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api'
import { useAppStore } from '@/stores/app'
import AppFooter from '@/components/AppFooter.vue'

const router = useRouter()
const appStore = useAppStore()

// 默认凭据提示仅开发态可见。写成三元表达式而非 v-if 包字面量，是为了让 Vite 在
// 生产构建时把 import.meta.env.DEV 替换为 false 后常量折叠掉整段，避免明文进入
// dist 产物被安全扫描命中（仅靠 v-if 挡渲染是不够的，字符串仍会留在 bundle 里）。
const devTip = import.meta.env.DEV ? '默认账号：admin / admin123' : ''
const systemName = computed(() => appStore.systemName || 'Base Admin')
const subtitle = computed(() => appStore.subtitle)

const formRef = ref(null)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}
const loading = ref(false)
const errorMsg = ref('')
const capsLockOn = ref(false)

// 背景图：预加载成功后淡入，失败则保持渐变（CSS background-image 无 onerror，故用 Image 探测）
const bgUrl = ref('')
const bgLoaded = ref(false)
const targetBg = computed(() => appStore.loginBg)
const bgStyle = computed(() => (bgLoaded.value ? { backgroundImage: `url(${bgUrl.value})` } : {}))

watch(
  targetBg,
  (url) => {
    bgLoaded.value = false
    bgUrl.value = ''
    if (!url) return
    const img = new Image()
    img.onload = () => {
      bgUrl.value = url
      bgLoaded.value = true
    }
    img.onerror = () => {
      console.warn(`[登录页] 背景图加载失败，已回退渐变：${url}`)
    }
    img.src = url
  },
  { immediate: true }
)

function checkCapsLock(e) {
  capsLockOn.value = !!(e.getModifierState && e.getModifierState('CapsLock'))
}

async function handleLogin() {
  errorMsg.value = ''
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    // skipGlobalError：错误由本页内联展示，豁免全局 toast 与「登录已过期」弹窗
    const res = await login({ username: form.username, password: form.password }, { skipGlobalError: true })
    if (res.code === 200) {
      const { token, user, permissions } = res.data
      localStorage.setItem('token', token)
      appStore.setUserInfo(user)
      appStore.setPermissions(permissions || [])
      ElMessage.success('登录成功')
      router.push('/')
    }
  } catch (e) {
    errorMsg.value = e?.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  appStore.fetchSystemInfo()
})
</script>

<style scoped>
.login-page {
  /* 登录页局部主题 token */
  --brand-from: #1f6feb;
  --brand-to: #0d3b8c;
  --card-width: 380px;

  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.login-body {
  flex: 1;
  display: flex;
}

/* ---------------- 左栏：品牌展示区 ---------------- */
.login-brand {
  position: relative;
  flex: 1.1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  overflow: hidden;
  /* 无背景图时的兜底：渐变（scrim 仅在有图时叠加，避免渐变发灰） */
  background: linear-gradient(135deg, var(--brand-from) 0%, var(--brand-to) 100%);
}

.brand-bg {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  opacity: 0;
  transition: opacity 0.6s ease;
}
.brand-bg--visible {
  opacity: 1;
}

.brand-scrim {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(31, 111, 235, 0.72) 0%, rgba(13, 59, 140, 0.85) 100%);
}

.brand-content {
  position: relative;
  z-index: 1;
  color: #fff;
  text-align: center;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
}

.brand-logo {
  height: 56px;
  width: auto;
  max-width: 220px;
  object-fit: contain;
  margin-bottom: 20px;
}
.brand-logo-text {
  width: 56px;
  height: 56px;
  line-height: 56px;
  margin: 0 auto 20px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.2);
  font-size: 26px;
  font-weight: bold;
}
.brand-name {
  font-size: 30px;
  font-weight: 600;
  margin-bottom: 10px;
}
.brand-subtitle {
  font-size: 15px;
  opacity: 0.9;
}

/* ---------------- 右栏：表单区 ---------------- */
.login-form-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
}
.login-card {
  width: 100%;
  max-width: var(--card-width);
}
.card-title {
  font-size: 24px;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 6px;
}
.card-desc {
  color: #8a919f;
  margin-bottom: 28px;
}
.login-btn {
  width: 100%;
}

.inline-tip {
  margin: -6px 0 12px;
  font-size: 12px;
  color: #e6a23c;
}
.inline-tip--error {
  color: #f56c6c;
}

.login-tip {
  margin-top: 16px;
  text-align: center;
  color: #bbb;
  font-size: 12px;
}

/* ---------------- 页脚 ---------------- */
.login-footer {
  flex-shrink: 0;
  padding: 0 24px;
}

/* ---------------- 窄屏：折叠为单列（品牌带在上、表单在下） ---------------- */
@media (max-width: 899px) {
  .login-body {
    flex-direction: column;
  }
  .login-brand {
    flex: none;
    height: 30vh;
    min-height: 180px;
    padding: 24px;
  }
  .brand-logo {
    height: 44px;
    margin-bottom: 12px;
  }
  .brand-logo-text {
    width: 44px;
    height: 44px;
    line-height: 44px;
    font-size: 20px;
    margin-bottom: 12px;
  }
  .brand-name {
    font-size: 22px;
  }
  .login-form-area {
    padding: 32px 24px;
  }
}
</style>
