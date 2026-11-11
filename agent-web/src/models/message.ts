export type XianyuMessage = {
  ID: number;
  CreatedAt: string;
  accountId: number;
  accountLabel: string;
  pnmId: string;
  sessionId: string;
  contentType: number;
  content: string;
  senderUserName: string;
  senderUserId: string;
  externalProductId: string;
  externalOrderId: string;
  messageTime: number;
  direction: string;
  source: string;
  takeoverActive: boolean;
};

export type MessageSearch = {
  page: number;
  pageSize: number;
  accountId?: number;
  productId?: number;
  sessionId?: string;
  direction?: string;
  externalProductId?: string;
  buyerName?: string;
  content?: string;
  startTime?: number;
  endTime?: number;
};
