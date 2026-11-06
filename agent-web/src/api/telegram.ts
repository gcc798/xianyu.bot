import type {
  TelegramAISettingsInput,
  TelegramDialogPage,
  TelegramQR,
  TelegramResourceFilePage,
  TelegramResourceOptions,
  TelegramResourceTask,
  TelegramResourceTaskInput,
  TelegramResourceTaskPage,
  TelegramScan,
  TelegramSettings,
  TelegramStatus,
} from "../models/telegram";
import { acceptRenewedToken, accessToken, apiBaseURL, queryString } from "./client";

const basePath = "/xy/proxy/telegram";

async function telegramRequest<T>(path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers);
  const token = accessToken();
  if (token) headers.set("x-token", token);
  if (init.body) headers.set("Content-Type", "application/json");
  const response = await fetch(`${apiBaseURL}${basePath}${path}`, { ...init, headers });
  acceptRenewedToken(response);
  if (response.status === 401) throw new Error("Telegram 未登录");
  const data = (await response.json()) as T & { detail?: string };
  if (!response.ok) throw new Error(data.detail || "Telegram 服务请求失败");
  return data;
}

async function telegramFile(path: string) {
  const headers = new Headers();
  const token = accessToken();
  if (token) headers.set("x-token", token);
  const response = await fetch(`${apiBaseURL}${basePath}${path}`, { headers });
  acceptRenewedToken(response);
  if (!response.ok) {
    const data = (await response.json().catch(() => ({ detail: "文件读取失败" }))) as {
      detail?: string;
    };
    throw new Error(data.detail || "文件读取失败");
  }
  return response.blob();
}

export const telegramApi = {
  status: () => telegramRequest<TelegramStatus>("/api/admin/status"),
  startQR: () => telegramRequest<TelegramQR>("/api/admin/login/qr", { method: "POST" }),
  loginStatus: () => telegramRequest<TelegramQR>("/api/admin/login/status"),
  submitPassword: (password: string) =>
    telegramRequest<TelegramQR>("/api/admin/login/2fa", {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
  startScan: () =>
    telegramRequest<{ started: boolean; task_id: number | null; progress: TelegramScan }>(
      "/api/admin/scan",
      { method: "POST" },
    ),
  scanStatus: () => telegramRequest<TelegramScan>("/api/admin/scan/status"),
  dialogs: (input: {
    page: number;
    page_size: number;
    q?: string;
    filter?: string;
    members_min?: number;
    members_max?: number;
  }) => telegramRequest<TelegramDialogPage>(`/api/admin/dialogs${queryString(input)}`),
  exportDialogs: (input: {
    q?: string;
    filter?: string;
    members_min?: number;
    members_max?: number;
  }) => telegramFile(`/api/admin/dialogs/export${queryString(input)}`),
  leaveDialogs: (dialog_ids: number[]) =>
    telegramRequest<{ success: number; failed: number }>("/api/admin/dialogs/leave", {
      method: "POST",
      body: JSON.stringify({ dialog_ids, confirm: "退出" }),
    }),
  settings: () => telegramRequest<TelegramSettings>("/api/admin/settings"),
  updateAdultMode: (adult_mode: boolean) =>
    telegramRequest<{ adult_mode: boolean; adult_dialog_count: number }>(
      "/api/admin/settings/adult-mode",
      {
        method: "POST",
        body: JSON.stringify({ adult_mode }),
      },
    ),
  updateAISettings: (input: TelegramAISettingsInput) =>
    telegramRequest<
      Pick<
        TelegramSettings,
        "ai_base_url" | "ai_model" | "ai_api_key_configured" | "ai_api_key_masked"
      >
    >("/api/admin/settings/ai", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  probeAISettings: (input: TelegramAISettingsInput) =>
    telegramRequest<{ available: boolean }>("/api/admin/settings/ai/probe", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  resourceOptions: () => telegramRequest<TelegramResourceOptions>("/api/resources/options"),
  resourceTasks: (input: { page: number; page_size: number; q?: string; status?: string }) =>
    telegramRequest<TelegramResourceTaskPage>(`/api/resources/tasks${queryString(input)}`),
  createResourceTask: (input: TelegramResourceTaskInput) =>
    telegramRequest<TelegramResourceTask>("/api/resources/tasks", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  cancelResourceTask: (id: number) =>
    telegramRequest<{ ok: boolean }>(`/api/resources/tasks/${id}/cancel`, { method: "POST" }),
  retryResourceTask: (id: number) =>
    telegramRequest<{ ok: boolean }>(`/api/resources/tasks/${id}/retry`, { method: "POST" }),
  deleteResourceTask: (id: number, deleteFiles: boolean) =>
    telegramRequest<{ ok: boolean }>(
      `/api/resources/tasks/${id}${queryString({ delete_files: deleteFiles })}`,
      { method: "DELETE" },
    ),
  resourceFiles: (input: {
    page: number;
    page_size: number;
    task_id?: number;
    filename?: string;
    file_type?: string;
    status?: string;
    source_id?: number;
  }) => telegramRequest<TelegramResourceFilePage>(`/api/resources/files${queryString(input)}`),
  resourceFile: (id: number, action: "preview" | "download") =>
    telegramFile(`/api/resources/files/${id}/${action}`),
  deleteResourceFile: (id: number) =>
    telegramRequest<{ ok: boolean }>(`/api/resources/files/${id}`, { method: "DELETE" }),
};
