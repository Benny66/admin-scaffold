<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'

onLaunch(() => {
  // 应用启动：可在此处做全局初始化（如读 token 决定入口页）
})

onShow(() => {
  // 应用从后台切到前台
})

onHide(() => {
  // 应用切到后台
})
</script>

<style>
/* 全局样式 */
page {
  background-color: #f8f8f8;
}
</style>
