<template>
  <div class="workspace" v-loading="loading">
    <section class="workspace-hero">
      <div>
        <div class="eyebrow">OPERATIONS OVERVIEW</div>
        <h1>闲鱼运营工作台</h1>
        <p>聚合账号、消息、订单与自动化执行情况，异常会优先浮到视线中。</p>
      </div>
      <div class="service-state" :class="serviceHealthy ? 'is-healthy' : 'is-warning'">
        <span class="state-dot" />
        <div>
          <strong>{{ serviceHealthy ? '服务运行正常' : '存在待处理状态' }}</strong>
          <small>{{ overview.metrics.readyAccounts || 0 }}/{{ overview.metrics.accounts || 0 }} 个账号可用</small>
        </div>
      </div>
    </section>

    <section class="filter-bar">
      <el-select v-model="accountId" clearable placeholder="全部账号" class="account-select" @change="loadOverview">
        <el-option v-for="account in accounts" :key="account.ID" :label="account.label" :value="account.ID" />
      </el-select>
      <el-radio-group v-model="datePreset" @change="applyPreset">
        <el-radio-button value="today">今天</el-radio-button>
        <el-radio-button value="7d">近 7 天</el-radio-button>
        <el-radio-button value="30d">近 30 天</el-radio-button>
        <el-radio-button value="custom">自定义</el-radio-button>
      </el-radio-group>
      <el-date-picker
        v-if="datePreset === 'custom'"
        v-model="dateRange"
        type="datetimerange"
        value-format="x"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        :clearable="false"
        class="date-picker"
        @change="loadOverview"
      />
      <button class="refresh-button" type="button" :disabled="loading" @click="loadOverview">
        <el-icon><Refresh /></el-icon>
        刷新
      </button>
      <span class="updated-at">更新于 {{ updatedAt || '-' }}</span>
    </section>

    <section class="metric-grid">
      <button v-for="metric in metricCards" :key="metric.label" type="button" class="metric-card" :class="`tone-${metric.tone}`" @click="openMetric(metric.route)">
        <span class="metric-label">{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <span class="metric-note">{{ metric.note }}</span>
        <span class="metric-line" />
      </button>
    </section>

    <section class="dashboard-grid">
      <article class="panel message-panel">
        <header class="panel-header">
          <div><span class="panel-kicker">MESSAGE FLOW</span><h2>消息趋势</h2></div>
          <el-button link type="primary" @click="openMessages">查看消息明细 <el-icon><ArrowRight /></el-icon></el-button>
        </header>
        <Chart :options="messageChartOptions" height="340px" />
      </article>

      <article class="panel source-panel">
        <header class="panel-header"><div><span class="panel-kicker">MESSAGE MIX</span><h2>消息构成</h2></div></header>
        <Chart :options="messageSourceOptions" height="250px" />
        <div class="source-summary">
          <div v-for="source in overview.messageSources" :key="source.name"><span>{{ source.name }}</span><strong>{{ source.value }}</strong></div>
        </div>
      </article>

      <article class="panel automation-panel">
        <header class="panel-header">
          <div><span class="panel-kicker">AUTOMATION</span><h2>自动化效果</h2></div>
          <el-button link type="primary" @click="openEvents">查看业务日志 <el-icon><ArrowRight /></el-icon></el-button>
        </header>
        <div class="automation-list">
          <div v-for="row in automationRows" :key="row.label" class="automation-row">
            <div class="automation-title"><span>{{ row.label }}</span><span><b>{{ row.success }}</b> 成功 · <em>{{ row.failed }}</em> 异常</span></div>
            <div class="progress-track"><span :style="{ width: `${row.rate}%` }" /></div>
          </div>
        </div>
      </article>

      <article class="panel order-panel">
        <header class="panel-header">
          <div><span class="panel-kicker">ORDER DELIVERY</span><h2>订单与发货趋势</h2></div>
          <el-button link type="primary" @click="openOrders">查看订单 <el-icon><ArrowRight /></el-icon></el-button>
        </header>
        <Chart :options="orderChartOptions" height="280px" />
      </article>

      <article class="panel event-panel">
        <header class="panel-header">
          <div><span class="panel-kicker">LATEST EVENTS</span><h2>最近业务事件</h2></div>
          <el-button link type="primary" @click="openEvents">查看全部 <el-icon><ArrowRight /></el-icon></el-button>
        </header>
        <el-empty v-if="!overview.recentEvents.length" description="所选时间内暂无业务事件" :image-size="72" />
        <div v-else class="event-list">
          <div v-for="event in overview.recentEvents" :key="event.ID" class="event-row">
            <span class="event-status" :class="eventStatusClass(event.status)" />
            <div class="event-content">
              <div><strong>{{ event.summary }}</strong><el-tag size="small" effect="plain" :type="eventTagType(event.status)">{{ eventStatusText(event.status) }}</el-tag></div>
              <p>{{ event.accountLabel || '全部账号' }} · {{ eventTypeText(event.type) }}<template v-if="event.error"> · {{ event.error }}</template></p>
            </div>
            <time>{{ formatEventTime(event.CreatedAt) }}</time>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import Chart from '@/components/charts/index.vue'
  import { getXianyuAccountList } from '@/plugin/xianyu/api/account'
  import { getXianyuDashboardOverview } from '@/plugin/xianyu/api/status'

  defineOptions({ name: 'XianyuDashboard' })

  const router = useRouter()
  const appStore = useAppStore()
  const loading = ref(false)
  const accounts = ref([])
  const accountId = ref()
  const datePreset = ref('7d')
  const dateRange = ref([])
  const updatedAt = ref('')
  const overview = ref(emptyOverview())

  const serviceHealthy = computed(() => overview.value.metrics.accounts === 0 || overview.value.metrics.readyAccounts === overview.value.metrics.accounts)
  const metricCards = computed(() => {
    const metrics = overview.value.metrics
    return [
      { label: '账号在线', value: `${metrics.readyAccounts || 0}/${metrics.accounts || 0}`, note: '当前登录并可连接', tone: serviceHealthy.value ? 'blue' : 'amber', route: 'XianyuAccounts' },
      { label: '在售商品', value: formatNumber(metrics.onSaleProducts), note: '当前在售库存', tone: 'cyan', route: 'XianyuProducts' },
      { label: '收到消息', value: formatNumber(metrics.inboundMessages), note: `同时发出 ${formatNumber(metrics.outboundMessages)} 条`, tone: 'blue', route: 'XianyuMessages' },
      { label: '自动回复率', value: `${Number(metrics.autoReplyRate || 0).toFixed(1)}%`, note: `${formatNumber(metrics.autoReplies)} 条自动回复`, tone: 'green', route: 'XianyuMessages' },
      { label: '新增订单', value: formatNumber(metrics.newOrders), note: `${formatNumber(metrics.pendingOrders)} 笔待发货`, tone: 'violet', route: 'XianyuOrders' },
      { label: '自动化异常', value: formatNumber(metrics.failedAutomationTasks), note: '回复、发货与续售', tone: metrics.failedAutomationTasks ? 'red' : 'green', route: 'XianyuBusinessEvents' }
    ]
  })
  const automationRows = computed(() => {
    const automation = overview.value.automation
    return [automationRow('自动回复', automation.replySuccess, automation.replyFailed), automationRow('自动发货', automation.deliverySuccess, automation.deliveryFailed), automationRow('自动续售', automation.relistSuccess, automation.relistFailed)]
  })
  const chartTextColor = computed(() => appStore.isDark ? '#94a3b8' : '#64748b')
  const chartGridColor = computed(() => appStore.isDark ? '#263449' : '#e8edf4')
  const messageChartOptions = computed(() => ({
    color: ['#2563eb', '#0f9f8f', '#f59e0b'],
    tooltip: { trigger: 'axis', backgroundColor: appStore.isDark ? '#111827' : '#ffffff', borderColor: chartGridColor.value, textStyle: { color: appStore.isDark ? '#e5e7eb' : '#1f2937' } },
    legend: { top: 2, right: 8, textStyle: { color: chartTextColor.value } },
    grid: { left: 42, right: 18, top: 46, bottom: 36 },
    xAxis: { type: 'category', boundaryGap: false, data: overview.value.messageTrend.map(point => formatBucket(point.timestamp)), axisLabel: { color: chartTextColor.value }, axisLine: { lineStyle: { color: chartGridColor.value } }, axisTick: { show: false } },
    yAxis: { type: 'value', minInterval: 1, axisLabel: { color: chartTextColor.value }, splitLine: { lineStyle: { color: chartGridColor.value, type: 'dashed' } } },
    series: [lineSeries('收到消息', overview.value.messageTrend.map(point => point.inbound), '#2563eb', true), lineSeries('发出消息', overview.value.messageTrend.map(point => point.outbound), '#0f9f8f'), lineSeries('自动回复', overview.value.messageTrend.map(point => point.autoReplies), '#f59e0b')]
  }))
  const messageSourceOptions = computed(() => ({
    color: ['#2563eb', '#0f9f8f', '#94a3b8'], tooltip: { trigger: 'item', formatter: '{b}<br/>{c} 条 · {d}%' },
    series: [{ type: 'pie', radius: ['58%', '80%'], center: ['50%', '50%'], avoidLabelOverlap: true, itemStyle: { borderColor: appStore.isDark ? '#152033' : '#fff', borderWidth: 4, borderRadius: 8 }, label: { show: false }, emphasis: { scaleSize: 8 }, data: overview.value.messageSources }]
  }))
  const orderChartOptions = computed(() => ({
    color: ['#2563eb', '#0f9f8f', '#ef4444'], tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } }, legend: { top: 0, right: 8, textStyle: { color: chartTextColor.value } }, grid: { left: 42, right: 18, top: 42, bottom: 34 },
    xAxis: { type: 'category', data: overview.value.orderTrend.map(point => formatBucket(point.timestamp)), axisLabel: { color: chartTextColor.value }, axisLine: { lineStyle: { color: chartGridColor.value } }, axisTick: { show: false } },
    yAxis: { type: 'value', minInterval: 1, axisLabel: { color: chartTextColor.value }, splitLine: { lineStyle: { color: chartGridColor.value, type: 'dashed' } } },
    series: [{ name: '新增订单', type: 'bar', barMaxWidth: 20, data: overview.value.orderTrend.map(point => point.orders), itemStyle: { borderRadius: [5, 5, 0, 0] } }, { name: '发货成功', type: 'line', smooth: true, symbolSize: 7, data: overview.value.orderTrend.map(point => point.delivered) }, { name: '发货异常', type: 'line', smooth: true, symbolSize: 7, data: overview.value.orderTrend.map(point => point.failed) }]
  }))

  function emptyOverview() { return { metrics: {}, messageTrend: [], messageSources: [], orderTrend: [], automation: {}, recentEvents: [] } }
  function formatNumber(value) { return Number(value || 0).toLocaleString('zh-CN') }
  function lineSeries(name, data, color, area = false) { return { name, type: 'line', data, smooth: 0.35, showSymbol: false, symbolSize: 8, lineStyle: { width: 3 }, areaStyle: area ? { color: `${color}18` } : undefined, emphasis: { focus: 'series' } } }
  function automationRow(label, success = 0, failed = 0) { const total = Number(success) + Number(failed); return { label, success, failed, rate: total ? Math.round(Number(success) * 100 / total) : 0 } }
  function startOfDay(date) { const value = new Date(date); value.setHours(0, 0, 0, 0); return value.getTime() }
  function setPresetRange(preset) { const now = Date.now(); const days = preset === '30d' ? 29 : preset === '7d' ? 6 : 0; dateRange.value = [startOfDay(now - days * 86400000), now + 1000] }
  function applyPreset(preset) { if (preset !== 'custom') { setPresetRange(preset); loadOverview() } }
  function formatBucket(timestamp) { const date = new Date(timestamp); return overview.value.bucketMillis <= 3600000 ? date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false }) : `${date.getMonth() + 1}/${date.getDate()}` }
  function formatEventTime(value) { return new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }) }
  function eventStatusClass(status) { return status === 'success' || status === 'ready' ? 'is-success' : status === 'unknown' ? 'is-warning' : 'is-danger' }
  function eventTagType(status) { return status === 'success' || status === 'ready' ? 'success' : status === 'unknown' ? 'warning' : 'danger' }
  function eventStatusText(status) { return status === 'success' || status === 'ready' ? '成功' : status === 'unknown' ? '待确认' : '异常' }
  function eventTypeText(type) { return ({ message: '消息', delivery: '自动发货', relist: '自动续售', login_check: '登录检查', websocket_alert: '消息连接', order_sync: '订单同步', product_sync: '商品同步' })[type] || type }
  function filterQuery() { return { accountId: accountId.value || undefined, startTime: dateRange.value[0], endTime: dateRange.value[1] } }
  function openMetric(name) { if (name === 'XianyuMessages') return openMessages(); if (name === 'XianyuBusinessEvents') return openEvents(); router.push({ name }) }
  function openMessages() { router.push({ name: 'XianyuMessages', query: filterQuery() }) }
  function openOrders() { router.push({ name: 'XianyuOrders', query: { accountId: accountId.value || undefined } }) }
  function openEvents() { router.push({ name: 'XianyuBusinessEvents', query: { accountId: accountId.value || undefined } }) }
  async function loadOverview() {
    if (!dateRange.value?.length) return
    loading.value = true
    try {
      const response = await getXianyuDashboardOverview(filterQuery())
      if (response.code === 0) { overview.value = { ...emptyOverview(), ...response.data }; updatedAt.value = new Date().toLocaleTimeString('zh-CN', { hour12: false }) }
    } catch { ElMessage.error('工作台数据加载失败') } finally { loading.value = false }
  }
  onMounted(async () => {
    setPresetRange(datePreset.value)
    const accountResponse = await getXianyuAccountList({ page: 1, pageSize: 100 })
    if (accountResponse.code === 0) accounts.value = accountResponse.data.list || []
    loadOverview()
  })
