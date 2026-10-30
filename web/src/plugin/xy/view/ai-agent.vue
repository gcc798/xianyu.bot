<template>
  <div class="agent-workbench" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <aside class="session-rail" aria-label="对话列表">
      <div class="rail-head">
        <div v-if="!sidebarCollapsed" class="brand-lockup"><span>XY</span><div><strong>XyAgent</strong><small>RESEARCH DESK</small></div></div>
        <button class="icon-button" :aria-label="sidebarCollapsed ? '展开会话列表' : '收起会话列表'" @click="sidebarCollapsed = !sidebarCollapsed">{{ sidebarCollapsed ? '›' : '‹' }}</button>
      </div>
      <el-button v-if="!sidebarCollapsed" class="new-session" @click="startNewSession"><span>＋</span> 新对话</el-button>
      <div v-if="!sidebarCollapsed" class="session-list">
        <button v-for="session in sessions" :key="session.ID" class="session-item" :class="{ active: session.ID === activeSessionId }" @click="selectSession(session.ID)">
          <span class="session-signal" />
          <span class="session-copy"><strong>{{ session.title || '新对话' }}</strong><small>{{ formatTime(session.lastMessageAt || session.createdAt) }}</small></span>
          <span class="session-delete" title="删除会话" @click.stop="removeSession(session)">×</span>
        </button>
        <div v-if="!sessions.length && !loadingSessions" class="rail-empty">暂无对话记录</div>
      </div>
    </aside>

    <main class="conversation-stage">
      <header class="conversation-head">
        <div><span class="status-lamp" :class="{ busy: sending }" /><strong>{{ activeTitle }}</strong><small>{{ sending ? 'Agent 正在执行' : 'READY' }}</small></div>
        <div class="head-actions"><button v-if="sending" class="cancel-button" @click="cancelRun">停止运行</button><button class="run-button" :disabled="!runEvents.length" @click="detailsOpen = true"><span>≡</span> 本轮详情</button></div>
      </header>

      <section ref="messageViewport" class="message-viewport" aria-live="polite">
        <div v-if="!messages.length" class="welcome-panel">
          <p class="kicker">XY / AGENT CONSOLE</p>
          <h1>把目标交代清楚，<br><em>其余交给能力。</em></h1>
          <p>直接对话，或输入 <kbd>/</kbd> 精确指定本轮 Skill 与 Tool。写操作始终需要你的确认。</p>
          <div class="starter-grid">
            <button @click="useStarter('帮我检索三份 Go 并发编程的 PDF，并比较来源和许可证信息')"><span>01</span>检索并比较技术资料</button>
            <button @click="useStarter('列出当前可用的资源来源，并说明适合检索什么内容')"><span>02</span>查看可用资源来源</button>
            <button @click="useStarter('检查当前 MCP 能力状态，只报告不可用的能力')"><span>03</span>检查能力状态</button>
          </div>
        </div>

        <template v-for="message in messages" :key="message.localId || message.ID">
          <article v-if="message.role === 'user' || message.role === 'assistant'" class="message-row" :class="message.role">
            <div class="message-avatar">{{ message.role === 'user' ? 'YOU' : 'XY' }}</div>
            <div class="message-body">
              <div class="message-meta"><strong>{{ message.role === 'user' ? '你' : 'XyAgent' }}</strong><span>{{ formatTime(message.createdAt) }}</span></div>
              <div class="message-content" :class="{ streaming: message.streaming }">{{ message.content }}<span v-if="message.streaming" class="cursor" /></div>
            </div>
          </article>
          <article v-else-if="message.role === 'tool'" class="event-card tool-card">
            <span class="event-icon">⌘</span><div><small>TOOL</small><strong>{{ message.tool }}</strong></div><el-tag size="small" :type="message.status === 'failed' ? 'danger' : message.status === 'completed' ? 'success' : 'warning'">{{ toolStatus(message.status) }}</el-tag>
          </article>
          <article v-else-if="message.role === 'approval'" class="event-card approval-card">
            <span class="event-icon">!</span><div><small>CONFIRMATION</small><strong>{{ message.tool }} 等待确认</strong><p>该写操作尚未执行。</p></div><el-button size="small" type="warning" :loading="message.approving" @click="approveRun(message)">确认执行</el-button>
          </article>
          <article v-else-if="message.role === 'skill'" class="event-card skill-card">
            <span class="event-icon">◇</span><div><small>SKILL</small><strong>{{ message.skill }}</strong></div><el-tag size="small" type="success">已加载</el-tag>
          </article>
          <article v-else-if="message.role === 'task'" class="event-card task-card">
            <span class="event-icon">↧</span><div><small>RESOURCE TASK #{{ message.taskId }}</small><strong>{{ taskStatus(message.state) }}</strong><p>{{ message.stage || '等待处理' }} · {{ message.progressPercent || 0 }}%</p></div><el-progress type="circle" :width="38" :stroke-width="4" :percentage="message.progressPercent || 0" :status="message.state === 'failed' ? 'exception' : message.state === 'completed' ? 'success' : undefined" />
          </article>
        </template>
      </section>

      <footer class="composer-zone">
        <div v-if="directives.length" class="directive-strip">
          <span v-for="directive in directives" :key="directive.type" class="directive-chip" :class="directive.type">
            <b>{{ directive.type === 'skill' ? '◇ SKILL' : '⌘ TOOL' }}</b>{{ directive.name }}
            <button :aria-label="`移除 ${directive.name}`" @click="removeDirective(directive.type)">×</button>
          </span>
        </div>

        <div v-if="slashOpen" class="slash-menu" role="listbox">
          <div class="slash-head"><span>能力选择器</span><small>↑↓ 选择 · Enter 确认 · Esc 关闭</small></div>
          <button v-for="(capability, index) in filteredCapabilities" :key="`${capability.type}-${capability.name}`" class="slash-option" :class="[capability.type, { active: index === slashIndex }]" @mouseenter="slashIndex = index" @mousedown.prevent="selectCapability(capability)">
            <span class="capability-icon">{{ capabilityIcon(capability.type) }}</span>
            <span><strong>{{ capability.name }}</strong><small>{{ capability.description }}</small></span>
            <em>{{ capability.type.toUpperCase() }}</em>
          </button>
          <div v-if="!filteredCapabilities.length" class="slash-empty">没有匹配的能力</div>
        </div>

        <div class="composer">
          <textarea ref="composerInput" v-model="draft" rows="1" :disabled="sending" aria-label="发送消息给 XyAgent" placeholder="描述目标，或输入 / 选择 Skill、Tool…" @input="updateSlashMenu" @keydown="handleComposerKeydown" />
          <div class="composer-foot">
            <div class="composer-tools"><button class="model-pill" type="button"><span>◉</span> 默认模型</button><span>Enter 发送 · Shift+Enter 换行</span></div>
            <button class="send-button" :disabled="!canSend" aria-label="发送消息" @click="sendMessage">{{ sending ? '执行中' : '发送 ↗' }}</button>
          </div>
        </div>
      </footer>
    </main>

    <el-drawer v-model="detailsOpen" title="本轮运行详情" size="390px" class="run-drawer">
      <div class="run-summary"><span class="status-lamp" :class="{ busy: sending }" /><div><small>RUN STATUS</small><strong>{{ sending ? '运行中' : '已结束' }}</strong></div><code v-if="activeRunId">#{{ activeRunId }}</code></div>
      <el-timeline class="run-timeline">
        <el-timeline-item v-for="event in runEvents" :key="event.id" :timestamp="formatTime(event.at)" placement="top" :type="eventColor(event.type)">
          <strong>{{ eventLabel(event) }}</strong><p>{{ eventDescription(event) }}</p>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-if="!runEvents.length" description="发送消息后显示运行事件" />
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { approveAgentRun, cancelAgentRun, createAgentSession, deleteAgentSession, getAgentSession, getAgentSessions, sendAgentMessage, streamAgentEvents } from '@/plugin/xy/api/agent'

