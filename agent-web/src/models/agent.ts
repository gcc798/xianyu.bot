export type Directive = { type: "skill" | "tool"; name: string };

export type AgentToolCatalogEntry = {
  code: string;
  name: string;
  description: string;
  source: string;
  providerCode?: string;
  type: "tool";
};

export type Session = {
  ID: number;
  title: string;
  CreatedAt?: string;
  lastMessageAt?: string;
};

export type ChatMessage = {
  ID?: number;
  localId?: string;
  runId?: number;
  role: "user" | "assistant";
  content: string;
  CreatedAt?: string;
  streaming?: boolean;
};

export type ToolTrace = {
  id: number;
  runId: number;
  callId?: string;
  toolCode: string;
  toolName: string;
  toolSource: string;
  arguments?: string;
  result?: string;
  status: string;
  durationMillis: number;
  createdAt: string;
};

export type Approval = {
  runId: number;
  toolCode: string;
  toolName: string;
  toolSource: string;
  createdAt: string;
};

export type AgentTask = {
  id: number;
  runId: number;
  state: string;
  stage: string;
  progressPercent: number;
  progressDone: number;
  progressTotal: number;
  progressUnit: string;
};

export type SessionDetail = Session & {
  messages: ChatMessage[];
  toolTraces: ToolTrace[];
  pendingApprovals: Approval[];
  tasks: AgentTask[];
};

export type AgentEvent = {
  id: number;
  type: string;
  data: Record<string, unknown>;
  createdAt?: string;
  callId?: string;
  runId?: number;
  stepNo?: number;
};

export type Operation = {
  callId: string;
  runId: number;
  tool: string;
  server?: string;
  status: string;
  durationMillis?: number;
  summary?: string;
};

export type ToolCallDetail = {
  id: number;
  runId: number;
  callId: string;
  mcpServerCode: string;
  toolCode: string;
  toolName: string;
  toolSource: string;
  arguments: unknown;
  result: unknown;
  status: string;
  durationMillis: number;
  errorCode?: string;
  errorMessage?: string;
  startedAt: string;
  completedAt?: string;
};
