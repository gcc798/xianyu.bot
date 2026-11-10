import type {
  TelegramAPIConfigInput,
  TelegramAPIConfigPage,
} from "../models/telegram-api";
import type { EmptyResponse } from "../models/http";
import { queryString, request } from "./client";

export const telegramAPIConfigApi = {
  list: (input: { page: number; pageSize: number; q?: string }) =>
    request<TelegramAPIConfigPage>(`/xy/telegram-api-configs${queryString(input)}`),
  create: (input: TelegramAPIConfigInput) =>
    request<EmptyResponse>("/xy/telegram-api-configs", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  update: (id: number, input: TelegramAPIConfigInput) =>
    request<EmptyResponse>(`/xy/telegram-api-configs/${id}`, {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  delete: (id: number) =>
    request<EmptyResponse>(`/xy/telegram-api-configs/${id}`, { method: "DELETE" }),
};
