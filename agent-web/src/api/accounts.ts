import type { EmptyResponse, PageResult } from "../models/http";
import type { AccountCredentialInput, AccountInput, AccountSearch, LoginStatus, QRLogin, WebSocketStatus, XianyuAccount } from "../models/account";
import { queryString, request } from "./client";

export const accountsApi = {
  list: (input: AccountSearch) => request<PageResult<XianyuAccount>>(`/xy/account/list${queryString(input)}`),
  create: (input: Omit<AccountInput, "ID" | "enabled">) => request<{ account: XianyuAccount }>("/xy/account/create", { method: "POST", body: JSON.stringify(input) }),
  update: (input: Required<AccountInput>) => request<EmptyResponse>("/xy/account/update", { method: "PUT", body: JSON.stringify(input) }),
  delete: (ID: number) => request<EmptyResponse>(`/xy/account/delete${queryString({ ID })}`, { method: "DELETE" }),
  saveCredential: (input: AccountCredentialInput) => request<EmptyResponse>("/xy/account/credential", { method: "PUT", body: JSON.stringify(input) }),
  checkLogin: (ID: number) => request<{ login: LoginStatus }>("/xy/account/check-login", { method: "POST", body: JSON.stringify({ ID }) }),
  startQRLogin: (ID: number) => request<{ login: QRLogin }>("/xy/qr-login/start", { method: "POST", body: JSON.stringify({ ID }) }),
  qrLoginStatus: (sessionId: string) => request<{ login: QRLogin }>(`/xy/qr-login/status${queryString({ sessionId })}`),
  cancelQRLogin: (sessionId: string) => request<EmptyResponse>(`/xy/qr-login/cancel${queryString({ sessionId })}`, { method: "DELETE" }),
  connectionStatus: (ID: number) => request<{ connection: WebSocketStatus }>(`/xy/websocket/status${queryString({ ID })}`),
  startConnection: (ID: number) => request<EmptyResponse>("/xy/websocket/start", { method: "POST", body: JSON.stringify({ ID }) }),
  stopConnection: (ID: number) => request<EmptyResponse>("/xy/websocket/stop", { method: "POST", body: JSON.stringify({ ID }) }),
};