defineOptions({ name: 'XyAgent' })

const capabilities = [
  { type: 'skill', name: 'resource-discovery', description: '提取检索目标，先比较候选，获取前请求确认' },
  { type: 'skill', name: 'telegram-resource', description: '使用已扫描来源检索，并明确许可证未知风险' },
  { type: 'tool', name: 'channel_get_info', description: '读取能力范围与文件限制' },
  { type: 'tool', name: 'channel_get_status', description: '检查 MCP 能力是否可用' },
  { type: 'tool', name: 'source_list', description: '列出当前可检索的来源' },
  { type: 'tool', name: 'source_sync', description: '同步来源，执行前需要确认' },
  { type: 'tool', name: 'resource_search', description: '按条件搜索资源候选' },
  { type: 'tool', name: 'resource_inspect', description: '查看候选文件与来源详情' },
  { type: 'tool', name: 'resource_acquire', description: '获取资源，执行前需要确认' },
  { type: 'tool', name: 'task_get', description: '查询异步任务进度' },
  { type: 'tool', name: 'task_cancel', description: '取消任务，执行前需要确认' },
  { type: 'command', name: 'new', description: '创建一条新对话' },
  { type: 'command', name: 'clear', description: '清空当前输入与已选能力' },
  { type: 'model', name: 'default', description: '使用基础配置中的默认模型' }
]

