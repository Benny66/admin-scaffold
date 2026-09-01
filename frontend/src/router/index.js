import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layout/Layout.vue'
import { useAppStore } from '@/stores/app'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: Layout,
    redirect: '/system/user',
    children: [
      {
        path: 'system/user',
        name: 'SystemUser',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '用户管理', icon: 'User', permission: 'users:view' },
      },
      {
        path: 'system/role',
        name: 'SystemRole',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', permission: 'roles:view' },
      },
      {
        path: 'system/permission',
        name: 'SystemPermission',
        component: () => import('@/views/system/permission/index.vue'),
        meta: { title: '权限管理', icon: 'Key', permission: 'permissions:view' },
      },
      {
        path: 'system/dict',
        name: 'SystemDict',
        component: () => import('@/views/system/dict/index.vue'),
        meta: { title: '字典管理', icon: 'Collection', permission: 'dict:view' },
      },
      {
        path: 'system/log',
        name: 'SystemLog',
        component: () => import('@/views/system/log/index.vue'),
        meta: { title: '操作日志', icon: 'Document', permission: 'logs:view' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/ErrorPage.vue'),
    meta: { title: '无权限', code: 403 },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/ErrorPage.vue'),
    meta: { title: '页面不存在', code: 404 },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 系统名称优先取后端配置，路由守卫早于 store 初始化时用安全兜底
  const appStore = useAppStore()
  const systemName = appStore.systemName || '管理系统'
  document.title = (to.meta.title ? to.meta.title + ' - ' : '') + systemName
  const token = localStorage.getItem('token')

  // 白名单：登录页与错误页无条件放行。必须放在权限判断之前，
  // 否则 /403 页面本身会因「无权限」再次被重定向到 /403，形成死循环（见 design D5）。
  if (to.path === '/login' || to.path === '/403') {
    if (to.path === '/login' && token) {
      next('/')
    } else {
      next()
    }
    return
  }

  // 未登录访问受保护路由 → 登录页
  if (!token) {
    next('/login')
    return
  }

  // 已登录访问声明了权限码的路由 → 无权限则进 403
  const required = to.meta && to.meta.permission
  if (required && !appStore.hasPermission(required)) {
    next('/403')
    return
  }

  next()
})

export default router
