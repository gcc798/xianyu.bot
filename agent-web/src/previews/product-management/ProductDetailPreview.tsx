import {
  Button,
  Descriptions,
  Empty,
  Image,
  Space,
  Tabs,
  Tag,
  Timeline,
  Typography,
} from "antd";
import type { DescriptionsProps } from "antd";
import type { PreviewProduct } from "./previewData";

const productStatusMeta: Record<PreviewProduct["status"], { label: string; color?: string }> = {
  on_sale: { label: "在售", color: "success" },
  offline: { label: "已下架" },
  sold: { label: "已售出", color: "warning" },
};

export default function ProductDetailPreview({
  product,
  onBack,
  onOpenSetting,
  onSync,
}: {
  product: PreviewProduct;
  onBack: () => void;
  onOpenSetting: () => void;
  onSync: () => void;
}) {
  const status = productStatusMeta[product.status];
  const productFields: DescriptionsProps["items"] = [
    { key: "id", label: "闲鱼商品 ID", children: product.externalProductId },
    { key: "account", label: "所属账号", children: product.accountLabel },
    { key: "price", label: "销售价格", children: `¥${(product.priceCent / 100).toFixed(2)}` },
    { key: "status", label: "商品状态", children: <Tag color={status.color}>{status.label}</Tag> },
    { key: "type", label: "商品类型", children: product.virtualProduct ? "虚拟商品" : "实物商品" },
    { key: "updated", label: "最后同步", children: product.updatedAt },
  ];
  const automationFields: DescriptionsProps["items"] = [
    {
      key: "reply",
      label: "自动回复",
      children: (
        <AutomationValue
          enabled={product.autoReplyEnabled}
          detail={product.autoReplyEnabled ? `延迟 ${product.replyDelaySeconds} 秒` : "未启用回复规则"}
        />
      ),
    },
    {
      key: "delivery",
      label: "自动发货",
      children: (
        <AutomationValue
          enabled={product.autoDeliveryEnabled}
          detail={product.autoDeliveryEnabled ? (product.virtualProduct ? "虚拟商品自动交付" : "已启用发货规则") : "未启用发货规则"}
        />
      ),
    },
    {
      key: "relist",
      label: "自动续售",
      children: (
        <AutomationValue
          enabled={product.autoRelistEnabled}
          detail={product.autoRelistEnabled ? `售出后 ${product.relistDelaySeconds} 秒创建任务` : "未启用续售任务"}
        />
      ),
    },
  ];

  const eventItems = [
    { color: "green", children: `今天 10:42 · ${status.label}商品同步成功` },
    product.status === "on_sale"
      ? { color: "blue", children: "今天 10:29 · 商品状态保持在售" }
      : { color: "gray", children: `昨天 18:06 · 商品状态变更为${status.label}` },
    ...(product.autoReplyEnabled
      ? [{ color: "gray", children: "昨天 18:06 · 自动回复规则命中" }]
      : []),
  ];

  return (
    <div className="detail-page">
      <div className="detail-heading">
        <div>
          <Button type="link" className="back-button" onClick={onBack}>返回商品列表</Button>
          <div className="detail-title-row">
            <Typography.Title level={2}>{product.title}</Typography.Title>
            <Tag color={status.color}>{status.label}</Tag>
          </div>
        </div>
        <Space>
          <Button onClick={onSync}>同步商品</Button>
          <Button type="primary" disabled={product.status !== "on_sale"} onClick={onOpenSetting}>
            业务设置
          </Button>
        </Space>
      </div>

      <section className="detail-workspace">
        <Tabs
          className="detail-tabs"
          defaultActiveKey="product"
          items={[
            {
              key: "product",
              label: "商品信息",
              children: (
                <div className="detail-tab-body">
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>基础信息</Typography.Title>
                    </div>
                    <Descriptions bordered size="small" column={{ xs: 1, md: 2, xl: 3 }} items={productFields} />
                  </section>
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>自动化设置</Typography.Title>
                    </div>
                    <Descriptions bordered size="small" column={{ xs: 1, lg: 3 }} items={automationFields} />
                  </section>
                </div>
              ),
            },
            {
              key: "images",
              label: "商品图片",
              children: (
                <div className="detail-tab-body">
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>商品图片</Typography.Title>
                    </div>
                    <Image.PreviewGroup>
                      <div className="detail-image-grid">
                        {product.imageUrls.map((imageUrl, index) => (
                          <Image
                            key={imageUrl}
                            src={imageUrl}
                            alt={`${product.title}商品图片 ${index + 1}`}
                          />
                        ))}
                      </div>
                    </Image.PreviewGroup>
                  </section>
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>发布快照</Typography.Title>
                    </div>
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="发布快照仅在具备权限时展示" />
                  </section>
                </div>
              ),
            },
            {
              key: "records",
              label: "操作记录",
              children: (
                <div className="detail-tab-body">
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>操作记录</Typography.Title>
                    </div>
                    <Timeline className="detail-timeline" items={eventItems} />
                  </section>
                </div>
              ),
            },
          ]}
        />
      </section>
    </div>
  );
}

function AutomationValue({ enabled, detail }: { enabled: boolean; detail: string }) {
  return (
    <div className="automation-value">
      <span>{detail}</span>
      <Tag color={enabled ? "processing" : "default"}>{enabled ? "已开启" : "已关闭"}</Tag>
    </div>
  );
}
