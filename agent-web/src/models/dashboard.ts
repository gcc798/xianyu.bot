import type { BusinessEvent } from "./automation";

export type DashboardMetrics = {
  accounts: number;
  readyAccounts: number;
  onSaleProducts: number;
  inboundMessages: number;
  outboundMessages: number;
  autoReplies: number;
  autoReplyRate: number;
  newOrders: number;
  pendingOrders: number;
  failedAutomationTasks: number;
};

export type DashboardMessagePoint = {
  timestamp: number;
  inbound: number;
  outbound: number;
  autoReplies: number;
};
export type DashboardOrderPoint = {
  timestamp: number;
  orders: number;
  delivered: number;
  failed: number;
};
export type DashboardNameValue = { name: string; value: number };
export type DashboardAutomation = {
  replySuccess: number;
  replyFailed: number;
  deliverySuccess: number;
  deliveryFailed: number;
  relistSuccess: number;
  relistFailed: number;
};

export type DashboardOverview = {
  startTime: number;
  endTime: number;
  bucketMillis: number;
  metrics: DashboardMetrics;
  messageTrend: DashboardMessagePoint[];
  messageSources: DashboardNameValue[];
  orderTrend: DashboardOrderPoint[];
  automation: DashboardAutomation;
  recentEvents: BusinessEvent[];
};
