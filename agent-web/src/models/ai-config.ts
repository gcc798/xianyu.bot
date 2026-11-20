export type AIModel = {
  ID: number;
  name: string;
  provider: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  enabled: boolean;
  priority: number;
  supportsTools: boolean;
  supportsStreaming: boolean;
  contextWindow: number;
  maxOutputTokens: number;
  requestTimeoutSeconds: number;
  extraConfig: string;
  hasApiKey: boolean;
};

export type AIModelInput = Omit<AIModel, "ID" | "hasApiKey">;

export type MCPServer = {
  ID: number;
  name: string;
  code: string;
  description: string;
  tags: string[];
  baseUrl: string;
  authHeader: string;
  enabled: boolean;
  enabledTools: string[];
  disabledTools: string[];
  requestTimeoutSeconds: number;
  taskTimeoutSeconds: number;
  catalogTTLSeconds: number;
  connectionTTLSeconds: number;
  connectionScope: "shared" | "owner";
  maxConcurrentCalls: number;
  catalogStatus: string;
  catalogToolCount: number;
  catalogRefreshedAt: string | null;
  catalogError: string;
  lastStatus: string;
  lastCheckedAt: string | null;
  lastCheckDurationMillis: number;
  lastError: string;
  hasAuthToken: boolean;
};

export type MCPServerInput = Omit<
  MCPServer,
  | "ID"
  | "catalogStatus"
  | "catalogToolCount"
  | "catalogRefreshedAt"
  | "catalogError"
  | "lastStatus"
  | "lastCheckedAt"
  | "lastCheckDurationMillis"
  | "lastError"
  | "hasAuthToken"
> & { authToken: string };

export type MCPTool = {
  name: string;
  description: string;
  definitionHash: string;
  discoveredAt: string;
};

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
