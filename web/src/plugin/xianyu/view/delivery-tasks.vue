<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="账号">
          <el-select v-model="searchInfo.accountId" clearable placeholder="全部账号" class="w-52">
            <el-option v-for="item in accounts" :key="item.ID" :label="item.label" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部状态" class="w-40">
            <el-option label="待执行" value="pending" />
            <el-option label="发送中" value="sending" />
            <el-option label="已发送待确认" value="sent_unconfirmed" />
            <el-option label="已确认" value="confirmed" />
            <el-option label="失败" value="failed" />
            <el-option label="结果未知" value="unknown" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchInfo.keyword" clearable placeholder="订单号、商品或买家" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button :loading="loading" @click="load">刷新</el-button>
      </div>
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="创建时间" width="180">
          <template #default="scope">{{ formatTime(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="订单号" prop="externalOrderId" min-width="180" />
        <el-table-column label="商品" prop="productTitle" min-width="200" />
        <el-table-column label="买家" prop="buyerUserName" min-width="130" />
        <el-table-column label="账号" prop="accountLabel" min-width="130" />
        <el-table-column label="模式" width="110">
          <template #default="scope">{{ scope.row.mode === 'card' ? '卡密发货' : '文本发货' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="140">
          <template #default="scope">
            <el-tag :type="taskTag(scope.row.status)">{{ taskStatus(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="尝试" prop="attempts" width="80" />
        <el-table-column label="说明" prop="lastError" min-width="220" />
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="scope">
            <el-button v-if="scope.row.status === 'failed'" link type="primary" @click="retryTask(scope.row)">重试</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="gva-pagination">
        <el-pagination
          v-model:current-page="searchInfo.page"
          v-model:page-size="searchInfo.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getXianyuAccountList } from '@/plugin/xianyu/api/account'
import { getXianyuDeliveryTaskList, retryXianyuDeliveryTask } from '@/plugin/xianyu/api/delivery-rule'

defineOptions({ name: 'XianyuDeliveryTasks' })

const loading = ref(false)
const accounts = ref([])
const tableData = ref([])
const total = ref(0)
const searchInfo = reactive({ page: 1, pageSize: 10, accountId: undefined, status: '', keyword: '' })

const load = async () => {
  loading.value = true
  try {
    const res = await getXianyuDeliveryTaskList(searchInfo)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

const loadAccounts = async () => {
  const res = await getXianyuAccountList({ page: 1, pageSize: 100 })
  if (res.code === 0) accounts.value = res.data.list || []
}

const search = () => {
  searchInfo.page = 1
  load()
}

const resetSearch = () => {
  Object.assign(searchInfo, { page: 1, pageSize: 10, accountId: undefined, status: '', keyword: '' })
  load()
}

const retryTask = async (row) => {
  const res = await retryXianyuDeliveryTask({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('任务已执行')
    load()
  }
}

const formatTime = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
const taskStatus = (status) => ({ pending: '待执行', sending: '发送中', sent_unconfirmed: '已发送待确认', failed: '失败', unknown: '结果未知', confirmed: '已确认' }[status] || status)
const taskTag = (status) => ({ confirmed: 'success', sent_unconfirmed: 'warning', failed: 'danger', unknown: 'danger' }[status] || 'info')

onMounted(() => {
  loadAccounts()
  load()
})
</script>
