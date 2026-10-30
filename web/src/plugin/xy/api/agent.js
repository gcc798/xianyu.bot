import service from '@/utils/request'
import { useUserStore } from '@/pinia/modules/user'

export const getAgentSessions = (params = { page: 1, pageSize: 50 }) => service({
  url: '/xy/ai/agent/sessions', method: 'get', params, donNotShowLoading: true
})

export const createAgentSession = (data = {}) => service({
  url: '/xy/ai/agent/sessions', method: 'post', data, donNotShowLoading: true
})

export const getAgentSession = (id) => service({
  url: `/xy/ai/agent/sessions/${id}`, method: 'get', donNotShowLoading: true
})

export const deleteAgentSession = (id) => service({
  url: `/xy/ai/agent/sessions/${id}`, method: 'delete', donNotShowLoading: true
})

export const sendAgentMessage = (sessionId, data) => service({
  url: `/xy/ai/agent/sessions/${sessionId}/messages`, method: 'post', data,
  headers: { 'Idempotency-Key': data.idempotencyKey }, donNotShowLoading: true
})

export const cancelAgentRun = (sessionId, runId) => service({
  url: `/xy/ai/agent/sessions/${sessionId}/runs/${runId}`, method: 'delete', donNotShowLoading: true
})

export const approveAgentRun = (runId) => service({
  url: `/xy/ai/agent/runs/${runId}/approve`, method: 'post', donNotShowLoading: true
})

export const cancelResourceTask = (taskId) => service({
  url: `/xy/resource-tasks/${taskId}/cancel`, method: 'post', donNotShowLoading: true
})

export const streamAgentEvents = async (sessionId, { afterId = 0, signal, onEvent }) => {
  const userStore = useUserStore()
  const response = await fetch(`${import.meta.env.VITE_BASE_API || ''}/xy/ai/agent/sessions/${sessionId}/events`, {
    headers: {
      Accept: 'text/event-stream',
      'Last-Event-ID': String(afterId || ''),
      'x-token': userStore.token || '',
      'x-user-id': userStore.userInfo.ID || userStore.userInfo.id || ''
    },
    signal
  })
  if (!response.ok || !response.body) throw new Error('无法连接 Agent 事件流')

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  const emitBlock = (block) => {
    if (!block || block.startsWith(':')) return
    let id = 0
    let type = 'message'
    const data = []
    for (const line of block.split('\n')) {
      if (line.startsWith('id:')) id = Number(line.slice(3).trim()) || 0
      else if (line.startsWith('event:')) type = line.slice(6).trim()
      else if (line.startsWith('data:')) data.push(line.slice(5).trimStart())
    }
    if (!data.length) return
    let payload = data.join('\n')
    try { payload = JSON.parse(payload) } catch { /* keep text payload */ }
    onEvent?.({ id, type, data: payload })
  }

  while (true) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true }).replaceAll('\r\n', '\n')
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      emitBlock(buffer.slice(0, boundary))
      buffer = buffer.slice(boundary + 2)
      boundary = buffer.indexOf('\n\n')
    }
  }
}
