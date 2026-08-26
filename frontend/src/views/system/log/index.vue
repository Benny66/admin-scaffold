<template>
  <div>
    <el-card class="search-card" shadow="never">
      <el-form inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="用户名/路径" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="操作日志" name="operation" />
        <el-tab-pane label="登录日志" name="login" />
      </el-tabs>

      <div class="toolbar">
        <el-button v-if="activeTab === 'operation'" type="danger" @click="handleClear">清空操作日志</el-button>
      </div>

      <!-- 操作日志 -->
      <el-table v-if="activeTab === 'operation'" :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户" width="100" />
        <el-table-column prop="method" label="方法" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时(ms)" width="90" />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>

      <!-- 登录日志 -->
      <el-table v-else :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="fail_msg" label="失败原因" />
        <el-table-column prop="login_time" label="登录时间" width="180" />
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOperationLogList, getLoginLogList, clearOperationLog } from '@/api'

const activeTab = ref('operation')
const list = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, page_size: 10, keyword: '' })

async function fetchList() {
  loading.value = true
  try {
    const fetchFn = activeTab.value === 'operation' ? getOperationLogList : getLoginLogList
    const res = await fetchFn(query)
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
function handleTabChange() {
  query.page = 1
  fetchList()
}

function handleClear() {
  ElMessageBox.confirm('确定清空所有操作日志？此操作不可恢复', '警告', { type: 'error' })
    .then(async () => {
      await clearOperationLog()
      ElMessage.success('清空成功')
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