const sessions = ref([])
const activeSessionId = ref(0)
const messages = ref([])
const draft = ref('')
const directives = ref([])
const loadingSessions = ref(false)
const sending = ref(false)
const sidebarCollapsed = ref(false)
const detailsOpen = ref(false)
const slashOpen = ref(false)
const slashIndex = ref(0)
const slashQuery = ref('')
const slashRange = ref(null)
const activeRunId = ref(0)
const runEvents = ref([])
const watchingTask = ref(false)
const composerInput = ref(null)
const messageViewport = ref(null)
let streamController = null
let reconnectTimer = null
let lastEventId = 0

const activeTitle = computed(() => sessions.value.find(item => item.ID === activeSessionId.value)?.title || '新对话')
const canSend = computed(() => draft.value.trim() && !sending.value)
const filteredCapabilities = computed(() => {
  const raw = slashQuery.value.trim().toLowerCase()
  const [prefix = '', ...rest] = raw.split(/\s+/)
  const type = ['skill', 'tool', 'command', 'model'].includes(prefix) ? prefix : ''
  const keyword = type ? rest.join(' ') : raw
  return capabilities.filter(item => (!type || item.type === type) && (!keyword || `${item.name} ${item.description}`.toLowerCase().includes(keyword)))
})

const loadSessions = async () => {
  loadingSessions.value = true
  try {
    const res = await getAgentSessions({ page: 1, pageSize: 100 })
    if (res.code === 0) sessions.value = res.data.list || []
  } finally { loadingSessions.value = false }
}

const startNewSession = async () => {
  const res = await createAgentSession({ title: '新对话' })
  if (res.code !== 0) return
  await loadSessions()
  await selectSession(res.data.session.ID)
  nextTick(() => composerInput.value?.focus())
}

const selectSession = async (id) => {
  stopStream()
  activeSessionId.value = id
  activeRunId.value = 0
  sending.value = false
  watchingTask.value = false
  runEvents.value = []
  lastEventId = 0
  await reloadConversation(id)
  const activeTask = messages.value.find(item => item.role === 'task' && ['queued', 'running'].includes(item.state))
  if (activeTask) {
    activeRunId.value = activeTask.runId
    watchingTask.value = true
    startStream(id, activeTask.runId)
  }
}

const reloadConversation = async id => {
  const res = await getAgentSession(id)
  if (res.code === 0) {
    const session = res.data.session
    messages.value = [
      ...(session.messages || []),
      ...(session.pendingApprovals || []).map(item => ({
        localId: `approval-${item.runId}-${item.toolName}-${item.createdAt}`, role: 'approval', runId: item.runId, tool: item.toolName, createdAt: item.createdAt
      })),
      ...(session.tasks || []).map(item => ({ localId: `task-${item.id}`, role: 'task', taskId: item.id, ...item }))
    ]
  }
  scrollToBottom()
}

const removeSession = async (session) => {
  try { await ElMessageBox.confirm(`删除“${session.title || '新对话'}”及其消息记录？`, '删除对话', { type: 'warning' }) } catch { return }
  const res = await deleteAgentSession(session.ID)
  if (res.code !== 0) return
  if (activeSessionId.value === session.ID) { activeSessionId.value = 0; messages.value = [] }
  await loadSessions()
}

const updateSlashMenu = () => {
  const match = draft.value.match(/(?:^|\s)\/([^\n/]*)$/)
  slashOpen.value = Boolean(match)
  slashQuery.value = match?.[1] || ''
  slashRange.value = match ? { start: match.index + (match[0].startsWith(' ') ? 1 : 0), end: draft.value.length } : null
  slashIndex.value = 0
}

const selectCapability = async (capability) => {
  if (capability.type === 'command') {
    if (capability.name === 'new') await startNewSession()
    if (capability.name === 'clear') { draft.value = ''; directives.value = [] }
  } else if (capability.type === 'skill' || capability.type === 'tool') {
    directives.value = [...directives.value.filter(item => item.type !== capability.type), { type: capability.type, name: capability.name }]
  }
  if (slashRange.value) draft.value = `${draft.value.slice(0, slashRange.value.start)}${draft.value.slice(slashRange.value.end)}`.trimStart()
  slashOpen.value = false
  await nextTick()
  composerInput.value?.focus()
}

