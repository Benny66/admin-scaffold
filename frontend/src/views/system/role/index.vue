<template>
  <div>
    <el-card class="search-card" shadow="never">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="角色名/编码" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <div class="toolbar">
        <el-button v-permission="'roles:create'" type="primary" @click="openCreate">新增角色</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="角色名" width="140" />
        <el-table-column prop="code" label="编码" width="140" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'roles:edit'" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'roles:edit'" size="small" type="primary" @click="openAssign(row)">分配权限</el-button>
            <el-button v-permission="'roles:delete'" size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="角色名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码" required>
          <el-input v-model="form.code" :disabled="isEdit" placeholder="如 user / admin" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
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

    <!-- 分配权限弹窗 -->
    <el-dialog v-model="assignVisible" title="分配权限" width="480px">
      <el-tree
        ref="treeRef"
        :data="permissionTree"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        default-expand-all
      />
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRoleList, createRole, updateRole, deleteRole,
  getAllPermissions, getRolePermissions, assignPermissionsToRole,
} from '@/api'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, page_size: 10, keyword: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = reactive({ id: null, name: '', code: '', description: '', sort: 0 })

const assignVisible = ref(false)
const assignRoleId = ref(null)
const permissionTree = ref([])
const treeRef = ref(null)

async function fetchList() {
  loading.value = true
  try {
    const res = await getRoleList(query)
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
  Object.assign(form, { id: null, name: '', code: '', description: '', sort: 0 })
  dialogVisible.value = true
}
function openEdit(row) {
  isEdit.value = true
  Object.assign(form, { id: row.id, name: row.name, code: row.code, description: row.description, sort: row.sort })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.name || !form.code) {
    ElMessage.warning('请填写角色名和编码')
    return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateRole(form.id, { name: form.name, description: form.description, sort: form.sort })
      ElMessage.success('更新成功')
    } else {
      await createRole({ name: form.name, code: form.code, description: form.description, sort: form.sort })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除角色「${row.name}」？`, '警告', { type: 'error' })
    .then(async () => {
      await deleteRole(row.id)
      ElMessage.success('删除成功')
      fetchList()
    })
    .catch(() => {})
}

async function openAssign(row) {
  assignRoleId.value = row.id
  // 构建权限树
  const permsRes = await getAllPermissions()
  const perms = permsRes.data || []
  const tree = []
  const map = {}
  perms.forEach((p) => {
    map[p.id] = { ...p, children: [] }
  })
  perms.forEach((p) => {
    if (p.parent_id && map[p.parent_id]) {
      map[p.parent_id].children.push(map[p.id])
    } else {
      tree.push(map[p.id])
    }
  })
  permissionTree.value = tree

  // 回显已分配权限
  const assignedRes = await getRolePermissions(row.id)
  const assigned = assignedRes.data || []
  assignVisible.value = true
  setTimeout(() => {
    treeRef.value?.setCheckedKeys(assigned)
  }, 100)
}

async function handleAssign() {
  const checkedKeys = treeRef.value?.getCheckedKeys() || []
  submitting.value = true
  try {
    await assignPermissionsToRole(assignRoleId.value, { permission_ids: checkedKeys })
    ElMessage.success('权限分配成功')
    assignVisible.value = false
  } finally {
    submitting.value = false
  }
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
