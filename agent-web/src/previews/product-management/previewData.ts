export type ProductStatus = "on_sale" | "offline" | "sold";

// 字段与后端 response.Product 保持一致，便于选定方案后迁移到真实接口。
export type PreviewProduct = {
  ID: number;
  accountId: number;
  accountLabel: string;
  externalProductId: string;
  title: string;
  priceCent: number;
  status: ProductStatus;
  imageUrls: string[]; // 预览专用图片，不映射真实商品接口。
  virtualProduct: boolean;
  autoReplyEnabled: boolean;
  autoDeliveryEnabled: boolean;
  autoRelistEnabled: boolean;
  replyDelaySeconds: number;
  relistDelaySeconds: number;
  relistDescription: string;
  relistImageUrls: string;
  relistAddress: string;
  updatedAt: string;
};

export const previewAccounts = [
  { label: "数码严选店", value: 11 },
  { label: "青屿书房", value: 12 },
  { label: "轻物研究所", value: 13 },
];

// 全部方案共用同一份明确标注的模拟数据，避免视觉差异被内容差异干扰。
export const previewProducts: PreviewProduct[] = [
  {
    ID: 101,
    accountId: 11,
    accountLabel: "数码严选店",
    externalProductId: "873291604521",
    title: "Apple 妙控键盘 2 代｜深空灰｜成色 95 新",
    priceCent: 46800,
    status: "on_sale",
    imageUrls: [
      "https://images.unsplash.com/photo-1587829741301-dc798b83add3?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1496181133206-80ce9b88a853?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: false,
    autoReplyEnabled: true,
    autoDeliveryEnabled: false,
    autoRelistEnabled: true,
    replyDelaySeconds: 3,
    relistDelaySeconds: 60,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "杭州市西湖区",
    updatedAt: "今天 10:42",
  },
  {
    ID: 102,
    accountId: 12,
    accountLabel: "青屿书房",
    externalProductId: "872944103807",
    title: "设计中的设计｜原研哉｜正版精装版",
    priceCent: 5900,
    status: "on_sale",
    imageUrls: [
      "https://images.unsplash.com/photo-1544947950-fa07a98d237f?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1524995997946-a1c2e315a42f?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1495446815901-a7297e633e8d?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: false,
    autoReplyEnabled: true,
    autoDeliveryEnabled: true,
    autoRelistEnabled: false,
    replyDelaySeconds: 5,
    relistDelaySeconds: 60,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "",
    updatedAt: "今天 09:18",
  },
  {
    ID: 103,
    accountId: 13,
    accountLabel: "轻物研究所",
    externalProductId: "872516830429",
    title: "Notion 商业运营模板｜含 12 个数据库",
    priceCent: 3900,
    status: "on_sale",
    imageUrls: [
      "https://images.unsplash.com/photo-1496181133206-80ce9b88a853?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1552664730-d307ca884978?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: true,
    autoReplyEnabled: true,
    autoDeliveryEnabled: true,
    autoRelistEnabled: true,
    replyDelaySeconds: 0,
    relistDelaySeconds: 30,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "上海市静安区",
    updatedAt: "昨天 23:06",
  },
  {
    ID: 104,
    accountId: 11,
    accountLabel: "数码严选店",
    externalProductId: "871993602118",
    title: "SONY WH-1000XM5 无线降噪耳机",
    priceCent: 169900,
    status: "offline",
    imageUrls: [
      "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1484704849700-f032a568e944?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1545127398-14699f92334b?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: false,
    autoReplyEnabled: false,
    autoDeliveryEnabled: false,
    autoRelistEnabled: false,
    replyDelaySeconds: 0,
    relistDelaySeconds: 60,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "",
    updatedAt: "11 月 1 日",
  },
  {
    ID: 105,
    accountId: 12,
    accountLabel: "青屿书房",
    externalProductId: "869421570306",
    title: "Kinfolk 生活美学杂志 2024 秋季刊",
    priceCent: 8600,
    status: "sold",
    imageUrls: [
      "https://images.unsplash.com/photo-1504711434969-e33886168f5c?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1516979187457-637abb4f9353?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1543002588-bfa74002ed7e?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: false,
    autoReplyEnabled: true,
    autoDeliveryEnabled: false,
    autoRelistEnabled: true,
    replyDelaySeconds: 8,
    relistDelaySeconds: 120,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "宁波市鄞州区",
    updatedAt: "10 月 29 日",
  },
  {
    ID: 106,
    accountId: 13,
    accountLabel: "轻物研究所",
    externalProductId: "868730145902",
    title: "小型团队客服话术库｜可直接导入",
    priceCent: 2900,
    status: "on_sale",
    imageUrls: [
      "https://images.unsplash.com/photo-1556761175-b413da4baf72?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1521737711867-e3b97375f902?auto=format&fit=max&w=1600&q=85",
      "https://images.unsplash.com/photo-1556761175-4b46a572b786?auto=format&fit=max&w=1600&q=85",
    ],
    virtualProduct: true,
    autoReplyEnabled: false,
    autoDeliveryEnabled: true,
    autoRelistEnabled: false,
    replyDelaySeconds: 0,
    relistDelaySeconds: 60,
    relistDescription: "",
    relistImageUrls: "",
    relistAddress: "",
    updatedAt: "10 月 28 日",
  },
];
