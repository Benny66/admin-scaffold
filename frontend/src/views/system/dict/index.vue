<template>
  <div class="dict-page">
    <!-- 左栏：字典类型 -->
    <div class="left-pane">
      <div class="pane-header">
        <span class="pane-title">字典类型</span>
        <div class="pane-header-actions">
          <el-button v-permission="'dict:create'" type="primary" @click="openTypeCreate()">新增</el-button>
          <el-button v-permission="'dict:view'" @click="handleExportAll">导出全部</el-button>
        </div>
      </div>

      <div class="pane-search">
        <el-input v-model="typeQuery.keyword" placeholder="字典名/编码" clearable @keyup.enter="searchTypes" />
        <el-button type="primary" @click="searchTypes">查询</el-button>
      </div>

      <el-table
        :data="typeList"
        v-loading="typeLoading"
        border
        highlight-current-row
        :row-class-name="typeRowClass"
        @row-click="handleTypeRowClick"
      >
        <el-table-column prop="id" label="编号" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" show-overflow-tooltip />
        <el-table-column prop="code" label="编码" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <span class="row-ops">
              <el-button v-permission="'dict:edit'" link type="primary" size="small" @click.stop="openTypeEdit(row)">
                <el-icon><Edit />编辑</el-icon>
              </el-button>
              <el-button v-permission="'dict:delete'" link type="danger" size="small" @click.stop="handleTypeDelete(row)">
                <el-icon><Delete />删除</el-icon>
              </el-button>
            </span>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        layout="total, prev, pager, next"
        :current-page="typeQuery.page"
        :page-size="typeQuery.page_size"
        :total="typeTotal"
        @current-change="typePageChange"
      />
    </div>

    <!-- 右栏：当前选中类型的字典项 -->
    <div class="right-pane">
      <template v-if="currentType">
        <div class="right-head">
          <span class="type-title">{{ currentType.name }}</span>
          <span class="type-code">{{ currentType.code }}</span>
          <div class="toolbar">
            <el-button v-permission="'dict:create'" type="primary" @click="openItemCreate()">新增</el-button>
            <el-upload :show-file-list="false" accept=".xlsx" :http-request="handleImportUpload">
              <el-button v-permission="'dict:create'" :loading="uploading">导入</el-button>
            </el-upload>
            <el-button v-permission="'dict:view'" @click="handleExport">导出</el-button>
            <el-button v-permission="'dict:view'" @click="handleDownloadTemplate">下载模板</el-button>
          </div>
        </div>

        <el-table :data="itemList" v-loading="itemLoading" border>
          <el-table-column prop="label" label="标签" min-width="140" show-overflow-tooltip />
          <el-table-column prop="value" label="值" min-width="140" show-overflow-tooltip />
          <el-table-column prop="sort" label="排序" width="70" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" show-overflow-tooltip />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button v-permission="'dict:edit'" link type="primary" size="small" @click="openItemEdit(row)">编辑</el-button>
              <el-button v-permission="'dict:delete'" link type="danger" size="small" @click="handleItemDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          class="pagination"
          layout="total, prev, pager, next"
          :current-page="itemQuery.page"
          :page-size="itemQuery.page_size"
          :total="itemTotal"
          @current-change="itemPageChange"
        />
      </template>

      <el-empty v-else description="请选择左侧字典类型" />
    </div>

    <!-- 新增/编辑字典类型弹窗 -->
    <el-dialog v-model="typeDialogVisible" :title="typeIsEdit ? '编辑字典类型' : '新增字典类型'" width="420px">
      <el-form :model="typeForm" label-width="80px">
        <el-form-item label="字典名" required>
          <el-input v-model="typeForm.name" />
        </el-form-item>
        <el-form-item label="编码" required>
          <el-input v-model="typeForm.code" :disabled="typeIsEdit" placeholder="如 user_status" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="typeForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="typeForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleTypeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑字典项弹窗 -->
    <el-dialog v-model="itemDialogVisible" :title="itemIsEdit ? '编辑字典项' : '新增字典项'" width="420px">
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
import { Edit, Delete } from '@element-plus/icons-vue'
import {
  getDictTypeList, createDictType, updateDictType, deleteDictType,
  getDictItemList, createDictItem, updateDictItem, deleteDictItem,
  exportDictItems, exportAllDictTypes, importDictItems,
} from '@/api'

// ==================== 左栏：字典类型 ====================
const typeList = ref([])
const typeTotal = ref(0)
const typeLoading = ref(false)
const typeQuery = reactive({ page: 1, page_size: 10, keyword: '' })

// ==================== 右栏：字典项（currentTypeId 与左栏分页解耦）====================
const currentTypeId = ref(null)
const currentType = ref(null)
const itemList = ref([])
const itemTotal = ref(0)
const itemLoading = ref(false)
const itemQuery = reactive({ page: 1, page_size: 10 })

// ==================== 弹窗状态 ====================
const typeDialogVisible = ref(false)
const typeIsEdit = ref(false)
const typeForm = reactive({ id: null, name: '', code: '', description: '', sort: 0 })

const itemDialogVisible = ref(false)
const itemIsEdit = ref(false)
const itemForm = reactive({ id: null, label: '', value: '', sort: 0, remark: '' })

const submitting = ref(false)
const uploading = ref(false)

