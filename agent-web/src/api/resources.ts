import type { PageResult } from "../models/http";
import type { AgentResourceTask, Resource, ResourceDetail } from "../models/resource";
import { queryString, request } from "./client";

export const resourcesApi = {
  list: (input: {
    page: number;
    pageSize: number;
    keyword?: string;
    mimeType?: string;
    state?: string;
  }) => request<PageResult<Resource>>(`/xy/resources${queryString(input)}`),
  get: (ID: number) => request<{ resource: ResourceDetail }>(`/xy/resources/${ID}`),
  preview: (ID: number) =>
    request<{ previewUrl: string; expiresAt: string }>(`/xy/resources/${ID}/preview`),
  download: (ID: number) =>
    request<{ downloadUrl: string; expiresAt: string }>(`/xy/resources/${ID}/download`),
  tasks: (input: { page: number; pageSize: number; state?: string }) =>
    request<PageResult<AgentResourceTask>>(`/xy/resource-tasks${queryString(input)}`),
  cancelTask: (ID: number) =>
    request<{ task: AgentResourceTask }>(`/xy/resource-tasks/${ID}/cancel`, { method: "POST" }),
};
