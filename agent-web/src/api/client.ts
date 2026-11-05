import type { ApiEnvelope } from "../models/http";
import type { AuthSession } from "../models/auth";

const TOKEN_KEY = "token";
const SESSION_KEY = "xy-auth-session";

export const apiBaseURL = (import.meta.env.VITE_BASE_API || "/api").replace(/\/$/, "");

export function accessToken() {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function storeAccessToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(SESSION_KEY);
}

export function acceptRenewedToken(response: Response) {
  const token = response.headers.get("new-token");
  if (!token) return;
  storeAccessToken(token);
  const storedSession = localStorage.getItem(SESSION_KEY);
  if (!storedSession) return;
  try {
    const session = JSON.parse(storedSession) as AuthSession;
    const expiresAt = Number(response.headers.get("new-expires-at"));
    localStorage.setItem(
      SESSION_KEY,
      JSON.stringify({ ...session, token, expiresAt: expiresAt > 0 ? expiresAt * 1000 : session.expiresAt }),
    );
  } catch {
    clearStoredAuth();
  }
}

export function queryString(params: Record<string, string | number | boolean | undefined | null>) {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") query.set(key, String(value));
  });
  const value = query.toString();
  return value ? `?${value}` : "";
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = accessToken();
  if (token) headers.set("x-token", token);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");

  const response = await fetch(`${apiBaseURL}${path}`, { ...init, headers });
  acceptRenewedToken(response);

  if (response.status === 401) {
    clearStoredAuth();
    if (!window.location.pathname.endsWith("/login")) window.location.assign("/agent/login");
    throw new Error("登录已失效，请重新登录");
  }

  const text = await response.text();
  let envelope: ApiEnvelope<T>;
  try {
    envelope = JSON.parse(text) as ApiEnvelope<T>;
  } catch {
    throw new Error("服务返回了无法识别的数据");
  }
  if (!response.ok || envelope.code !== 0) throw new Error(envelope.msg || "请求失败");
  return envelope.data;
}
