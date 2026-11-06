import type { TelegramAISettingsInput, TelegramDialogPage, TelegramQR, TelegramScan, TelegramSettings, TelegramStatus } from "../models/telegram";
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
  const data = await response.json() as T & { detail?: string };
  if (!response.ok) throw new Error(data.detail || "Telegram 服务请求失败");
  return data;
}

export const telegramApi = {
  status: () => telegramRequest<TelegramStatus>("/api/admin/status"),
  startQR: () => telegramRequest<TelegramQR>("/api/admin/login/qr", { method: "POST" }),
  loginStatus: () => telegramRequest<TelegramQR>("/api/admin/login/status"),
  submitPassword: (password: string) => telegramRequest<TelegramQR>("/api/admin/login/2fa", {
    method: "POST", body: JSON.stringify({ password }),
  }),
  startScan: () => telegramRequest<{ started: boolean; task_id: number | null; progress: TelegramScan }>("/api/admin/scan", { method: "POST" }),
  scanStatus: () => telegramRequest<TelegramScan>("/api/admin/scan/status"),
  dialogs: (input: { page: number; page_size: number; q?: string; filter?: string; members_min?: number; members_max?: number }) =>
    telegramRequest<TelegramDialogPage>(`/api/admin/dialogs${queryString(input)}`),
  leaveDialogs: (dialog_ids: number[]) => telegramRequest<{ success: number; failed: number }>("/api/admin/dialogs/leave", {
    method: "POST", body: JSON.stringify({ dialog_ids, confirm: "退出" }),
  }),
  settings: () => telegramRequest<TelegramSettings>("/api/admin/settings"),
  updateAdultMode: (adult_mode: boolean) => telegramRequest<{ adult_mode: boolean; adult_dialog_count: number }>("/api/admin/settings/adult-mode", {
    method: "POST", body: JSON.stringify({ adult_mode }),
  }),
  updateAISettings: (input: TelegramAISettingsInput) => telegramRequest<Pick<TelegramSettings, "ai_base_url" | "ai_model" | "ai_api_key_configured" | "ai_api_key_masked">>("/api/admin/settings/ai", {
    method: "POST", body: JSON.stringify(input),
  }),
  probeAISettings: (input: TelegramAISettingsInput) => telegramRequest<{ available: boolean }>("/api/admin/settings/ai/probe", {
    method: "POST", body: JSON.stringify(input),
  }),
};
