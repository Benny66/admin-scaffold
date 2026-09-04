import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
  },
  {
    path: '/mine',
    name: 'Mine',
    component: () => import('@/views/Mine.vue'),
  },
]

const router = createRouter({
  // history base 取自 vite base（import.meta.env.BASE_URL）：
  //   build 时 = '/m/'（与产物静态资源前缀一致，后端在 /m/ 子路径托管）
  //   dev    时 = '/'（直接访问 http://localhost:5174/）
  // base 是 vite.config.js 里的唯一配置点，此处禁止手写死值以免两处漂移。
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path === '/login') {
    if (token) {
      next('/')
    } else {
      next()
    }
  } else {
    if (!token) {
      next('/login')
    } else {
      next()
    }
  }
})

export default router
