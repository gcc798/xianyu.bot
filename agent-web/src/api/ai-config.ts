import type {
  AgentConfig,
  AIModel,
  AIModelInput,
  MCPServer,
  MCPServerInput,
} from "../models/ai-config";
import type { EmptyResponse } from "../models/http";
import { request } from "./client";

export const aiConfigApi = {
  models: () => request<{ list: AIModel[] }>("/xy/ai/models"),
  createModel: (input: AIModelInput) =>
    request<{ model: AIModel }>("/xy/ai/models", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateModel: (ID: number, input: AIModelInput) =>
    request<EmptyResponse>(`/xy/ai/models/${ID}`, {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  deleteModel: (ID: number) => request<EmptyResponse>(`/xy/ai/models/${ID}`, { method: "DELETE" }),
  setDefaultModel: (ID: number) =>
    request<EmptyResponse>(`/xy/ai/models/${ID}/default`, { method: "PUT" }),
  mcpServers: () => request<{ list: MCPServer[] }>("/xy/ai/mcp-servers"),
  createMCPServer: (input: MCPServerInput) =>
    request<{ server: MCPServer }>("/xy/ai/mcp-servers", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateMCPServer: (ID: number, input: MCPServerInput) =>
    request<EmptyResponse>(`/xy/ai/mcp-servers/${ID}`, {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  deleteMCPServer: (ID: number) =>
    request<EmptyResponse>(`/xy/ai/mcp-servers/${ID}`, { method: "DELETE" }),
  agentConfig: () => request<{ config: AgentConfig }>("/xy/ai/agent-config"),
  updateAgentConfig: (input: AgentConfig) =>
    request<EmptyResponse>("/xy/ai/agent-config", {
      method: "PUT",
      body: JSON.stringify(input),
    }),
};
