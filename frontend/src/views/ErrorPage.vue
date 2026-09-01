<template>
  <div class="error-page">
    <div class="error-code">{{ code }}</div>
    <div class="error-title">{{ title }}</div>
    <div class="error-desc">{{ desc }}</div>
    <el-button type="primary" @click="goHome">返回首页</el-button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const code = computed(() => router.currentRoute.value.meta.code || 404)
const isForbidden = computed(() => code.value === 403)
const title = computed(() => (isForbidden.value ? '没有访问权限' : '页面不存在'))
const desc = computed(() =>
  isForbidden.value
    ? '您没有权限访问该页面，请联系管理员分配相应权限'
    : '您访问的页面不存在或已被移除'
)

function goHome() {
  router.push('/')
}
</script>

<style scoped>
.error-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}
.error-code {
  font-size: 72px;
  font-weight: bold;
  color: #909399;
  line-height: 1;
}
.error-title {
  font-size: 20px;
  color: #303133;
}
.error-desc {
  color: #909399;
  margin-bottom: 12px;
}
</style>
