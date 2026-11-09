export type Product = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  accountId: number;
  accountLabel: string;
  externalProductId: string;
  title: string;
  priceCent: number;
  status: string;
  imageUrl: string;
  imageUrls: string[];
  virtualProduct: boolean;
  autoReplyEnabled: boolean;
  autoDeliveryEnabled: boolean;
  autoRelistEnabled: boolean;
  replyDelaySeconds: number;
  relistDelaySeconds: number;
  relistDescription: string;
  relistImageUrls: string;
  relistAddress: string;
};

export type ProductSearch = {
  page: number;
  pageSize: number;
  accountId?: number;
  titleOrId?: string;
  status?: string;
};

export type ProductSettingInput = {
  productId: number;
  virtualProduct: boolean;
  autoReplyEnabled: boolean;
  autoDeliveryEnabled: boolean;
  autoRelistEnabled: boolean;
  replyDelaySeconds: number;
  relistDelaySeconds: number;
  relistDescription: string;
  relistImageUrls: string;
  relistAddress: string;
};
