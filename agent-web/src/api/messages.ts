import type { EmptyResponse, PageResult } from "../models/http";
import type { MessageSearch, XianyuMessage } from "../models/message";
import { queryString, request } from "./client";

export const messagesApi = {
  list: (input: MessageSearch) => request<PageResult<XianyuMessage>>(`/xy/message/list${queryString(input)}`),
  reply: (messageId: number, text: string) => request<EmptyResponse>("/xy/message/reply", { method: "POST", body: JSON.stringify({ messageId, text }) }),
  updateTakeover: (messageId: number, active: boolean) => request<EmptyResponse>("/xy/message/takeover", { method: "PUT", body: JSON.stringify({ messageId, active }) }),
};
