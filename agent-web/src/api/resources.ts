import type { PageResult } from "../models/http";
import type { Resource, ResourceDetail } from "../models/resource";
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
  download: (ID: number) =>
    request<{ downloadUrl: string; expiresAt: string }>(`/xy/resources/${ID}/download`),
};
