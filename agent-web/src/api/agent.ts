import type {
  AgentToolCatalogEntry,
  AgentEvent,
  Directive,
  Session,
  SessionDetail,
  ToolCallDetail,
} from "../models/agent";
import type { EmptyResponse, PageResult } from "../models/http";
import { acceptRenewedToken, accessToken, apiBaseURL, request } from "./client";

export const listSessions = () =>
  request<PageResult<Session>>("/xy/ai/agent/sessions?page=1&pageSize=100");

export const listAgentTools = () =>
  request<{ list: AgentToolCatalogEntry[] }>("/xy/ai/agent/tools");

export const createSession = () =>
  request<{ session: Session }>("/xy/ai/agent/sessions", {
    method: "POST",
    body: JSON.stringify({ title: "新对话" }),
  });

export const deleteSession = (id: number) =>
  request<EmptyResponse>(`/xy/ai/agent/sessions/${id}`, { method: "DELETE" });

export const updateSession = (id: number, title: string) =>
  request<EmptyResponse>(`/xy/ai/agent/sessions/${id}`, {
    method: "PUT",
    body: JSON.stringify({ title }),
  });

export const getSession = (id: number) =>
  request<{ session: SessionDetail }>(`/xy/ai/agent/sessions/${id}`);

export const getTrajectory = (id: number, afterId = 0) =>
  request<{ events: AgentEvent[]; nextAfterId: number; hasMore: boolean }>(
    `/xy/ai/agent/sessions/${id}/trajectory?afterId=${afterId}&limit=500`,
  );

export const getToolCall = (id: number, callId: string) =>
  request<{ toolCall: ToolCallDetail }>(
    `/xy/ai/agent/sessions/${id}/tool-calls/${encodeURIComponent(callId)}`,
  );

export function sendMessage(sessionId: number, message: string, directives: Directive[]) {
  // 同一个键同时放入请求头和请求体，供网关与业务层共同识别重复提交。
  const idempotencyKey = crypto.randomUUID();
  return request<{ run: { id: number; sessionId: number; status: string } }>(
    `/xy/ai/agent/sessions/${sessionId}/messages`,
    {
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: JSON.stringify({ message, directives, idempotencyKey }),
    },
  );
}

export const cancelRun = (sessionId: number, runId: number) =>
  request<EmptyResponse>(`/xy/ai/agent/sessions/${sessionId}/runs/${runId}`, {
    method: "DELETE",
  });

export const approveRun = (runId: number) =>
  request<{ tool: string; status: string }>(`/xy/ai/agent/runs/${runId}/approve`, {
    method: "POST",
  });

export async function streamEvents(
  sessionId: number,
  afterId: number,
  signal: AbortSignal,
  onEvent: (event: AgentEvent) => void,
) {
  const response = await fetch(`${apiBaseURL}/xy/ai/agent/sessions/${sessionId}/events`, {
    headers: {
      Accept: "text/event-stream",
      "Last-Event-ID": String(afterId || ""),
      "x-token": accessToken(),
    },
    signal,
  });
  acceptRenewedToken(response);
  if (!response.ok || !response.body) throw new Error("无法连接 Bot 事件流");

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  while (true) {
    const { value, done } = await reader.read();
    if (done) return;
    // 网络分片不保证落在 SSE 事件边界，必须保留未完成片段等待下一次读取。
    buffer += decoder.decode(value, { stream: true }).replaceAll("\r\n", "\n");
    let boundary = buffer.indexOf("\n\n");
    while (boundary >= 0) {
      const block = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);
      boundary = buffer.indexOf("\n\n");
      if (!block || block.startsWith(":")) continue;
      let id = 0;
      let type = "message";
      const lines: string[] = [];
      for (const line of block.split("\n")) {
        if (line.startsWith("id:")) id = Number(line.slice(3).trim()) || 0;
        else if (line.startsWith("event:")) type = line.slice(6).trim();
        else if (line.startsWith("data:")) lines.push(line.slice(5).trimStart());
      }
      if (lines.length) {
        onEvent({
          id,
          type,
          data: JSON.parse(lines.join("\n")) as Record<string, unknown>,
        });
      }
    }
  }
}
