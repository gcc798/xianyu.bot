import type { DashboardOverview } from "../models/dashboard";
import { queryString, request } from "./client";

export const dashboardApi = {
  overview: (input: { accountId?: number; startTime: number; endTime: number }) =>
    request<DashboardOverview>(`/xy/dashboard/overview${queryString(input)}`),
};
