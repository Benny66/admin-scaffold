<template>
  <div>
    <el-card class="search-card" shadow="never">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="字典名/编码" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <div class="toolbar">
        <el-button type="primary" @click="openCreate">新增字典类型</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe @row-click="handleRowClick" highlight-current-row>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="字典名" width="140" />
        <el-table-column prop="code" label="编码" width="160" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click.stop="openEdit(row)">编辑</el-button>
            <el-button size="small" type="primary" @click.stop="openItems(row)">字典项</el-button>
            <el-button size="small" type="danger" @click.stop="handleDelete(row)">删除</el-button>
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

    <!-- 新增/编辑字典类型弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑字典类型' : '新增字典类型'" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="字典名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码" required>
          <el-input v-model="form.code" :disabled="isEdit" placeholder="如 user_status" />
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

    <!-- 字典项弹窗 -->
    <el-dialog v-model="itemsVisible" :title="`字典项 - ${currentType?.name || ''}`" width="620px">
      <div class="toolbar">
        <el-button type="primary" size="small" @click="openItemCreate">新增字典项</el-button>
      </div>
      <el-table :data="items" v-loading="itemsLoading" border stripe>
        <el-table-column prop="label" label="标签" width="140" />
        <el-table-column prop="value" label="值" width="140" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" />
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button size="small" @click="openItemEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleItemDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 字典项新增/编辑弹窗 -->
    <el-dialog v-model="itemDialogVisible" :title="itemIsEdit ? '编辑字典项' : '新增字典项'" width="420px" append-to-body>
      <el-form :model="itemForm" label-width="80px">
        <el-form-item label="标签" required>
          <el-input v-model="itemForm.label" />
        </el-form-item>
        <el-form-item label="值" required>
          <el-input v-model="itemForm.value" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="itemForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="itemForm.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleItemSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDictTypeList, createDictType, updateDictType, deleteDictType,
  getDictItemList, createDictItem, updateDictItem, deleteDictItem,
} from '@/api'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, page_size: 10, keyword: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = reactive({ id: null, name: '', code: '', description: '', sort: 0 })

const itemsVisible = ref(false)
const itemsLoading = ref(false)
const items = ref([])
const currentType = ref(null)

const itemDialogVisible = ref(false)
const itemIsEdit = ref(false)
const itemForm = reactive({ id: null, label: '', value: '', sort: 0, remark: '' })

async function fetchList() {
  loading.value = true
  try {
    const res = await getDictTypeList(query)
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
    ElMessage.warning('请填写字典名和编码')
    return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateDictType(form.id, { name: form.name, description: form.description, sort: form.sort })
      ElMessage.success('更新成功')
    } else {
      await createDictType({ name: form.name, code: form.code, description: form.description, sort: form.sort })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除字典类型「${row.name}」？其下字典项将一并删除`, '警告', { type: 'error' })
    .then(async () => {
      await deleteDictType(row.id)
      ElMessage.success('删除成功')
      fetchList()
    })
    .catch(() => {})
}

function handleRowClick(row) {
  openItems(row)
}

async function openItems(row) {
  currentType.value = row
  itemsVisible.value = true
  await fetchItems()
}

async function fetchItems() {
  itemsLoading.value = true
  try {
    const res = await getDictItemList({ page: 1, page_size: 100, dict_type_id: currentType.value.id })
    items.value = res.data.list
  } finally {
    itemsLoading.value = false
  }
}

function openItemCreate() {
  itemIsEdit.value = false
  Object.assign(itemForm, { id: null, label: '', value: '', sort: 0, remark: '' })
  itemDialogVisible.value = true
}
function openItemEdit(row) {
  itemIsEdit.value = true
  Object.assign(itemForm, { id: row.id, label: row.label, value: row.value, sort: row.sort, remark: row.remark })
  itemDialogVisible.value = true
}

async function handleItemSubmit() {
  if (!itemForm.label || !itemForm.value) {
    ElMessage.warning('请填写标签和值')
    return
  }
  submitting.value = true
  try {
    const payload = { label: itemForm.label, value: itemForm.value, sort: itemForm.sort, remark: itemForm.remark }
    if (itemIsEdit.value) {
      await updateDictItem(itemForm.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createDictItem({ ...payload, dict_type_id: currentType.value.id })
      ElMessage.success('创建成功')
    }
    itemDialogVisible.value = false
    fetchItems()
  } finally {
    submitting.value = false
  }
}

function handleItemDelete(row) {
  ElMessageBox.confirm(`确定删除字典项「${row.label}」？`, '警告', { type: 'error' })
    .then(async () => {
      await deleteDictItem(row.id)
      ElMessage.success('删除成功')
      fetchItems()
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
