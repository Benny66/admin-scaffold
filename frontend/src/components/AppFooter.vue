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
  margin-top: 24px;
  padding-top: 16px;
  text-align: center;
  color: #999;
  font-size: 13px;
  border-top: 1px solid #e4e7ed;
}
</style>
