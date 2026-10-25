<template>
  <div class="gva-table-box">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-medium">闲鱼助手</h1>
        <p class="mt-2 text-gray-500">真实账号、商品、消息、订单和自动化任务概览。</p>
      </div>
      <el-tag :type="status.plugin === 'ready' ? 'success' : 'danger'">
        {{ loading ? '检查中' : status.plugin === 'ready' ? '服务正常' : '服务异常' }}
      </el-tag>
    </div>

    <el-row :gutter="16" class="mt-6">
      <el-col v-for="item in metrics" :key="item.label" :xs="12" :sm="8" :lg="4">
        <el-card shadow="never" class="mb-4">
          <div class="text-sm text-gray-500">{{ item.label }}</div>
          <div class="mt-2 text-2xl font-semibold">{{ item.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-descriptions :column="1" border>
      <el-descriptions-item label="插件状态">{{ status.plugin || '-' }}</el-descriptions-item>
      <el-descriptions-item label="数据库">{{ status.database || '-' }}</el-descriptions-item>
    </el-descriptions>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getXianyuStatus } from '@/plugin/xianyu/api/status'

  defineOptions({ name: 'XianyuDashboard' })

  const loading = ref(true)
  const status = ref({})
  const metrics = computed(() => [
    { label: '账号', value: status.value.accounts || 0 },
    { label: '登录正常', value: status.value.readyAccounts || 0 },
    { label: '在售商品', value: status.value.onSaleProducts || 0 },
    { label: '消息', value: status.value.messages || 0 },
    { label: '待发货', value: status.value.pendingOrders || 0 },
    { label: '发货异常', value: status.value.failedDeliveries || 0 }
  ])

  onMounted(async () => {
    try {
      const res = await getXianyuStatus()
      if (res.code === 0) status.value = res.data
    } catch {
      ElMessage.error('闲鱼助手服务状态检查失败')
    } finally {
      loading.value = false
    }
  })
</script>
