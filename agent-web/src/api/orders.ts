import type { PageResult } from "../models/http";
import type { Order, OrderSearch } from "../models/order";
import { queryString, request } from "./client";

export const ordersApi = {
  list: (input: OrderSearch) => request<PageResult<Order>>(`/xy/order/list${queryString(input)}`),
  detail: (id: number) => request<Order>(`/xy/order/detail${queryString({ id })}`),
  syncPending: (accountId: number) => request<{ count: number }>("/xy/order/sync-pending", { method: "POST", body: JSON.stringify({ accountId }) }),
};