const handleComposerKeydown = (event) => {
  if (slashOpen.value) {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      const step = event.key === 'ArrowDown' ? 1 : -1
      slashIndex.value = (slashIndex.value + step + filteredCapabilities.value.length) % Math.max(filteredCapabilities.value.length, 1)
      return
    }
    if (event.key === 'Enter' && filteredCapabilities.value.length) { event.preventDefault(); selectCapability(filteredCapabilities.value[slashIndex.value]); return }
    if (event.key === 'Escape') { event.preventDefault(); slashOpen.value = false; return }
  }
  if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); sendMessage() }
}

const sendMessage = async () => {
  const content = draft.value.trim()
  if (!content || sending.value) return
  if (!activeSessionId.value) await startNewSession()
  if (!activeSessionId.value) return
  const chosen = directives.value.map(item => ({ ...item }))
  const idempotencyKey = crypto.randomUUID()
  messages.value.push({ localId: `user-${idempotencyKey}`, role: 'user', content, createdAt: new Date().toISOString() })
  draft.value = ''
  directives.value = []
  slashOpen.value = false
  sending.value = true
  watchingTask.value = false
  runEvents.value = []
  scrollToBottom()
  try {
    const res = await sendAgentMessage(activeSessionId.value, { message: content, directives: chosen, idempotencyKey })
    if (res.code !== 0) throw new Error(res.msg || '消息提交失败')
    activeRunId.value = res.data.run.id
    if (res.data.run.status !== 'running') {
      sending.value = false
      await reloadConversation(activeSessionId.value)
      return
    }
    messages.value.push({ localId: `assistant-${activeRunId.value}`, runId: activeRunId.value, role: 'assistant', content: '', streaming: true, createdAt: new Date().toISOString() })
    startStream(activeSessionId.value, activeRunId.value)
  } catch (error) {
    sending.value = false
    ElMessage.error(error.message || '消息提交失败')
  }
}

const startStream = (sessionId, runId) => {
  stopStream()
  streamController = new AbortController()
  const controller = streamController
  streamAgentEvents(sessionId, { afterId: lastEventId, signal: controller.signal, onEvent: event => handleRunEvent(event, runId) })
    .then(() => { if (!controller.signal.aborted && (sending.value || watchingTask.value) && activeRunId.value === runId) reconnectTimer = setTimeout(() => startStream(sessionId, runId), 800) })
    .catch(() => { if (!controller.signal.aborted && (sending.value || watchingTask.value) && activeRunId.value === runId) reconnectTimer = setTimeout(() => startStream(sessionId, runId), 1200) })
}

const handleRunEvent = (event, runId) => {
  lastEventId = Math.max(lastEventId, event.id || 0)
  if (event.data?.runId !== runId) return
  if (watchingTask.value && !sending.value && !event.type.startsWith('task.')) return
  runEvents.value.push({ ...event, at: new Date().toISOString() })
  if (event.type === 'assistant.delta') {
    const target = messages.value.find(item => item.runId === runId && item.role === 'assistant')
    if (target) target.content += event.data.content || ''
  } else if (event.type === 'tool.started') {
    messages.value.push({ localId: `tool-${event.id}`, role: 'tool', tool: event.data.tool, status: 'running' })
  } else if (event.type === 'tool.completed') {
    const target = [...messages.value].reverse().find(item => item.role === 'tool' && item.tool === event.data.tool)
    if (target) target.status = event.data.status
  } else if (event.type === 'approval.required') {
    messages.value.push({ localId: `approval-${event.id}`, role: 'approval', runId: event.data.runId, tool: event.data.tool })
  } else if (event.type === 'skill.loaded') {
    messages.value.push({ localId: `skill-${event.id}`, role: 'skill', skill: event.data.skill })
  } else if (event.type === 'task.created' || event.type === 'task.updated' || event.type === 'task.completed') {
    let task = messages.value.find(item => item.role === 'task' && item.taskId === event.data.taskId)
    if (!task) {
      task = { localId: `task-${event.data.taskId}`, role: 'task', taskId: event.data.taskId }
      messages.value.push(task)
    }
    Object.assign(task, event.data)
    if (event.type === 'task.completed' || ['failed', 'cancelled'].includes(event.data.state)) {
      watchingTask.value = false
      stopStream()
    }
  } else if (event.type === 'run.completed' || event.type === 'run.failed' || event.type === 'run.cancelled') {
    finishRun(event.type === 'run.failed' ? event.data.message : '')
  }
  scrollToBottom()
}

