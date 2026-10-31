<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="账号"><el-select v-model="searchInfo.accountId" clearable placeholder="全部账号" class="w-52"><el-option v-for="item in accounts" :key="item.ID" :label="item.label" :value="item.ID" /></el-select></el-form-item>
        <el-form-item label="类型"><el-select v-model="searchInfo.type" clearable placeholder="全部类型" class="w-40"><el-option label="登录检查" value="login_check" /><el-option label="商品同步" value="product_sync" /><el-option label="订单同步" value="order_sync" /><el-option label="消息" value="message" /><el-option label="自动发货" value="delivery" /></el-select></el-form-item>
        <el-form-item label="状态"><el-select v-model="searchInfo.status" clearable placeholder="全部状态" class="w-36"><el-option label="成功" value="success" /><el-option label="失败" value="failed" /><el-option label="结果未知" value="unknown" /></el-select></el-form-item>
        <el-form-item label="关键词"><el-input v-model="searchInfo.keyword" clearable /></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="rows" row-key="ID" show-overflow-tooltip>
        <el-table-column label="时间" width="180"><template #default="scope">{{ new Date(scope.row.CreatedAt).toLocaleString('zh-CN', { hour12: false }) }}</template></el-table-column>
        <el-table-column label="账号" prop="accountLabel" width="140" />
        <el-table-column label="类型" prop="type" width="130" />
        <el-table-column label="状态" width="100"><template #default="scope"><el-tag :type="scope.row.status === 'success' || scope.row.status === 'ready' ? 'success' : 'danger'">{{ scope.row.status }}</el-tag></template></el-table-column>
        <el-table-column label="摘要" prop="summary" min-width="220" />
        <el-table-column label="错误" prop="error" min-width="240" show-overflow-tooltip />
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="searchInfo.page" v-model:page-size="searchInfo.pageSize" :total="total" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { getXyAccountList } from '@/api/xianyu/account'
import { getXyBusinessEventList } from '@/api/xianyu/business-event'

defineOptions({ name: 'XyBusinessEvents' })
const loading = ref(false); const rows = ref([]); const accounts = ref([]); const total = ref(0)
const searchInfo = reactive({ page: 1, pageSize: 10, accountId: undefined, type: '', status: '', keyword: '' })
const load = async () => { loading.value = true; try { const res = await getXyBusinessEventList(searchInfo); if (res.code === 0) { rows.value = res.data.list || []; total.value = res.data.total } } finally { loading.value = false } }
const search = () => { searchInfo.page = 1; load() }
const reset = () => { Object.assign(searchInfo, { page: 1, pageSize: 10, accountId: undefined, type: '', status: '', keyword: '' }); load() }
onMounted(async () => { const res = await getXyAccountList({ page: 1, pageSize: 100 }); if (res.code === 0) accounts.value = res.data.list || []; load() })
</script>
