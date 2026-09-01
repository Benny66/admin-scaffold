import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import App from './App.vue'
import router from './router'
import permission from './directives/permission'
import './assets/styles/global.css'

const app = createApp(App)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const pinia = createPinia()
app.use(pinia)
// 注册 v-permission 指令：必须在 pinia 之后，否则指令内 useAppStore() 取不到实例
app.directive('permission', permission)
app.use(router)
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')