const finishRun = async (failureMessage = '') => {
  const target = messages.value.find(item => item.runId === activeRunId.value && item.role === 'assistant')
  if (target) { target.streaming = false; if (!target.content && failureMessage) target.content = failureMessage }
  sending.value = false
  stopStream()
  await Promise.all([loadSessions(), reloadConversation(activeSessionId.value)])
}

const cancelRun = async () => {
  if (!activeSessionId.value || !activeRunId.value) return
  const res = await cancelAgentRun(activeSessionId.value, activeRunId.value)
  if (res.code === 0) ElMessage.info('正在停止 Agent 运行')
}

const approveRun = async message => {
  if (!message.runId || message.approving) return
  message.approving = true
  try {
    const res = await approveAgentRun(message.runId)
    if (res.code !== 0) return
    ElMessage.success('操作已确认并提交')
    await reloadConversation(activeSessionId.value)
    if (message.tool === 'resource_acquire') {
      activeRunId.value = message.runId
      watchingTask.value = true
      startStream(activeSessionId.value, message.runId)
    }
  } finally { message.approving = false }
}

const stopStream = () => { if (reconnectTimer) clearTimeout(reconnectTimer); reconnectTimer = null; streamController?.abort(); streamController = null }
const removeDirective = type => { directives.value = directives.value.filter(item => item.type !== type) }
const useStarter = text => { draft.value = text; nextTick(() => composerInput.value?.focus()) }
const scrollToBottom = () => nextTick(() => { if (messageViewport.value) messageViewport.value.scrollTop = messageViewport.value.scrollHeight })
const capabilityIcon = type => ({ skill: '◇', tool: '⌘', command: '›_', model: '◉' }[type])
const toolStatus = status => ({ running: '执行中', completed: '已完成', failed: '不可用' }[status] || status)
const taskStatus = status => ({ queued: '等待处理', running: '正在获取资源', completed: '资源已就绪', failed: '资源获取失败', cancelled: '任务已取消' }[status] || status)
const formatTime = value => value ? new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value)) : ''
const eventColor = type => type.includes('failed') ? 'danger' : type.includes('cancelled') || type.includes('tool') ? 'warning' : type.includes('completed') ? 'success' : 'primary'
const eventLabel = event => ({ 'run.started': '开始运行', 'assistant.delta': '生成回答', 'skill.loaded': '加载 Skill', 'tool.started': '调用 Tool', 'tool.completed': 'Tool 返回', 'approval.required': '等待确认', 'task.created': '资源任务已创建', 'task.updated': '资源任务更新', 'task.completed': '资源已就绪', 'run.completed': '运行完成', 'run.failed': '运行失败', 'run.cancelled': '运行已取消' }[event.type] || event.type)
const eventDescription = event => event.data?.tool || event.data?.skill || event.data?.message || `事件 #${event.id}`

onMounted(async () => { await loadSessions(); if (sessions.value.length) await selectSession(sessions.value[0].ID) })
onBeforeUnmount(stopStream)
</script>

