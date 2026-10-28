<template>
  <div class="gva-table-box settings-page">
    <div class="settings-heading"><div><div class="eyebrow">SERVICE BOUNDARY</div><h1>Telegram 服务设置</h1><p>敏感配置仍由容器环境变量管理；这里只展示可安全确认的运行状态。</p></div><el-button @click="load">刷新</el-button></div>
    <el-alert title="修改 TG_API_ID、TG_API_HASH、代理或内部访问令牌后，需要重建 telegram 与 server 服务。登录会话保存在数据库中。" type="info" show-icon :closable="false" class="mb-5" />
    <el-descriptions v-loading="loading" :column="1" border>
      <el-descriptions-item label="Telegram API"><el-tag :type="settings.telegram_api_configured ? 'success' : 'danger'">{{ settings.telegram_api_configured ? '已配置' : '未配置' }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="登录状态"><el-tag :type="settings.authorized ? 'success' : 'warning'">{{ settings.authorized ? `已登录 · ${settings.account || ''}` : '未登录' }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="SOCKS5 代理">{{ settings.proxy || '未配置' }}</el-descriptions-item>
      <el-descriptions-item label="容器监听">{{ settings.listen || '-' }}</el-descriptions-item>
      <el-descriptions-item label="资源并发">{{ settings.resource_concurrency || 1 }} 个任务</el-descriptions-item>
      <el-descriptions-item label="单文件上限">{{ formatSize(settings.resource_max_file_size) }}</el-descriptions-item>
      <el-descriptions-item label="调用策略">串行扫描；FloodWait 等待后仅重试一次普通 Telegram API 调用</el-descriptions-item>
    </el-descriptions>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { getTelegramSettings } from '@/plugin/xianyu/api/telegram-resource'

defineOptions({ name: 'XianyuTelegramSettings' })

const settings = reactive({})
const loading = ref(false)
const load = async () => {
  loading.value = true
  try { const res = await getTelegramSettings(); if (res.code === 0) Object.assign(settings, res.data) } finally { loading.value = false }
}
const formatSize = (size = 0) => size ? `${Math.round(size / 1024 / 1024)} MB` : '-'
onMounted(load)
</script>

<style scoped>
.settings-page { max-width: 980px; }
.settings-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 24px; }
.settings-heading h1 { margin: 4px 0 8px; font-size: 24px; }
.settings-heading p { margin: 0; color: var(--el-text-color-secondary); }
.eyebrow { color: #0f8f86; font-size: 11px; letter-spacing: .2em; }
</style>
