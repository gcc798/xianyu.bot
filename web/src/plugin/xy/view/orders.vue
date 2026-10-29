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
          <el-select v-model="searchInfo.status" clearable placeholder="全部状态" class="w-44">
            <el-option label="待发货" value="pending_delivery" />
            <el-option label="已发货" value="shipped" />
            <el-option label="待识别" value="unknown" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchInfo.keyword" clearable placeholder="订单、商品或买家" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" :loading="syncing" @click="syncPending">同步待发货订单</el-button>
      </div>
      <el-alert title="订单来自真实交易消息或闲鱼卖家接口；开启商品自动发货后会调用闲鱼虚拟发货接口。" type="warning" :closable="false" class="mb-4" />
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="订单 ID" prop="externalOrderId" min-width="190" />
        <el-table-column label="商品" min-width="220">
          <template #default="scope">
            <div>{{ scope.row.productTitle || '-' }}</div>
            <div class="subtle">{{ scope.row.externalProductId || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="买家" min-width="160">
          <template #default="scope">
            <div>{{ scope.row.buyerUserName || '-' }}</div>
            <div class="subtle">{{ scope.row.buyerUserId || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="账号" prop="accountLabel" width="140" />
        <el-table-column label="金额" width="110">
          <template #default="scope">¥ {{ (scope.row.amountCent / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'shipped' ? 'success' : scope.row.status === 'pending_delivery' ? 'warning' : 'info'">
              {{ scope.row.status === 'shipped' ? '已发货' : scope.row.status === 'pending_delivery' ? '待发货' : '待识别' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="180">
          <template #default="scope">{{ new Date(scope.row.UpdatedAt).toLocaleString('zh-CN', { hour12: false }) }}</template>
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
  import { getXyAccountList } from '@/plugin/xy/api/account'
  import { getXyOrderList, syncXyPendingOrders } from '@/plugin/xy/api/order'

  defineOptions({ name: 'XyOrders' })

  const loading = ref(false)
  const syncing = ref(false)
  const tableData = ref([])
  const accounts = ref([])
  const total = ref(0)
  const searchInfo = reactive({ page: 1, pageSize: 10, keyword: '', accountId: undefined, status: '' })

  const load = async () => {
    loading.value = true
    try {
      const res = await getXyOrderList(searchInfo)
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }
  const search = () => { searchInfo.page = 1; load() }
  const resetSearch = () => { Object.assign(searchInfo, { page: 1, keyword: '', accountId: undefined, status: '' }); load() }
  const syncPending = async () => {
    if (!searchInfo.accountId) return ElMessage.warning('请先选择一个闲鱼账号')
    syncing.value = true
    try {
      const res = await syncXyPendingOrders({ accountId: searchInfo.accountId })
      if (res.code === 0) {
        ElMessage.success(`已同步 ${res.data.count} 条待发货订单`)
        load()
      }
    } finally {
      syncing.value = false
    }
  }

  onMounted(async () => {
    const res = await getXyAccountList({ page: 1, pageSize: 100 })
    if (res.code === 0) accounts.value = res.data.list || []
    load()
  })
</script>

<style scoped>
  .subtle {
    margin-top: 2px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
</style>
