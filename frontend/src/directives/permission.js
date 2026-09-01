import { useAppStore } from '@/stores/app'

// v-permission 指令：用户无对应权限时，从 DOM 中移除该元素（见 design D4）。
// 移除而非 v-show 隐藏 / disabled 禁用：隐藏的按钮仍可被 Tab 聚焦、仍占布局空间，
// 禁用的按钮用户无法理解为什么点不了。
//
// 用法：
//   v-permission="'users:create'"                 单个权限码
//   v-permission="['users:edit', 'users:create']"  任一满足即可
export default {
  mounted(el, binding) {
    const appStore = useAppStore()
    const value = binding.value
    const codes = Array.isArray(value) ? value : [value]
    const allowed = codes.some((code) => appStore.hasPermission(code))
    if (!allowed && el.parentNode) {
      el.parentNode.removeChild(el)
    }
  },
}
