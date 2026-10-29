<template>
  <div class="telegram-page">
    <section class="telegram-hero">
      <div>
        <div class="telegram-eyebrow">TELEGRAM SOURCE</div>
        <h1>资源入口概览</h1>
        <p>在 GVA 内完成登录、会话扫描和资源入口维护，Python 服务继续负责 Telegram 协议连接。</p>
      </div>
      <el-tag :type="status.authorized ? 'success' : 'warning'" size="large" effect="dark">
        {{ status.authorized ? `已登录 · ${status.account || 'Telegram'}` : '未登录' }}
      </el-tag>
    </section>

    <el-row :gutter="16" class="mb-4">
      <el-col v-for="metric in metrics" :key="metric.label" :xs="12" :sm="8" :lg="4">
        <div class="metric-card">
          <span>{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="14">
        <div class="gva-table-box h-full">
          <div class="section-title">
            <div><h2>会话扫描</h2><p>刷新已加入的群组和频道，检索范围以最近一次扫描结果为准。</p></div>
            <el-button type="primary" :disabled="!status.authorized || scan.status === 'running'" @click="startScan">{{ scan.status === 'running' ? '扫描中' : '开始扫描' }}</el-button>
          </div>
          <el-progress v-if="scan.status === 'running'" :percentage="scanPercentage" :stroke-width="12" striped striped-flow />
          <el-descriptions :column="1" border class="mt-4">
            <el-descriptions-item label="扫描状态">{{ scanLabel }}</el-descriptions-item>
            <el-descriptions-item label="当前会话">{{ scan.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="最近完成">{{ formatTime(status.latest_scan?.finished_at) }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </el-col>

      <el-col :xs="24" :lg="10">
        <div class="gva-table-box login-panel">
          <div class="section-title"><div><h2>Telegram 登录</h2><p>二维码和两步验证均在当前页面完成。</p></div></div>
          <el-alert v-if="!status.configured" title="服务尚未配置 TG_API_ID / TG_API_HASH" type="warning" show-icon :closable="false" />
          <template v-else-if="!status.authorized">
            <div v-if="login.image" class="qr-shell"><img :src="login.image" alt="Telegram 登录二维码"><span>Telegram → 设置 → 设备 → 链接桌面设备</span></div>
            <el-alert v-if="login.error" :title="login.error" type="error" show-icon :closable="false" class="mb-3" />
            <el-input v-if="login.status === '2fa'" v-model="password" type="password" show-password placeholder="输入 Telegram 云密码" class="mb-3" @keyup.enter="submitPassword" />
            <el-button v-if="login.status === '2fa'" type="primary" class="w-full" @click="submitPassword">完成两步验证</el-button>
            <el-button v-else type="primary" class="w-full" :loading="login.status === 'connecting'" @click="createQRCode">{{ login.image ? '刷新二维码' : '生成登录二维码' }}</el-button>
          </template>
          <el-result v-else icon="success" title="Telegram 已连接" :sub-title="status.account || '登录会话可用'" />
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getTelegramAdminStatus, getTelegramLoginStatus, getTelegramScanStatus, startTelegramLogin, startTelegramScan, submitTelegramPassword } from '@/plugin/xy/api/telegram-resource'

defineOptions({ name: 'XyTelegramOverview' })

const status = reactive({ authorized: false, configured: false, summary: {}, latest_scan: null, account: '' })
const login = reactive({ status: 'idle', image: '', error: '' })
const scan = reactive({ status: 'idle', current: 0, total: 0, name: '', message: '' })
const password = ref('')
let timer

const metrics = computed(() => [
  { label: '会话总数', value: status.summary?.total || 0 },
  { label: '群组', value: status.summary?.groups_count || 0 },
  { label: '频道', value: status.summary?.channels_count || 0 },
  { label: '异常', value: status.summary?.abnormal || 0 },
  { label: '静音', value: status.summary?.muted || 0 },
  { label: '未读', value: status.summary?.unread || 0 }
])
const scanPercentage = computed(() => scan.total ? Math.round(scan.current / scan.total * 100) : 0)
const scanLabel = computed(() => scan.message || ({ idle: '尚未开始', completed: '扫描完成', failed: '扫描失败' }[scan.status] || scan.status))

const loadStatus = async () => {
  const res = await getTelegramAdminStatus()
  if (res.code === 0) {
    Object.assign(status, res.data)
    Object.assign(login, res.data.qr || {})
  }
}
const createQRCode = async () => {
  const res = await startTelegramLogin()
  if (res.code === 0) Object.assign(login, res.data)
}
const submitPassword = async () => {
  if (!password.value) return
  const res = await submitTelegramPassword(password.value)
  if (res.code === 0) {
    password.value = ''
    Object.assign(login, res.data)
  }
}
const startScan = async () => {
  const res = await startTelegramScan()
  if (res.code === 0) {
    Object.assign(scan, res.data.progress)
    ElMessage.success(res.data.started ? '扫描已开始' : '扫描正在进行')
  }
}
const poll = async () => {
  if (!status.authorized && ['connecting', 'waiting', '2fa'].includes(login.status)) {
    const res = await getTelegramLoginStatus()
    if (res.code === 0) {
      Object.assign(login, res.data)
      if (res.data.status === 'success') await loadStatus()
    }
  }
  if (status.authorized && scan.status === 'running') {
    const res = await getTelegramScanStatus()
    if (res.code === 0) {
      Object.assign(scan, res.data)
      if (res.data.status === 'completed') await loadStatus()
    }
  }
}
const formatTime = (value) => value ? new Date(value).toLocaleString() : '尚未扫描'

onMounted(async () => {
  try { await loadStatus() } catch { ElMessage.error('Telegram 服务状态获取失败') }
  timer = setInterval(poll, 1500)
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<style scoped>
.telegram-page { padding: 0; }
.telegram-hero { display: flex; align-items: flex-start; justify-content: space-between; gap: 24px; margin-bottom: 16px; padding: 28px 32px; color: #eef9ff; border-radius: 12px; background: linear-gradient(118deg, #13283d, #0d6b78 68%, #22a59a); box-shadow: 0 16px 40px rgb(10 55 70 / 16%); }
.telegram-hero h1 { margin: 3px 0 8px; font-size: 26px; font-weight: 650; letter-spacing: .02em; }
.telegram-hero p { margin: 0; color: rgb(238 249 255 / 76%); }
.telegram-eyebrow { font-size: 11px; letter-spacing: .22em; color: #78eadc; }
.metric-card { margin-bottom: 16px; padding: 18px 20px; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-bg-color); }
.metric-card span { display: block; color: var(--el-text-color-secondary); font-size: 13px; }
.metric-card strong { display: block; margin-top: 6px; font-size: 25px; font-variant-numeric: tabular-nums; }
.section-title { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-title h2 { margin: 0; font-size: 18px; }
.section-title p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
.login-panel { min-height: 310px; }
.qr-shell { display: flex; flex-direction: column; align-items: center; gap: 10px; margin-bottom: 16px; color: var(--el-text-color-secondary); font-size: 12px; }
.qr-shell img { width: 190px; height: 190px; padding: 8px; border: 1px solid var(--el-border-color); border-radius: 12px; background: #fff; }
@media (max-width: 768px) { .telegram-hero { padding: 22px; } .telegram-hero { flex-direction: column; } }
</style>
