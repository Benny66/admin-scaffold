<template>
  <div v-if="text" class="app-footer">{{ text }}</div>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
// footer 为空时不渲染占位；内容按纯文本展示，不解析 HTML（避免配置文件成为注入面）
const text = computed(() => appStore.footer)
</script>

<style scoped>
.app-footer {
  padding: 16px;
  /* 底部安全区：避免 iPhone 系统横条遮挡页脚 */
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
  text-align: center;
  color: #999;
  font-size: 12px;
}
</style>
