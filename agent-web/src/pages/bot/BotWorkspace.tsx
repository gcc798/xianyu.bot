import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { approveRun, cancelRun, createSession, deleteSession, getSession, getTrajectory, listSessions, sendMessage, streamEvents } from '../../api/agent'
import type { AgentEvent, Approval, ChatMessage, Directive, Operation, Session } from '../../models/agent'
import TrajectoryPanel from './components/TrajectoryPanel'

const capabilities = [
  ['skill', 'resource-discovery', '理解目标并编排资源检索'], ['skill', 'telegram-resource', '使用当前 MCP 的资源能力'],
  ['tool', 'dialog_list', '列出可访问的群组与频道'], ['tool', 'dialog_resolve', '解析群组、频道或链接'],
  ['tool', 'message_send', '向指定会话发送检索词'], ['tool', 'message_list', '等待并读取新消息'],
  ['tool', 'message_search', '检索指定会话历史消息'], ['tool', 'resource_inspect', '查看候选资源详情'],
  ['tool', 'resource_acquire', '确认后获取候选资源'], ['tool', 'task_get', '查询异步任务进度']
] as const
const terminalEvents = new Set(['run.completed', 'run.failed', 'run.cancelled'])

export default function BotWorkspace() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [sessionId, setSessionId] = useState(0)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [operations, setOperations] = useState<Operation[]>([])
  const [approvals, setApprovals] = useState<Approval[]>([])
  const [draft, setDraft] = useState('')
  const [directives, setDirectives] = useState<Directive[]>([])
  const [running, setRunning] = useState(false)
  const [runId, setRunId] = useState(0)
  const [lastEventId, setLastEventId] = useState(0)
  const [error, setError] = useState('')
  const [slashOpen, setSlashOpen] = useState(false)
  const [operationsOpen, setOperationsOpen] = useState(false)
	const [trajectoryOpen, setTrajectoryOpen] = useState(false)
  const [historyCollapsed, setHistoryCollapsed] = useState(false)
  const streamRef = useRef<AbortController | null>(null)
  const viewportRef = useRef<HTMLDivElement>(null)

  const activeSession = sessions.find((item) => item.ID === sessionId)
  const slashQuery = draft.match(/(?:^|\s)\/([^\n/]*)$/)?.[1]?.trim().toLowerCase() || ''
  const [slashType, ...slashWords] = slashQuery.split(/\s+/)
  const slashTerm = slashType === 'skill' || slashType === 'tool' ? slashWords.join(' ') : slashQuery
  const filteredCapabilities = capabilities.filter(([type, name, description]) =>
    (!['skill', 'tool'].includes(slashType) || type === slashType) && (!slashTerm || `${name} ${description}`.toLowerCase().includes(slashTerm)))

  const refreshSessions = useCallback(async () => {
    const data = await listSessions()
    setSessions(data.list || [])
    return data.list || []
  }, [])

  const openSession = useCallback(async (id: number) => {
    streamRef.current?.abort()
    setSessionId(id); setError(''); setRunning(false); setRunId(0)
    const [{ session }, trajectory] = await Promise.all([getSession(id), getTrajectory(id)])
    setMessages(session.messages || [])
    setOperations((session.toolTraces || []).map((item) => ({
      callId: item.callId || `stored-${item.id}`, runId: item.runId, tool: item.toolName,
      status: item.status, durationMillis: item.durationMillis
    })))
    setApprovals(session.pendingApprovals || [])
    setLastEventId(trajectory.nextAfterId || 0)
  }, [])

  useEffect(() => {
    let alive = true
    refreshSessions().then((items) => { if (alive && items.length) void openSession(items[0].ID) })
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : '加载失败'))
    return () => { alive = false; streamRef.current?.abort() }
  }, [openSession, refreshSessions])

  useEffect(() => { viewportRef.current?.scrollTo({ top: viewportRef.current.scrollHeight, behavior: 'smooth' }) }, [messages, operations, approvals])

  const reloadActive = useCallback(async () => {
    if (sessionId) await Promise.all([openSession(sessionId), refreshSessions()])
  }, [openSession, refreshSessions, sessionId])

  const applyEvent = useCallback((event: AgentEvent) => {
    if (event.id) setLastEventId((current) => Math.max(current, event.id))
    const data = event.data
    if (event.type === 'assistant.delta') {
      setMessages((items) => items.map((item) => item.runId === Number(data.runId) && item.role === 'assistant'
        ? { ...item, content: item.content + String(data.content || '') } : item))
    } else if (event.type === 'tool.started') {
      setOperations((items) => [...items, { callId: String(data.callId || `event-${event.id}`), runId: Number(data.runId), tool: String(data.tool), server: String(data.server || ''), status: 'running' }])
    } else if (event.type === 'tool.completed' || event.type === 'tool.failed') {
      setOperations((items) => items.map((item) => item.callId === String(data.callId) ? {
        ...item, status: event.type === 'tool.failed' ? 'failed' : String(data.status || 'completed'),
        durationMillis: Number(data.durationMillis || 0), summary: String((data.error as { message?: string } | undefined)?.message || '')
      } : item))
    } else if (event.type === 'approval.required') {
      setApprovals((items) => [...items, {
        runId: Number(data.runId), toolCode: String(data.toolCode || data.tool),
        toolName: String(data.toolName || data.tool), toolSource: String(data.server ? 'mcp' : 'local'),
        createdAt: new Date().toISOString(),
      }])
    }
    if (terminalEvents.has(event.type)) {
      setRunning(false)
      setMessages((items) => items.map((item) => item.runId === Number(data.runId) ? { ...item, streaming: false } : item))
      streamRef.current?.abort()
      void reloadActive()
    }
  }, [reloadActive])

  const connectStream = useCallback((id: number, cursor: number) => {
    streamRef.current?.abort()
    const controller = new AbortController()
    streamRef.current = controller
    void streamEvents(id, cursor, controller.signal, applyEvent).catch((cause: unknown) => {
      if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : '事件流已断开')
    })
  }, [applyEvent])

  const newConversation = async () => {
    try { const { session } = await createSession(); await refreshSessions(); await openSession(session.ID) }
    catch (cause) { setError(cause instanceof Error ? cause.message : '创建失败') }
  }

  const removeConversation = async (id: number) => {
    if (!window.confirm('删除这条对话及其审计轨迹？')) return
    try {
      await deleteSession(id)
      const items = await refreshSessions()
      if (id === sessionId) {
        if (items.length) await openSession(items[0].ID)
        else { setSessionId(0); setMessages([]); setOperations([]) }
      }
    } catch (cause) { setError(cause instanceof Error ? cause.message : '删除失败') }
  }

  const submit = async () => {
    const content = draft.trim()
    if (!content || running) return
    try {
      let targetSession = sessionId
      if (!targetSession) { const created = await createSession(); targetSession = created.session.ID; await refreshSessions(); setSessionId(targetSession) }
      setDraft(''); setSlashOpen(false); setError('')
      const response = await sendMessage(targetSession, content, directives)
      const nextRunId = response.run.id
      setRunId(nextRunId); setRunning(true)
      setMessages((items) => [...items,
        { localId: `user-${nextRunId}`, runId: nextRunId, role: 'user', content },
        { localId: `assistant-${nextRunId}`, runId: nextRunId, role: 'assistant', content: '', streaming: true }
      ])
      setDirectives([])
      connectStream(targetSession, lastEventId)
    } catch (cause) { setRunning(false); setError(cause instanceof Error ? cause.message : '发送失败') }
  }

  const selectCapability = (type: 'skill' | 'tool', name: string) => {
    setDirectives((items) => [...items.filter((item) => item.type !== type), { type, name }])
    setDraft((value) => value.replace(/(?:^|\s)\/[^\n/]*$/, '').trimStart())
    setSlashOpen(false)
  }
  const approve = async (approval: Approval) => {
    try { await approveRun(approval.runId); await reloadActive() } catch (cause) { setError(cause instanceof Error ? cause.message : '确认失败') }
  }

  const groupedOperations = useMemo(() => operations.filter((item) => !runId || item.runId === runId), [operations, runId])
  return (
    <main className={`agent-app ${historyCollapsed ? 'history-collapsed' : ''}`}>
      <aside className="session-sidebar">
        <div className="session-toolbar">
          <button className="new-chat" onClick={() => void newConversation()}><span>＋</span> 新对话</button>
          <button className="history-toggle" title="收起历史对话" aria-label="收起历史对话" onClick={() => setHistoryCollapsed(true)}><MenuFoldOutlined /></button>
        </div>
        <nav className="session-list" aria-label="对话列表">{sessions.map((session) =>
          <button key={session.ID} className={`session-row ${session.ID === sessionId ? 'active' : ''}`} onClick={() => void openSession(session.ID)}>
            <span><strong>{session.title || '新对话'}</strong><small>{formatTime(session.lastMessageAt || session.CreatedAt)}</small></span>
            <i role="button" aria-label="删除对话" tabIndex={0} onClick={(event) => { event.stopPropagation(); void removeConversation(session.ID) }}>×</i>
          </button>)}</nav>
      </aside>

      <section className="chat-stage">
        <header className="chat-header">
          <div className="chat-title">
            {historyCollapsed && <button className="history-toggle" title="展开历史对话" aria-label="展开历史对话" onClick={() => setHistoryCollapsed(false)}><MenuUnfoldOutlined /></button>}
            <div><strong>{activeSession?.title || '新对话'}</strong>{running && <small>正在执行</small>}</div>
          </div>
          <div className="header-actions">{running && <button className="stop-button" onClick={() => void cancelRun(sessionId, runId)}>停止</button>}<button className="trace-button" disabled={!sessionId} onClick={() => setTrajectoryOpen(true)}>轨迹 <span>⌘</span></button></div>
        </header>
        <div className="message-scroll" ref={viewportRef} aria-live="polite">
          {!messages.length && <section className="welcome"><h1>今天想完成什么？</h1>
            <div className="prompt-grid">{['帮我检索斗破苍穹相关资源，并展示完整检索过程', '列出当前可访问的来源，优先找名称包含“搜”的群组', '检查当前 MCP 能力，只说明不可用项'].map((text) =>
              <button key={text} onClick={() => setDraft(text)}>{text}<span>↗</span></button>)}</div></section>}
          <div className="message-column">
            {messages.map((message) => <article key={message.ID || message.localId} className={`message ${message.role}`}><div>{message.content || (message.streaming ? ' ' : '')}{message.streaming && <span className="typing-cursor" />}</div></article>)}
            {!!groupedOperations.length && <section className="operation-group">
              <button className="operation-summary" onClick={() => setOperationsOpen((value) => !value)}><span className="operation-icon">⌘</span><span>执行了 {groupedOperations.length} 个操作</span><small>{groupedOperations.some((item) => item.status === 'failed') ? '包含失败' : running ? '执行中' : '已完成'}</small><b>{operationsOpen ? '⌃' : '⌄'}</b></button>
              {operationsOpen && <div className="operation-list">{groupedOperations.map((operation) => <div className="operation-row" key={operation.callId}><span className={`op-state ${operation.status}`} /><div><strong>{operation.tool}</strong><small>{operation.server || 'MCP'} · {operation.callId}</small></div><em>{operation.durationMillis ? `${operation.durationMillis} ms` : operation.status}</em></div>)}</div>}
            </section>}
            {approvals.map((approval) => <section className="approval" key={`${approval.runId}-${approval.toolName}-${approval.createdAt}`}><div><strong>{approval.toolName} 需要确认</strong><p>该操作尚未执行，确认后会继续。</p></div><button onClick={() => void approve(approval)}>确认执行</button></section>)}
          </div>
        </div>

        <footer className="composer-wrap">
          {error && <div className="error-banner"><span>!</span>{error}<button onClick={() => setError('')}>×</button></div>}
          {!!directives.length && <div className="directive-list">{directives.map((directive) => <span key={directive.type} className={directive.type}><b>{directive.type}</b>{directive.name}<button onClick={() => setDirectives((items) => items.filter((item) => item.type !== directive.type))}>×</button></span>)}</div>}
          {slashOpen && <div className="slash-picker"><header><strong>选择 Skill 或 Tool</strong><small>输入 /skill 或 /tool 筛选</small></header>{filteredCapabilities.map(([type, name, description]) =>
            <button key={`${type}-${name}`} className={type} onMouseDown={(event) => { event.preventDefault(); selectCapability(type, name) }}><span>{type === 'skill' ? '◇' : '⌘'}</span><div><strong>{name}</strong><small>{description}</small></div><em>{type}</em></button>)}</div>}
          <div className="composer"><textarea value={draft} rows={2} disabled={running} placeholder="输入运营问题或任务，输入 / 选择 Skill 或 Tool" onChange={(event) => { setDraft(event.target.value); setSlashOpen(/(?:^|\s)\/[^\n/]*$/.test(event.target.value)) }} onKeyDown={(event) => {
            if (event.key === 'Escape') setSlashOpen(false)
            if (event.key === 'Enter' && !event.shiftKey && !slashOpen) { event.preventDefault(); void submit() }
          }} /><div className="composer-meta"><span className="model-chip"><i /> 默认模型</span><small>Enter 发送 · Shift + Enter 换行</small><button className="send-button" disabled={!draft.trim() || running} onClick={() => void submit()} aria-label="发送">↑</button></div></div>
          <p className="disclaimer">AI 生成内容可能有误，重要操作请核实。完整调用记录保存在审计轨迹中。</p>
        </footer>
      </section>
	  {trajectoryOpen && sessionId > 0 && <TrajectoryPanel sessionId={sessionId} onClose={() => setTrajectoryOpen(false)} />}
    </main>
  )
}

function formatTime(value?: string) {
  return value ? new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value)) : ''
}
