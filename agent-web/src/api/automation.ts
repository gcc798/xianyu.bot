import type {
  BusinessEvent,
  DeliveryInventory,
  DeliveryRule,
  DeliveryRuleInput,
  DeliveryTask,
  ReplyRule,
  ReplyRuleInput,
} from "../models/automation";
import type { EmptyResponse, PageResult } from "../models/http";
import { queryString, request } from "./client";

export const automationApi = {
  replyRules: (input: { page: number; pageSize: number; keyword?: string; productId?: number; type?: string }) =>
    request<PageResult<ReplyRule>>(`/xy/reply-rule/list${queryString(input)}`),
  saveReplyRule: (input: ReplyRuleInput) =>
    request<EmptyResponse>("/xy/reply-rule/save", { method: "POST", body: JSON.stringify(input) }),
  deleteReplyRule: (ID: number) =>
    request<EmptyResponse>(`/xy/reply-rule/delete${queryString({ ID })}`, { method: "DELETE" }),
  deliveryRules: (input: { page: number; pageSize: number; keyword?: string; productId?: number; mode?: string }) =>
    request<PageResult<DeliveryRule>>(`/xy/delivery-rule/list${queryString(input)}`),
  saveDeliveryRule: (input: DeliveryRuleInput) =>
    request<EmptyResponse>("/xy/delivery-rule/save", { method: "POST", body: JSON.stringify(input) }),
  deleteDeliveryRule: (ID: number) =>
    request<EmptyResponse>(`/xy/delivery-rule/delete${queryString({ ID })}`, { method: "DELETE" }),
  importResources: (ruleId: number, values: string[]) =>
    request<{ count: number }>("/xy/delivery-rule/resource-import", {
      method: "POST",
      body: JSON.stringify({ ruleId, values }),
    }),
  inventory: (ID: number) =>
    request<DeliveryInventory>(`/xy/delivery-rule/inventory${queryString({ ID })}`),
  deliveryTasks: (input: { page: number; pageSize: number; keyword?: string; accountId?: number; status?: string }) =>
    request<PageResult<DeliveryTask>>(`/xy/delivery-rule/task-list${queryString(input)}`),
  retryDeliveryTask: (ID: number) =>
    request<EmptyResponse>("/xy/delivery-rule/task-retry", { method: "POST", body: JSON.stringify({ ID }) }),
  businessEvents: (input: {
    page: number;
    pageSize: number;
    keyword?: string;
    accountId?: number;
    productId?: number;
    orderId?: number;
    type?: string;
    status?: string;
  }) => request<PageResult<BusinessEvent>>(`/xy/event/list${queryString(input)}`),
};
