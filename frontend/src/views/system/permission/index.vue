<template>
  <div>
    <el-card class="search-card" shadow="never">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="权限名/编码" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <div class="toolbar">
        <el-button v-permission="'permissions:create'" type="primary" @click="openCreate">新增权限</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="权限名" width="150" />
        <el-table-column prop="code" label="权限编码" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'permissions:edit'" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'permissions:delete'" size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        :current-page="query.page"
        :page-size="query.page_size"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新增权限'" width="460px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="权限名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="权限编码" required>
          <el-input v-model="form.code" :disabled="isEdit" placeholder="如 users:view" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
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
import { getPermissionList, createPermission, updatePermission, deletePermission } from '@/api'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, page_size: 10, keyword: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = reactive({ id: null, name: '', code: '', type: 'api', sort: 0 })

function typeLabel(type) {
  return { menu: '菜单', button: '按钮', api: '接口' }[type] || type
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getPermissionList(query)
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

function openCreate() {
  isEdit.value = false
  Object.assign(form, { id: null, name: '', code: '', type: 'api', sort: 0 })
  dialogVisible.value = true
}
function openEdit(row) {
  isEdit.value = true
  Object.assign(form, { id: row.id, name: row.name, code: row.code, type: row.type, sort: row.sort })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.name || !form.code) {
    ElMessage.warning('请填写权限名和编码')
    return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await updatePermission(form.id, { name: form.name, type: form.type, sort: form.sort })
      ElMessage.success('更新成功')
    } else {
      await createPermission({ name: form.name, code: form.code, type: form.type, sort: form.sort })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除权限「${row.name}」？`, '警告', { type: 'error' })
    .then(async () => {
      await deletePermission(row.id)
      ElMessage.success('删除成功')
      fetchList()
    })
    .catch(() => {})
}

onMounted(() => {
  fetchList()
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
