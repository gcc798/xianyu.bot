export type Order = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  accountId: number;
  accountLabel: string;
  externalOrderId: string;
  externalProductId: string;
  productTitle: string;
  buyerUserId: string;
  buyerUserName: string;
  sessionId: string;
  skuKey: string;
  status: string;
  amountCent: number;
};

export type OrderSearch = {
  page: number;
  pageSize: number;
  keyword?: string;
  accountId?: number;
  status?: string;
};
