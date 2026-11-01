import type { AgentEvent, Directive, Session, SessionDetail, ToolCallDetail } from './types'

export const baseURL = (import.meta.env.VITE_BASE_API || '/api').replace(/\/$/, '')
export type Envelope<T> = { code: number; data: T; msg: string }

export function token() {
  const cookie = document.cookie.split('; ').find((item) => item.startsWith('x-token='))?.split('=')[1]
  return localStorage.getItem('token') || (cookie ? decodeURIComponent(cookie) : '')
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const user = JSON.parse(localStorage.getItem('xy-user') || '{}') as { ID?: number }
	const response = await fetch(`${baseURL}${path}`, { ...init, headers: { 'Content-Type': 'application/json', 'x-token': token(), 'x-user-id': String(user.ID || ''), ...init.headers } })
	const nextToken = response.headers.get('new-token')
	if (nextToken) localStorage.setItem('token', nextToken)
	if (response.status === 401) { localStorage.removeItem('token'); window.location.assign('/agent/'); throw new Error('登录已失效') }
  const payload = await response.text()
  let body: Envelope<T>
  try { body = JSON.parse(payload) as Envelope<T> } catch { throw new Error('Agent 服务暂不可用') }
  if (!response.ok || body.code !== 0) throw new Error(body.msg || '请求失败')
  return body.data
}

export const listSessions = () => request<{ list: Session[] }>('/xy/ai/agent/sessions?page=1&pageSize=100')
export const createSession = () => request<{ session: Session }>('/xy/ai/agent/sessions', { method: 'POST', body: JSON.stringify({ title: '新对话' }) })
export const deleteSession = (id: number) => request<Record<string, never>>(`/xy/ai/agent/sessions/${id}`, { method: 'DELETE' })
export const getSession = (id: number) => request<{ session: SessionDetail }>(`/xy/ai/agent/sessions/${id}`)
export const getTrajectory = (id: number, afterId = 0) => request<{ events: AgentEvent[]; nextAfterId: number; hasMore: boolean }>(`/xy/ai/agent/sessions/${id}/trajectory?afterId=${afterId}&limit=500`)
export const getToolCall = (id: number, callId: string) => request<{ toolCall: ToolCallDetail }>(`/xy/ai/agent/sessions/${id}/tool-calls/${encodeURIComponent(callId)}`)

export function sendMessage(sessionId: number, message: string, directives: Directive[]) {
  const idempotencyKey = crypto.randomUUID()
  return request<{ run: { id: number; sessionId: number; status: string } }>(`/xy/ai/agent/sessions/${sessionId}/messages`, {
    method: 'POST', headers: { 'Idempotency-Key': idempotencyKey }, body: JSON.stringify({ message, directives, idempotencyKey })
  })
}

export const cancelRun = (sessionId: number, runId: number) => request<Record<string, never>>(`/xy/ai/agent/sessions/${sessionId}/runs/${runId}`, { method: 'DELETE' })
export const approveRun = (runId: number) => request<{ tool: string; status: string }>(`/xy/ai/agent/runs/${runId}/approve`, { method: 'POST' })

export async function streamEvents(sessionId: number, afterId: number, signal: AbortSignal, onEvent: (event: AgentEvent) => void) {
  const response = await fetch(`${baseURL}/xy/ai/agent/sessions/${sessionId}/events`, {
    headers: { Accept: 'text/event-stream', 'Last-Event-ID': String(afterId || ''), 'x-token': token() }, signal
  })
  if (!response.ok || !response.body) throw new Error('无法连接 Agent 事件流')
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { value, done } = await reader.read()
    if (done) return
    buffer += decoder.decode(value, { stream: true }).replaceAll('\r\n', '\n')
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const block = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      boundary = buffer.indexOf('\n\n')
      if (!block || block.startsWith(':')) continue
      let id = 0
      let type = 'message'
      const data: string[] = []
      for (const line of block.split('\n')) {
        if (line.startsWith('id:')) id = Number(line.slice(3).trim()) || 0
        else if (line.startsWith('event:')) type = line.slice(6).trim()
        else if (line.startsWith('data:')) data.push(line.slice(5).trimStart())
      }
      if (data.length) onEvent({ id, type, data: JSON.parse(data.join('\n')) as Record<string, unknown> })
    }
  }
}