<style scoped>
.agent-workbench { --agent-ink: #18312f; --agent-teal: #0c887e; --agent-orange: #d56b2d; --agent-paper: color-mix(in srgb, var(--el-bg-color) 94%, #dce9df); height: calc(100vh - 112px); min-height: 650px; display: grid; grid-template-columns: 268px minmax(0, 1fr); overflow: hidden; border: 1px solid var(--el-border-color-light); border-radius: 14px; color: var(--el-text-color-primary); background: radial-gradient(circle at 72% 8%, color-mix(in srgb, var(--agent-teal) 7%, transparent), transparent 28%), var(--agent-paper); box-shadow: 0 22px 60px rgba(25, 49, 47, .08); }
.agent-workbench.sidebar-collapsed { grid-template-columns: 58px minmax(0, 1fr); }
.session-rail { min-width: 0; padding: 14px 12px; border-right: 1px solid var(--el-border-color-lighter); background: color-mix(in srgb, var(--el-fill-color-light) 72%, transparent); }
.rail-head { height: 48px; display: flex; align-items: center; justify-content: space-between; }.brand-lockup { display: flex; align-items: center; gap: 10px; }.brand-lockup > span { width: 34px; height: 34px; display: grid; place-items: center; border: 1px solid var(--agent-teal); border-radius: 9px; color: var(--agent-teal); font: 800 12px/1 ui-monospace, monospace; }.brand-lockup div { display: grid; }.brand-lockup strong { font-size: 14px; }.brand-lockup small { margin-top: 2px; color: var(--el-text-color-secondary); font: 600 8px/1 ui-monospace, monospace; letter-spacing: .15em; }
button { font: inherit; }.icon-button, .run-button { border: 0; color: var(--el-text-color-secondary); background: transparent; cursor: pointer; }.icon-button { width: 30px; height: 30px; border-radius: 8px; font-size: 23px; }.icon-button:hover { background: var(--el-fill-color); color: var(--agent-teal); }.new-session { width: 100%; margin: 12px 0 16px; --el-button-bg-color: var(--agent-ink); --el-button-border-color: var(--agent-ink); --el-button-text-color: white; }
.session-list { display: grid; gap: 5px; max-height: calc(100% - 88px); overflow: auto; }.session-item { width: 100%; display: grid; grid-template-columns: 8px 1fr 22px; align-items: center; gap: 7px; padding: 10px 7px; border: 0; border-radius: 9px; text-align: left; color: inherit; background: transparent; cursor: pointer; }.session-item:hover, .session-item.active { background: var(--el-bg-color); box-shadow: inset 0 0 0 1px var(--el-border-color-lighter); }.session-signal { width: 5px; height: 5px; border-radius: 50%; background: var(--el-border-color); }.session-item.active .session-signal { background: var(--agent-teal); box-shadow: 0 0 0 3px color-mix(in srgb, var(--agent-teal) 15%, transparent); }.session-copy { min-width: 0; display: grid; }.session-copy strong { overflow: hidden; color: var(--el-text-color-primary); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }.session-copy small { margin-top: 4px; color: var(--el-text-color-placeholder); font-size: 10px; }.session-delete { opacity: 0; color: var(--el-text-color-placeholder); font-size: 18px; text-align: center; }.session-item:hover .session-delete { opacity: 1; }.session-delete:hover { color: var(--el-color-danger); }.rail-empty { padding: 22px 8px; color: var(--el-text-color-placeholder); font-size: 12px; text-align: center; }
.conversation-stage { min-width: 0; display: grid; grid-template-rows: 58px minmax(0, 1fr) auto; }.conversation-head { display: flex; align-items: center; justify-content: space-between; padding: 0 24px; border-bottom: 1px solid var(--el-border-color-lighter); }.conversation-head > div { display: flex; align-items: center; gap: 9px; }.conversation-head strong { font-size: 14px; }.conversation-head small { color: var(--el-text-color-placeholder); font: 600 9px/1 ui-monospace, monospace; letter-spacing: .12em; }.status-lamp { width: 7px; height: 7px; border-radius: 50%; background: var(--agent-teal); box-shadow: 0 0 0 4px color-mix(in srgb, var(--agent-teal) 13%, transparent); }.status-lamp.busy { background: var(--agent-orange); animation: pulse 1.3s infinite; }.head-actions { display: flex; align-items: center; gap: 7px; }.run-button, .cancel-button { padding: 7px 10px; border: 0; border-radius: 8px; font-size: 12px; cursor: pointer; }.run-button { color: var(--el-text-color-secondary); background: transparent; }.cancel-button { color: var(--el-color-danger); background: var(--el-color-danger-light-9); }.run-button:not(:disabled):hover { color: var(--agent-teal); background: var(--el-fill-color-light); }.run-button:disabled { opacity: .4; cursor: default; }
.message-viewport { overflow-y: auto; padding: 34px max(24px, calc((100% - 820px) / 2)); scroll-behavior: smooth; }.welcome-panel { max-width: 760px; margin: 8vh auto 0; }.kicker { margin: 0 0 18px; color: var(--agent-teal); font: 700 10px/1 ui-monospace, monospace; letter-spacing: .22em; }.welcome-panel h1 { margin: 0; color: var(--el-text-color-primary); font: 600 clamp(34px, 5vw, 62px)/1.06 Georgia, 'Times New Roman', serif; letter-spacing: -.035em; }.welcome-panel h1 em { color: var(--agent-teal); font-weight: 400; }.welcome-panel > p:not(.kicker) { max-width: 610px; margin: 22px 0 30px; color: var(--el-text-color-secondary); font-size: 14px; line-height: 1.8; }.welcome-panel kbd { padding: 2px 7px; border: 1px solid var(--el-border-color); border-radius: 5px; background: var(--el-bg-color); font-family: ui-monospace, monospace; }.starter-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }.starter-grid button { min-height: 92px; padding: 14px; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; text-align: left; color: var(--el-text-color-regular); background: color-mix(in srgb, var(--el-bg-color) 82%, transparent); cursor: pointer; transition: .2s ease; }.starter-grid button:hover { transform: translateY(-2px); border-color: color-mix(in srgb, var(--agent-teal) 45%, var(--el-border-color)); box-shadow: 0 10px 25px rgba(25, 49, 47, .08); }.starter-grid span { display: block; margin-bottom: 14px; color: var(--agent-orange); font: 700 10px/1 ui-monospace, monospace; }
.message-row { display: grid; grid-template-columns: 38px minmax(0, 1fr); gap: 12px; max-width: 820px; margin: 0 auto 28px; }.message-avatar { width: 34px; height: 34px; display: grid; place-items: center; border: 1px solid var(--el-border-color); border-radius: 9px; color: var(--el-text-color-secondary); font: 700 9px/1 ui-monospace, monospace; }.message-row.assistant .message-avatar { border-color: color-mix(in srgb, var(--agent-teal) 50%, var(--el-border-color)); color: var(--agent-teal); }.message-meta { display: flex; align-items: baseline; gap: 9px; margin: 1px 0 8px; }.message-meta strong { font-size: 13px; }.message-meta span { color: var(--el-text-color-placeholder); font-size: 10px; }.message-content { color: var(--el-text-color-primary); font-size: 14px; line-height: 1.8; white-space: pre-wrap; overflow-wrap: anywhere; }.message-row.user .message-content { display: inline-block; padding: 10px 14px; border-radius: 3px 13px 13px 13px; background: var(--el-fill-color-light); }.cursor { display: inline-block; width: 7px; height: 16px; margin-left: 3px; vertical-align: -3px; background: var(--agent-teal); animation: blink .8s steps(1) infinite; }
.event-card { max-width: 740px; min-height: 54px; display: grid; grid-template-columns: 34px 1fr auto; align-items: center; gap: 12px; margin: -10px auto 22px 50px; padding: 9px 12px; border: 1px solid var(--el-border-color-lighter); border-left: 3px solid var(--agent-orange); border-radius: 8px; background: var(--el-bg-color); }.event-card.skill-card { border-left-color: var(--agent-teal); }.event-card.approval-card { border-left-color: #c58b21; }.event-card.task-card { border-left-color: #4b5f9c; }.event-icon { width: 30px; height: 30px; display: grid; place-items: center; border-radius: 7px; color: var(--agent-orange); background: color-mix(in srgb, var(--agent-orange) 10%, transparent); font: 700 13px/1 ui-monospace, monospace; }.skill-card .event-icon { color: var(--agent-teal); background: color-mix(in srgb, var(--agent-teal) 10%, transparent); }.task-card .event-icon { color: #4b5f9c; background: color-mix(in srgb, #4b5f9c 10%, transparent); }.event-card div { display: grid; }.event-card small { color: var(--el-text-color-placeholder); font: 700 8px/1 ui-monospace, monospace; letter-spacing: .14em; }.event-card strong { margin-top: 4px; font-size: 12px; }.event-card p { margin: 3px 0 0; color: var(--el-text-color-secondary); font-size: 11px; }
.composer-zone { position: relative; padding: 10px max(20px, calc((100% - 860px) / 2)) 18px; background: linear-gradient(to top, var(--agent-paper) 78%, transparent); }.composer { padding: 11px 12px 9px; border: 1px solid var(--el-border-color); border-radius: 13px; background: var(--el-bg-color); box-shadow: 0 13px 35px rgba(25, 49, 47, .1); }.composer:focus-within { border-color: color-mix(in srgb, var(--agent-teal) 60%, var(--el-border-color)); box-shadow: 0 13px 35px rgba(12, 136, 126, .12); }.composer textarea { width: 100%; min-height: 52px; max-height: 150px; resize: none; padding: 4px; border: 0; outline: 0; color: var(--el-text-color-primary); background: transparent; font: 14px/1.65 inherit; }.composer textarea::placeholder { color: var(--el-text-color-placeholder); }.composer-foot { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.composer-tools { display: flex; align-items: center; gap: 12px; color: var(--el-text-color-placeholder); font-size: 10px; }.model-pill { padding: 5px 8px; border: 1px solid var(--el-border-color-lighter); border-radius: 7px; color: #4b5f9c; background: var(--el-fill-color-light); }.send-button { min-width: 82px; padding: 8px 14px; border: 0; border-radius: 8px; color: white; background: var(--agent-ink); cursor: pointer; }.send-button:disabled { opacity: .45; cursor: default; }
.directive-strip { display: flex; flex-wrap: wrap; gap: 7px; margin-bottom: 7px; }.directive-chip { display: inline-flex; align-items: center; gap: 7px; padding: 6px 8px; border: 1px solid; border-radius: 7px; font-size: 11px; background: var(--el-bg-color); }.directive-chip b { font: 700 8px/1 ui-monospace, monospace; letter-spacing: .08em; }.directive-chip.skill { border-color: color-mix(in srgb, var(--agent-teal) 45%, var(--el-border-color)); color: var(--agent-teal); }.directive-chip.tool { border-color: color-mix(in srgb, var(--agent-orange) 50%, var(--el-border-color)); color: var(--agent-orange); }.directive-chip button { border: 0; color: inherit; background: transparent; cursor: pointer; }
.slash-menu { position: absolute; z-index: 20; right: max(20px, calc((100% - 860px) / 2)); bottom: calc(100% - 4px); left: max(20px, calc((100% - 860px) / 2)); max-height: 370px; overflow-y: auto; padding: 8px; border: 1px solid var(--el-border-color); border-radius: 12px; background: color-mix(in srgb, var(--el-bg-color) 96%, transparent); box-shadow: 0 20px 55px rgba(20, 38, 36, .18); backdrop-filter: blur(16px); }.slash-head { display: flex; justify-content: space-between; padding: 7px 9px 10px; color: var(--el-text-color-secondary); font-size: 11px; }.slash-head small { color: var(--el-text-color-placeholder); }.slash-option { width: 100%; display: grid; grid-template-columns: 34px 1fr auto; align-items: center; gap: 10px; padding: 9px; border: 0; border-radius: 8px; text-align: left; color: var(--el-text-color-primary); background: transparent; cursor: pointer; }.slash-option.active { background: var(--el-fill-color-light); }.capability-icon { width: 30px; height: 30px; display: grid; place-items: center; border-radius: 7px; font: 700 12px/1 ui-monospace, monospace; }.slash-option > span:nth-child(2) { display: grid; }.slash-option strong { font-size: 12px; }.slash-option small { margin-top: 3px; color: var(--el-text-color-secondary); font-size: 10px; }.slash-option em { padding: 3px 6px; border-radius: 4px; font: 700 8px/1 ui-monospace, monospace; letter-spacing: .08em; }.slash-option.skill .capability-icon, .slash-option.skill em { color: var(--agent-teal); background: color-mix(in srgb, var(--agent-teal) 10%, transparent); }.slash-option.tool .capability-icon, .slash-option.tool em { color: var(--agent-orange); background: color-mix(in srgb, var(--agent-orange) 11%, transparent); }.slash-option.command .capability-icon, .slash-option.command em { color: #60758e; background: color-mix(in srgb, #60758e 11%, transparent); }.slash-option.model .capability-icon, .slash-option.model em { color: #4b5f9c; background: color-mix(in srgb, #4b5f9c 11%, transparent); }.slash-empty { padding: 18px; color: var(--el-text-color-placeholder); font-size: 12px; text-align: center; }
.run-summary { display: grid; grid-template-columns: 10px 1fr auto; align-items: center; gap: 12px; padding: 15px; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-fill-color-light); }.run-summary div { display: grid; }.run-summary small { color: var(--el-text-color-placeholder); font: 700 8px/1 ui-monospace, monospace; letter-spacing: .14em; }.run-summary strong { margin-top: 5px; }.run-summary code { color: var(--agent-teal); }.run-timeline { margin-top: 28px; }.run-timeline p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 12px; }
@keyframes pulse { 50% { box-shadow: 0 0 0 7px color-mix(in srgb, var(--agent-orange) 5%, transparent); } } @keyframes blink { 50% { opacity: 0; } }
@media (max-width: 880px) { .agent-workbench { grid-template-columns: 58px minmax(0, 1fr); }.session-rail { padding-inline: 10px; }.session-rail > :not(.rail-head), .brand-lockup { display: none; }.starter-grid { grid-template-columns: 1fr; }.welcome-panel { margin-top: 3vh; } }
@media (max-width: 620px) { .agent-workbench { height: calc(100vh - 92px); border-radius: 0; }.message-viewport { padding-inline: 16px; }.conversation-head { padding-inline: 16px; }.composer-zone { padding-inline: 12px; }.slash-menu { right: 12px; left: 12px; }.composer-tools > span { display: none; }.welcome-panel h1 { font-size: 36px; }.event-card { margin-left: 0; } }
</style>
