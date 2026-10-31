export type Directive = { type: 'skill' | 'tool'; name: string }

export type Session = { ID: number; title: string; CreatedAt?: string; lastMessageAt?: string }
export type ChatMessage = { ID?: number; localId?: string; runId?: number; role: 'user' | 'assistant'; content: string; CreatedAt?: string; streaming?: boolean }
export type ToolTrace = { id: number; runId: number; callId?: string; toolName: string; arguments?: string; result?: string; status: string; durationMillis: number; createdAt: string }
export type Approval = { runId: number; toolName: string; createdAt: string }
export type SessionDetail = Session & { messages: ChatMessage[]; toolTraces: ToolTrace[]; pendingApprovals: Approval[]; tasks: unknown[] }
export type AgentEvent = { id: number; type: string; data: Record<string, unknown>; createdAt?: string; callId?: string; runId?: number; stepNo?: number }
export type Operation = { callId: string; runId: number; tool: string; server?: string; status: string; durationMillis?: number; summary?: string }
