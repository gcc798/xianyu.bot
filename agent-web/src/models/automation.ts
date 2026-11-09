export type ReplyRule = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  productId: number;
  productTitle: string;
  type: string;
  keyword: string;
  priority: number;
  enabled: boolean;
  content: string;
};

export type ReplyRuleInput = Pick<
  ReplyRule,
  "ID" | "productId" | "type" | "keyword" | "priority" | "enabled" | "content"
>;

export type DeliveryRule = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  productId: number;
  productTitle: string;
  skuKey: string;
  skuName: string;
  mode: string;
  template: string;
  enabled: boolean;
};

export type DeliveryRuleInput = Pick<
  DeliveryRule,
  "ID" | "productId" | "skuKey" | "skuName" | "mode" | "template" | "enabled"
>;

export type DeliveryInventory = {
  ruleId: number;
  available: number;
  reserved: number;
  consumed: number;
};

export type DeliveryTask = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  accountId: number;
  accountLabel: string;
  externalOrderId: string;
  productTitle: string;
  buyerUserName: string;
  mode: string;
  status: string;
  attempts: number;
  nextAttemptAt: string;
  lastError: string;
  sentAt: string | null;
  confirmedAt: string | null;
};

export type BusinessEvent = {
  ID: number;
  CreatedAt: string;
  accountId: number;
  accountLabel: string;
  productId: number;
  productTitle: string;
  orderId: number;
  externalOrderId: string;
  type: string;
  status: string;
  summary: string;
  error: string;
};