// ==================== 左栏逻辑 ====================
async function fetchTypes() {
  typeLoading.value = true
  try {
    const res = await getDictTypeList(typeQuery)
    typeList.value = res.data.list
    typeTotal.value = res.data.total
    // 首次加载自动选中第一个字典类型，让右栏默认有内容而非空态
    // （删除当前选中类型后也会复位后重新选中第一条，行为一致）
    if (!currentTypeId.value && typeList.value.length > 0) {
      handleTypeRowClick(typeList.value[0])
    }
  } finally {
    typeLoading.value = false
  }
}
function searchTypes() {
  typeQuery.page = 1
  fetchTypes()
}
function typePageChange(p) {
  typeQuery.page = p
  fetchTypes()
}
// 翻页后仍高亮当前选中类型（row-class-name，不依赖 el-table 内部选中态）
function typeRowClass({ row }) {
  return row.id === currentTypeId.value ? 'current-type-row' : ''
}
function handleTypeRowClick(row) {
  currentTypeId.value = row.id
  currentType.value = row
  itemQuery.page = 1
  fetchItems()
}

function openTypeCreate() {
  typeIsEdit.value = false
  Object.assign(typeForm, { id: null, name: '', code: '', description: '', sort: 0 })
  typeDialogVisible.value = true
}
function openTypeEdit(row) {
  typeIsEdit.value = true
  Object.assign(typeForm, { id: row.id, name: row.name, code: row.code, description: row.description, sort: row.sort })
  typeDialogVisible.value = true
}
async function handleTypeSubmit() {
  if (!typeForm.name || !typeForm.code) {
    ElMessage.warning('请填写字典名和编码')
    return
  }
  submitting.value = true
  try {
    if (typeIsEdit.value) {
      await updateDictType(typeForm.id, { name: typeForm.name, description: typeForm.description, sort: typeForm.sort })
      ElMessage.success('更新成功')
    } else {
      await createDictType({ name: typeForm.name, code: typeForm.code, description: typeForm.description, sort: typeForm.sort })
      ElMessage.success('创建成功')
    }
    typeDialogVisible.value = false
    fetchTypes()
  } finally {
    submitting.value = false
  }
}
function handleTypeDelete(row) {
  ElMessageBox.confirm(`确定删除字典类型「${row.name}」？其下字典项将一并删除`, '警告', { type: 'error' })
    .then(async () => {
      await deleteDictType(row.id)
      ElMessage.success('删除成功')
      // 删除的即当前选中类型 → 清空右栏复位为空态
      if (currentTypeId.value === row.id) {
        currentTypeId.value = null
        currentType.value = null
        itemList.value = []
        itemTotal.value = 0
      }
      fetchTypes()
    })
    .catch(() => {})
}

// ==================== 右栏逻辑 ====================
async function fetchItems() {
  if (!currentTypeId.value) return
  itemLoading.value = true
  try {
    const res = await getDictItemList({ ...itemQuery, dict_type_id: currentTypeId.value })
    itemList.value = res.data.list
    itemTotal.value = res.data.total
  } finally {
    itemLoading.value = false
  }
}
function itemPageChange(p) {
  itemQuery.page = p
  fetchItems()
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
      await createDictItem({ ...payload, dict_type_id: currentTypeId.value })
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

// ==================== 导入导出 ====================
// 触发 blob 文件下载
function downloadBlob(res, fallbackName) {
  const blob = new Blob([res.data], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
  let filename = fallbackName
  const cd = res.headers && res.headers['content-disposition']
  if (cd) {
    const m = /filename\*=UTF-8''([^;]+)/.exec(cd)
    if (m && m[1]) {
      try { filename = decodeURIComponent(m[1]) } catch (e) { /* 保留兜底文件名 */ }
    }
  }
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 导出当前选中类型
async function handleExport() {
  if (!currentTypeId.value) return
  const res = await exportDictItems(currentTypeId.value)
  downloadBlob(res, `${currentType.value.name}.xlsx`)
}

// 导出全部类型（多 sheet，全局操作）
async function handleExportAll() {
  const res = await exportAllDictTypes()
  downloadBlob(res, '全量字典.xlsx')
}

// 下载导入模板：复用单类型导出接口生成含表头、空数据行的文件
async function handleDownloadTemplate() {
  if (!currentTypeId.value) return
  const res = await exportDictItems(currentTypeId.value)
  downloadBlob(res, '字典导入模板.xlsx')
}

// 自定义上传：携带 currentTypeId 调后端导入接口
async function handleImportUpload(option) {
  if (!currentTypeId.value) {
    ElMessage.warning('请先选择字典类型')
    return
  }
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', option.file)
    const res = await importDictItems(currentTypeId.value, formData)
    ElMessage.success(`导入完成：新增 ${res.data.new} 条，更新 ${res.data.updated} 条`)
    fetchItems()
  } finally {
    uploading.value = false
  }
}

onMounted(() => {
  fetchTypes()
})
</script>

<style scoped>
.dict-page {
  display: flex;
  gap: 20px;
  align-items: stretch;
  padding: 4px;
}
.left-pane,
.right-pane {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
}
.left-pane {
  width: 500px;
  flex-shrink: 0;
}
.right-pane {
  flex: 1;
  min-width: 0;
}
.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.pane-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.pane-header-actions {
  display: flex;
  gap: 8px;
}
.pane-search {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
.right-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
}
.type-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.type-code {
  color: #909399;
  font-size: 13px;
}
.right-head .toolbar {
  margin-left: auto;
}
/* 左栏行内 hover 显隐编辑/删除图标 */
.row-ops {
  display: inline-flex;
}

:deep(.current-type-row) {
  background-color: #ecf5ff !important;
}
</style>