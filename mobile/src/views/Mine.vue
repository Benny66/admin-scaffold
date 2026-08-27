<template>
  <div class="mine-page">
    <van-nav-bar title="我的" left-arrow @click-left="$router.back()" />

    <div class="profile">
      <van-image round width="70" height="70" src="https://img.yzcdn.cn/vant/cat.jpeg" />
      <div class="profile-info">
        <div class="name">{{ userInfo?.real_name || userInfo?.username || '未登录' }}</div>
        <div class="desc">{{ userInfo?.email || '' }}</div>
      </div>
    </div>

    <van-cell-group inset title="账号">
      <van-cell title="修改密码" icon="lock" is-link @click="passwordDialog = true" />
    </van-cell-group>

    <div class="logout-btn">
      <van-button round block type="danger" @click="handleLogout">退出登录</van-button>
    </div>

    <van-dialog v-model:show="passwordDialog" title="修改密码" show-cancel-button @confirm="handleChangePassword">
      <div class="dialog-body">
        <van-field v-model="passwordForm.old_password" type="password" label="原密码" placeholder="请输入原密码" />
        <van-field v-model="passwordForm.new_password" type="password" label="新密码" placeholder="请输入新密码" />
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useAppStore } from '@/stores/app'
import { changePassword } from '@/api'

const router = useRouter()
const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)

const passwordDialog = ref(false)
const passwordForm = reactive({ old_password: '', new_password: '' })

async function handleChangePassword() {
  if (!passwordForm.old_password || !passwordForm.new_password) {
    showToast('请填写完整信息')
    return
  }
  const res = await changePassword({
    old_password: passwordForm.old_password,
    new_password: passwordForm.new_password,
  })
  if (res.code === 200) {
    showToast('密码修改成功，请重新登录')
    appStore.logout()
    router.push('/login')
  }
}

function handleLogout() {
  appStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.mine-page {
  min-height: 100%;
}
.profile {
  display: flex;
  align-items: center;
  padding: 24px 16px;
  background: #fff;
  margin-bottom: 16px;
}
.profile-info {
  margin-left: 16px;
}
.name {
  font-size: 18px;
  font-weight: bold;
}
.desc {
  color: #999;
  font-size: 13px;
  margin-top: 4px;
}
.logout-btn {
  margin: 32px 16px;
}
.dialog-body {
  padding: 16px;
}
</style>
