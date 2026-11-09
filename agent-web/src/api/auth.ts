import type { AuthSession, CurrentUser, LoginRequest, LoginResponse } from "../models/auth";
import type { EmptyResponse } from "../models/http";
import { clearStoredAuth, request, storeAccessToken } from "./client";

const SESSION_KEY = "xy-auth-session";

export function readAuthSession(): AuthSession | null {
  const value = localStorage.getItem(SESSION_KEY);
  if (!value) return null;
  try {
    return JSON.parse(value) as AuthSession;
  } catch {
    clearStoredAuth();
    return null;
  }
}

export function storeAuthSession(session: AuthSession) {
  storeAccessToken(session.token);
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

export async function login(input: LoginRequest) {
  const session = await request<LoginResponse>("/base/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
  storeAuthSession(session);
  return session;
}

export async function logout() {
  try {
    await request<EmptyResponse>("/jwt/jsonInBlacklist", { method: "POST" });
  } finally {
    clearStoredAuth();
  }
}

export function getCurrentUser() {
  return request<{ userInfo: CurrentUser }>("/user/getUserInfo");
}

export function updateCurrentUser(
  input: Pick<CurrentUser, "nickName" | "phone" | "email" | "headerImg">,
) {
  return request<EmptyResponse>("/user/setSelfInfo", {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function changePassword(input: { password: string; newPassword: string }) {
  return request<EmptyResponse>("/user/changePassword", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
