export type AIModel = {
  ID: number;
  name: string;
  provider: string;
  baseUrl: string;
  model: string;
  enabled: boolean;
  isDefault: boolean;
  supportsTools: boolean;
  supportsStreaming: boolean;
  contextWindow: number;
  maxOutputTokens: number;
  requestTimeoutSeconds: number;
  extraConfig: string;
  hasApiKey: boolean;
};

export type AIModelInput = Omit<AIModel, "ID" | "hasApiKey"> & { apiKey: string };

export type MCPServer = {
  ID: number;
  name: string;
  code: string;
  baseUrl: string;
  authHeader: string;
  enabled: boolean;
  requestTimeoutSeconds: number;
  taskTimeoutSeconds: number;
  lastStatus: string;
  lastCheckedAt: string | null;
  lastError: string;
  hasAuthToken: boolean;
};

export type MCPServerInput = Omit<
  MCPServer,
  "ID" | "lastStatus" | "lastCheckedAt" | "lastError" | "hasAuthToken"
> & { authToken: string };

export type AgentConfig = {
  ID: number;
  maxWallTimeSeconds: number;
  maxIterations: number;
  maxModelCalls: number;
  maxToolCalls: number;
  maxConcurrentToolCalls: number;
  maxSearchCalls: number;
  maxWriteToolCalls: number;
  maxOutputTokens: number;
  conversationHistoryLimit: number;
  summaryTriggerTokens: number;
};
