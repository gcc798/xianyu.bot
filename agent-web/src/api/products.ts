import type { EmptyResponse, PageResult } from "../models/http";
import type { Product, ProductSearch, ProductSettingInput } from "../models/product";
import { queryString, request } from "./client";

export const productsApi = {
  list: (input: ProductSearch) =>
    request<PageResult<Product>>(`/xy/product/list${queryString(input)}`),
  detail: (id: number) => request<Product>(`/xy/product/detail${queryString({ id })}`),
  sync: (accountId: number) =>
    request<{ count: number }>("/xy/product/sync", {
      method: "POST",
      body: JSON.stringify({ accountId }),
    }),
  updateSetting: (input: ProductSettingInput) =>
    request<EmptyResponse>("/xy/product/setting", { method: "PUT", body: JSON.stringify(input) }),
};
