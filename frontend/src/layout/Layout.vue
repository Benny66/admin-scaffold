<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="aside">
      <div class="logo">
        <img
          v-if="appStore.logoAvailable"
          :src="appStore.logo"
          class="logo-img"
          alt="logo"
          @error="appStore.markLogoFailed()"
        />
        <span v-else-if="!isCollapse">{{ systemName }}</span>
        <span v-else>{{ (systemName || 'B').charAt(0) }}</span>
      </div>
      <el-menu
        :default-active="$route.path"
        :collapse="isCollapse"
        router
        background-color="#001529"
        text-color="#a6adb4"
        active-text-color="#ffffff"
        class="menu"
      >
        <el-sub-menu
          v-for="group in menus"
          :key="group.path"
          :index="group.path"
        >
          <template #title>
            <el-icon><component :is="group.icon" /></el-icon>
            <span>{{ group.title }}</span>
          </template>
          <el-menu-item
            v-for="item in group.children"
            :key="item.path"
            :index="item.path"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
      <!-- 零可见菜单时的空态兜底，避免侧边栏一片空白（见 design D9） -->
      <el-empty
        v-if="menus.length === 0"
        :image-size="60"
        description="暂无可见菜单"
        class="menu-empty"
      />
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <span class="page-title">{{ $route.meta.title || '' }}</span>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="28" style="margin-right: 8px">
                {{ (userInfo?.real_name || userInfo?.username || 'U').charAt(0) }}
              </el-avatar>
              {{ userInfo?.real_name || userInfo?.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main">
        <router-view />
        <AppFooter />
      </el-main>
    </el-container>

    <!-- 修改密码弹窗 -->
    <el-dialog v-model="passwordDialog" title="修改密码" width="420px">
      <el-form :model="passwordForm" label-width="80px">
        <el-form-item label="原密码">
          <el-input v-model="passwordForm.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialog = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAppStore } from '@/stores/app'
import { changePassword } from '@/api'
import AppFooter from '@/components/AppFooter.vue'

const router = useRouter()
const appStore = useAppStore()

const isCollapse = ref(false)
const userInfo = computed(() => appStore.userInfo)
const systemName = computed(() => appStore.systemName || 'Base Admin')

// 菜单从路由声明派生（单一数据源），按两级分组渲染（见 menu-grouping design D1/D3）。
// 派生顺序：先按 meta.permission 过滤每组叶子（缺省可见、isAdmin 直通），再丢弃叶子数为 0
// 的分组——空分组不留空壳标题。新增业务模块只需在 router/index.js 注册分组+叶子路由。
const menus = computed(() => {
  const root = router.options.routes.find((r) => r.path === '/')
  const groups = (root && root.children) || []
  const result = []
  for (const group of groups) {
    // 只渲染分组（有 children 的节点）；裸叶子由 ESLint 规则禁止挂顶层
    if (!Array.isArray(group.children) || group.children.length === 0) continue
    const leaves = group.children
      .filter((item) => {
        const code = item.meta && item.meta.permission
        return !code || appStore.hasPermission(code)
      })
      .map((item) => ({
        // 父 path + 子 path 拼出完整 URL（如 system + user = /system/user）
        path: '/' + [group.path, (item.path || '').replace(/^\//, '')]
          .filter(Boolean)
          .join('/'),
        title: item.meta && item.meta.title,
        icon: item.meta && item.meta.icon,
      }))
    if (leaves.length === 0) continue // 空分组整组丢弃
    result.push({
      path: '/' + (group.path || '').replace(/^\//, ''),
      title: group.meta && group.meta.title,
      icon: group.meta && group.meta.icon,
      children: leaves,
    })
  }
  return result
})

const passwordDialog = ref(false)
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

function handleCommand(command) {
  if (command === 'logout') {
    handleLogout()
  } else if (command === 'password') {
    passwordDialog.value = true
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  }
}

function handleLogout() {
  appStore.logout()
  router.push('/login')
}

async function handleChangePassword() {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }
  const res = await changePassword({
    old_password: passwordForm.value.old_password,
    new_password: passwordForm.value.new_password,
  })
  if (res.code === 200) {
    ElMessage.success('密码修改成功，请重新登录')
    passwordDialog.value = false
    handleLogout()
  }
}

onMounted(() => {
  appStore.fetchSystemInfo()
})
</script>

<style scoped>
.layout-container {
  height: 100%;
}
.aside {
  background-color: #001529;
  transition: width 0.3s;
  overflow: hidden;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  white-space: nowrap;
}
.logo-img {
  height: 36px;
  width: auto;
  max-width: 160px;
  object-fit: contain;
}
.menu {
  border-right: none;
}
.menu-empty {
  padding: 24px 0;
}
.header {
  background-color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 16px;
}
.header-left {
  display: flex;
  align-items: center;
}
.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  margin-right: 16px;
}
.page-title {
  font-size: 16px;
  font-weight: 500;
}
.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  outline: none;
}
.main {
  background-color: #f5f7fa;
  padding: 16px;
  overflow: auto;
}
</style>
