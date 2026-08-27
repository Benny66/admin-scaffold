<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="aside">
      <div class="logo">
        <img v-if="appStore.logo" :src="appStore.logo" class="logo-img" alt="logo" />
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
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
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
        <div v-if="appStore.footer" class="app-footer">{{ appStore.footer }}</div>
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

const router = useRouter()
const appStore = useAppStore()

const isCollapse = ref(false)
const userInfo = computed(() => appStore.userInfo)
const systemName = computed(() => appStore.systemName || 'Base Admin')

const menus = [
  { path: '/system/user', title: '用户管理', icon: 'User' },
  { path: '/system/role', title: '角色管理', icon: 'UserFilled' },
  { path: '/system/permission', title: '权限管理', icon: 'Key' },
  { path: '/system/dict', title: '字典管理', icon: 'Collection' },
  { path: '/system/log', title: '操作日志', icon: 'Document' },
]

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
.app-footer {
  margin-top: 24px;
  padding-top: 16px;
  text-align: center;
  color: #999;
  font-size: 13px;
  border-top: 1px solid #e4e7ed;
}
</style>
