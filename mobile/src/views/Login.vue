<template>
  <div class="login-page">
    <!-- 上部：品牌区（背景图 + scrim + 品牌块） -->
    <header class="login-brand">
      <div class="brand-bg" :class="{ 'brand-bg--visible': bgLoaded }" :style="bgStyle"></div>
      <div v-if="bgLoaded" class="brand-scrim"></div>
      <div class="brand-content">
        <van-image
          v-if="appStore.logoAvailable"
          :src="appStore.logo"
          width="52"
          height="52"
          fit="contain"
          class="brand-logo"
          @error="appStore.markLogoFailed()"
        />
        <div v-else class="brand-logo-text">{{ (systemName || 'B').charAt(0) }}</div>
        <h1 class="brand-name">{{ systemName }}</h1>
        <p v-if="subtitle" class="brand-subtitle">{{ subtitle }}</p>
      </div>
    </header>

    <!-- 下部：表单区（纯色背景，不浮于背景图之上） -->
    <main class="login-form-area">
      <van-form @submit="handleLogin">
        <van-cell-group inset>
          <van-field
            v-model="form.username"
            name="username"
            label="用户名"
            placeholder="请输入用户名"
            autocomplete="username"
            :rules="[{ required: true, message: '请输入用户名' }]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            autocomplete="current-password"
            :rules="[{ required: true, message: '请输入密码' }]"
          />
        </van-cell-group>
        <p v-if="errorMsg" class="inline-tip">{{ errorMsg }}</p>
        <div class="login-btn">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            登录
          </van-button>
        </div>
      </van-form>

      <p v-if="devTip" class="login-tip">{{ devTip }}</p>
    </main>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { login } from '@/api'
import { useAppStore } from '@/stores/app'
import AppFooter from '@/components/AppFooter.vue'

const router = useRouter()
const appStore = useAppStore()

// 默认凭据提示仅开发态可见。写成三元表达式（而非 v-if 包字面量），让生产构建在
// import.meta.env.DEV 替换为 false 后常量折叠掉整段，避免明文进入 dist 产物。
const devTip = import.meta.env.DEV ? '默认账号：admin / admin123' : ''
const systemName = computed(() => appStore.systemName || 'Base Admin')
const subtitle = computed(() => appStore.subtitle)

const form = reactive({ username: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

// 背景图：移动端优先 login_bg_mobile，未配则回退 login_bg，都无则回退渐变（D2 回退链）
const bgUrl = ref('')
const bgLoaded = ref(false)
const targetBg = computed(() => appStore.loginBgMobile || appStore.loginBg)
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

async function handleLogin() {
  errorMsg.value = ''
  loading.value = true
  try {
    // skipGlobalError：错误由本页内联展示，豁免全局 toast 与「登录已过期」弹窗
    const res = await login({ username: form.username, password: form.password }, { skipGlobalError: true })
    if (res.code === 200) {
      const { token, user, permissions } = res.data
      localStorage.setItem('token', token)
      appStore.setUserInfo(user)
      appStore.setPermissions(permissions || [])
      showToast('登录成功')
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
  /* 登录页局部尺寸 token。
     --brand-from / --brand-to 由 brand-color-extract 注入 :root（取自 logo 主色），
     此处不再声明同名变量（会覆盖 :root 的值），只在渐变处用 var() fallback 兜底。 */
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}

/* ---------------- 上部：品牌区 ---------------- */
.login-brand {
  position: relative;
  flex: none;
  height: 32vh;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  overflow: hidden;
  /* 无背景图时的兜底：渐变（scrim 仅在有图时叠加）。
     色源取注入到 :root 的品牌变量，未注入时回退移动端现有的 #1989fa → #0d3b8c。 */
  background: linear-gradient(
    135deg,
    var(--brand-from, #1989fa) 0%,
    var(--brand-to, #0d3b8c) 100%
  );
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
  /* scrim 色由 deriveThemeVars 预派生后整体注入，此处只负责取用（rgba 不接受变量分量） */
  background: linear-gradient(
    135deg,
    var(--brand-scrim-from, rgba(25, 137, 250, 0.7)) 0%,
    var(--brand-scrim-to, rgba(13, 59, 140, 0.85)) 100%
  );
}

.brand-content {
  position: relative;
  z-index: 1;
  color: #fff;
  text-align: center;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
}
.brand-logo {
  margin-bottom: 12px;
}
.brand-logo-text {
  width: 52px;
  height: 52px;
  line-height: 52px;
  margin: 0 auto 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.2);
  font-size: 24px;
  font-weight: bold;
}
.brand-name {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 6px;
}
.brand-subtitle {
  font-size: 14px;
  opacity: 0.9;
}

/* ---------------- 下部：表单区 ---------------- */
.login-form-area {
  flex: 1;
  padding-top: 24px;
}
.login-btn {
  margin: 24px 16px 0;
}

.inline-tip {
  margin: 12px 16px 0;
  font-size: 12px;
  color: #ee0a24;
}

.login-tip {
  margin-top: 20px;
  text-align: center;
  color: #c8c9cc;
  font-size: 12px;
}
</style>