</script>

<style scoped>
  .workspace { min-height: 100%; padding: 22px; color: var(--el-text-color-primary); background: radial-gradient(circle at 95% 0, rgb(37 99 235 / 8%), transparent 28%), var(--el-bg-color-page); }
  .workspace-hero { display: flex; align-items: center; justify-content: space-between; gap: 24px; padding: 26px 28px; border: 1px solid var(--el-border-color-lighter); border-radius: 18px; background: var(--el-bg-color); box-shadow: 0 12px 36px rgb(15 23 42 / 6%); }
  .eyebrow, .panel-kicker { color: #2563eb; font-size: 11px; font-weight: 800; letter-spacing: .16em; }
  .workspace-hero h1 { margin: 7px 0 5px; font-size: 28px; line-height: 1.2; letter-spacing: -.03em; }
  .workspace-hero p { margin: 0; color: var(--el-text-color-secondary); font-size: 14px; }
  .service-state { display: flex; align-items: center; gap: 12px; min-width: 210px; padding: 13px 16px; border: 1px solid #dbe5f2; border-radius: 12px; background: #f8fbff; }
  .service-state.is-healthy { border-color: #bde8d9; background: #f2fbf7; } .service-state.is-warning { border-color: #fed7aa; background: #fff8ed; }
  .state-dot { width: 10px; height: 10px; border-radius: 50%; background: #f59e0b; box-shadow: 0 0 0 5px rgb(245 158 11 / 14%); } .is-healthy .state-dot { background: #10b981; box-shadow: 0 0 0 5px rgb(16 185 129 / 14%); }
  .service-state strong, .service-state small { display: block; } .service-state strong { font-size: 14px; } .service-state small { margin-top: 2px; color: var(--el-text-color-secondary); }
  .filter-bar { display: flex; align-items: center; gap: 12px; margin: 16px 0; padding: 12px; border: 1px solid var(--el-border-color-lighter); border-radius: 14px; background: var(--el-bg-color); }
  .account-select { width: 190px; } .date-picker { width: 360px; }
  .refresh-button { display: inline-flex; align-items: center; gap: 6px; height: 32px; padding: 0 14px; border: 1px solid #bfd2f5; border-radius: 8px; color: #1d4ed8; background: #eff6ff; cursor: pointer; } .refresh-button:disabled { opacity: .55; cursor: wait; }
  .updated-at { margin-left: auto; color: var(--el-text-color-secondary); font-size: 12px; }
  .metric-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 14px; }
  .metric-card { position: relative; overflow: hidden; min-height: 132px; padding: 18px; border: 1px solid var(--el-border-color-lighter); border-radius: 15px; text-align: left; color: inherit; background: var(--el-bg-color); cursor: pointer; transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease; }
  .metric-card:hover { transform: translateY(-2px); border-color: color-mix(in srgb, var(--metric-color) 42%, transparent); box-shadow: 0 14px 30px rgb(15 23 42 / 8%); }
  .metric-label { display: block; color: var(--el-text-color-secondary); font-size: 13px; } .metric-card strong { display: block; margin-top: 10px; font-size: 27px; line-height: 1; letter-spacing: -.03em; } .metric-note { display: block; margin-top: 10px; color: var(--el-text-color-placeholder); font-size: 12px; }
  .metric-line { position: absolute; right: 0; bottom: 0; left: 0; height: 3px; background: var(--metric-color); }
  .tone-blue { --metric-color: #2563eb; } .tone-cyan { --metric-color: #0891b2; } .tone-green { --metric-color: #10b981; } .tone-violet { --metric-color: #7c3aed; } .tone-amber { --metric-color: #f59e0b; } .tone-red { --metric-color: #ef4444; }
  .dashboard-grid { display: grid; grid-template-columns: repeat(12, minmax(0, 1fr)); gap: 16px; margin-top: 16px; }
  .panel { padding: 20px; border: 1px solid var(--el-border-color-lighter); border-radius: 16px; background: var(--el-bg-color); box-shadow: 0 8px 26px rgb(15 23 42 / 4%); }
  .message-panel { grid-column: span 8; } .source-panel { grid-column: span 4; } .automation-panel { grid-column: span 5; } .order-panel { grid-column: span 7; } .event-panel { grid-column: span 12; }
  .panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 8px; } .panel-header h2 { margin: 5px 0 0; font-size: 18px; letter-spacing: -.02em; }
  .source-summary { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-top: 4px; } .source-summary div { padding: 10px; border-radius: 9px; background: var(--el-fill-color-lighter); text-align: center; } .source-summary span, .source-summary strong { display: block; } .source-summary span { color: var(--el-text-color-secondary); font-size: 11px; } .source-summary strong { margin-top: 3px; font-size: 16px; }
  .automation-list { display: grid; gap: 26px; padding: 28px 4px 13px; } .automation-title { display: flex; align-items: center; justify-content: space-between; margin-bottom: 9px; font-size: 13px; } .automation-title > span:first-child { font-weight: 700; } .automation-title > span:last-child { color: var(--el-text-color-secondary); font-size: 12px; } .automation-title b { color: #059669; } .automation-title em { color: #dc2626; font-style: normal; font-weight: 700; }
  .progress-track { overflow: hidden; height: 8px; border-radius: 99px; background: var(--el-fill-color); } .progress-track span { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, #2563eb, #10b981); }
  .event-list { margin-top: 12px; } .event-row { display: grid; grid-template-columns: 12px minmax(0, 1fr) auto; align-items: center; gap: 13px; padding: 13px 4px; border-top: 1px solid var(--el-border-color-lighter); } .event-row:first-child { border-top: 0; }
  .event-status { width: 9px; height: 9px; border-radius: 50%; } .event-status.is-success { background: #10b981; } .event-status.is-warning { background: #f59e0b; } .event-status.is-danger { background: #ef4444; }
  .event-content > div { display: flex; align-items: center; gap: 8px; } .event-content strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; } .event-content p { overflow: hidden; margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; } .event-row time { color: var(--el-text-color-placeholder); font-size: 12px; }
  :global(.dark) .workspace { background: radial-gradient(circle at 95% 0, rgb(37 99 235 / 14%), transparent 28%), var(--el-bg-color-page); } :global(.dark) .service-state { color: #dbeafe; background: #101c2f; } :global(.dark) .service-state.is-healthy { background: #0d241f; }
  @media (max-width: 1280px) { .metric-grid { grid-template-columns: repeat(3, 1fr); } }
  @media (max-width: 960px) { .workspace-hero { align-items: flex-start; flex-direction: column; } .filter-bar { align-items: flex-start; flex-wrap: wrap; } .updated-at { margin-left: 0; width: 100%; } .message-panel, .source-panel, .automation-panel, .order-panel { grid-column: span 12; } }
  @media (max-width: 640px) { .workspace { padding: 12px; } .metric-grid { grid-template-columns: repeat(2, 1fr); } .workspace-hero { padding: 20px; } .service-state, .account-select, .date-picker { width: 100%; } .metric-card { min-height: 118px; padding: 15px; } .event-row { grid-template-columns: 10px minmax(0, 1fr); } .event-row time { display: none; } }
</style>
