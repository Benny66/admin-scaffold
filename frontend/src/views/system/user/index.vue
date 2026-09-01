<template>
  <div>
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="用户名/姓名/手机号" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <div class="toolbar">
        <el-button v-permission="'users:create'" type="primary" @click="openCreate">新增用户</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="real_name" label="姓名" width="100" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="角色" width="160">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r.id" size="small" style="margin-right: 4px">
              {{ r.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'users:edit'" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'users:edit'" size="small" type="warning" @click="handleResetPassword(row)">重置密码</el-button>
            <el-button v-permission="'users:edit'" size="small" :type="row.status === 1 ? 'danger' : 'success'" @click="handleToggleStatus(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button v-permission="'users:delete'" size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        :current-page="query.page"
        :page-size="query.page_size"
        :total="total"
        layout="total, prev, pager, next, sizes"
        :page-sizes="[10, 20, 50]"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="isEdit" placeholder="登录用户名" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="留空默认 123456" />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.real_name" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_ids" multiple placeholder="请选择角色" style="width: 100%">
            <el-option v-for="r in allRoles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getUserList, createUser, updateUser, deleteUser, resetPassword, toggleUserStatus,
  getAllRoles,
} from '@/api'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const allRoles = ref([])

const query = reactive({ page: 1, page_size: 10, keyword: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = reactive({
  id: null, username: '', password: '', real_name: '', email: '', phone: '',
  status: 1, role_ids: [],
})

async function fetchList() {
  loading.value = true
  try {
    const res = await getUserList(query)
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  fetchList()
}

function handleReset() {
  query.keyword = ''
  handleSearch()
}

function handlePageChange(p) {
  query.page = p
  fetchList()
}

function handleSizeChange(s) {
  query.page_size = s
  query.page = 1
  fetchList()
}

function openCreate() {
  isEdit.value = false
  Object.assign(form, { id: null, username: '', password: '', real_name: '', email: '', phone: '', status: 1, role_ids: [] })
  dialogVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, username: row.username, password: '', real_name: row.real_name,
    email: row.email, phone: row.phone, status: row.status,
    role_ids: (row.roles || []).map((r) => r.id),
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.username) {
    ElMessage.warning('请输入用户名')
    return
  }
  submitting.value = true
  try {
    const payload = {
      username: form.username, password: form.password, real_name: form.real_name,
      email: form.email, phone: form.phone, status: form.status, role_ids: form.role_ids,
    }
    if (isEdit.value) {
      await updateUser(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createUser(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

function handleResetPassword(row) {
  ElMessageBox.prompt(`请输入用户「${row.username}」的新密码`, '重置密码', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputType: 'password',
    inputValidator: (v) => (v && v.length >= 6) || '密码至少 6 位',
  }).then(async ({ value }) => {
    await resetPassword(row.id, { password: value })
    ElMessage.success('密码重置成功')
  }).catch(() => {})
}

function handleToggleStatus(row) {
  const action = row.status === 1 ? '禁用' : '启用'
  ElMessageBox.confirm(`确定${action}用户「${row.username}」？`, '提示', { type: 'warning' })
    .then(async () => {
      await toggleUserStatus(row.id, { status: row.status === 1 ? 0 : 1 })
      ElMessage.success(`${action}成功`)
      fetchList()
    })
    .catch(() => {})
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除用户「${row.username}」？此操作不可恢复`, '警告', { type: 'error' })
    .then(async () => {
      await deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchList()
    })
    .catch(() => {})
}

onMounted(async () => {
  fetchList()
  const res = await getAllRoles()
  allRoles.value = res.data || []
})
</script>

<style scoped>
.search-card {
  margin-bottom: 16px;
}
.toolbar {
  margin-bottom: 16px;
}
.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
