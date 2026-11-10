import {
  Alert,
  App,
  Button,
  Descriptions,
  Empty,
  Image,
  Space,
  Spin,
  Table,
  Tabs,
  Tag,
} from "antd";
import type { DescriptionsProps, TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { automationApi } from "../../../api/automation";
import { productsApi } from "../../../api/products";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { BusinessEvent } from "../../../models/automation";
import type { Product, ProductSettingInput } from "../../../models/product";
import pageStyles from "../../../styles/page.module.css";
import { formatDateTime, formatMoney } from "../../../utils/format";
import ProductSettingModal from "./components/ProductSettingModal";
import styles from "../../../styles/detail.module.css";

const productStatusColor: Record<string, string> = {
  on_sale: "success",
  offline: "default",
  sold: "warning",
};

const eventStatusColor: Record<string, string> = {
  success: "success",
  failed: "error",
  unknown: "warning",
};

export function Component() {
  const { id } = useParams();
  const productID = Number(id);
  const { message } = App.useApp();
  const { options: productStatuses } = useDictionaryOptions("product_status");
  const { options: eventStatuses } = useDictionaryOptions("business_event_status");
  const [product, setProduct] = useState<Product>();
  const [events, setEvents] = useState<BusinessEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!Number.isInteger(productID) || productID <= 0) {
      setError("商品 ID 无效");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    try {
      const [detail, eventPage] = await Promise.all([
        productsApi.detail(productID),
        automationApi.businessEvents({ page: 1, pageSize: 20, productId: productID }),
      ]);
      setProduct(detail);
      setEvents(eventPage.list);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "商品详情加载失败");
    } finally {
      setLoading(false);
    }
  }, [productID]);

  useEffect(() => {
    void load();
  }, [load]);

  const saveSetting = async (input: ProductSettingInput) => {
    setSaving(true);
    try {
      await productsApi.updateSetting({
        ...input,
        relistAddress: input.relistAddress.trim(),
        relistDescription: input.relistDescription.trim(),
        relistImageUrls: input.relistImageUrls.trim(),
      });
      message.success("商品设置已保存");
      setEditing(false);
      await load();
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <div className="app-loading">
        <Spin size="large" />
      </div>
    );

  if (!product) {
    return (
      <div className={pageStyles.page}>
        <Alert
          type="error"
          showIcon
          title={error || "商品不存在"}
          action={<Link to="/xianyu/products">返回商品列表</Link>}
        />
      </div>
    );
  }

  const statusColor = productStatusColor[product.status] ?? "default";
  const statusName = dictionaryName(productStatuses, product.status) || "未知";
  const overview: DescriptionsProps["items"] = [
    { key: "externalProductId", label: "闲鱼商品 ID", children: product.externalProductId },
    { key: "account", label: "所属账号", children: product.accountLabel },
    { key: "price", label: "销售价格", children: formatMoney(product.priceCent) },
    { key: "status", label: "商品状态", children: <Tag color={statusColor}>{statusName}</Tag> },
    { key: "type", label: "商品类型", children: product.virtualProduct ? "虚拟商品" : "实物商品" },
    { key: "updated", label: "最后同步", children: formatDateTime(product.UpdatedAt) },
  ];
  const automation: DescriptionsProps["items"] = [
    {
      key: "reply",
      label: "自动回复",
      children: (
        <AutomationStatus
          enabled={product.autoReplyEnabled}
          detail={`延迟 ${product.replyDelaySeconds} 秒`}
        />
      ),
    },
    {
      key: "delivery",
      label: "自动发货",
      children: <AutomationStatus enabled={product.autoDeliveryEnabled} detail="按发货规则执行" />,
    },
    {
      key: "relist",
      label: "自动续售",
      children: (
        <AutomationStatus
          enabled={product.autoRelistEnabled}
          detail={`延迟 ${product.relistDelaySeconds} 秒`}
        />
      ),
    },
    { key: "address", label: "续售所在地", children: product.relistAddress || "沿用原商品" },
    {
      key: "description",
      label: "续售描述",
      children: product.relistDescription || "沿用原商品",
      span: 2,
    },
  ];
  const eventColumns: TableColumnsType<BusinessEvent> = [
    { title: "发生时间", dataIndex: "CreatedAt", width: 160, render: formatDateTime },
    { title: "摘要", dataIndex: "summary", ellipsis: true },
    {
      title: "状态",
      dataIndex: "status",
      width: 110,
      render: (value: string) => {
        return (
          <Tag color={eventStatusColor[value] ?? "default"}>
            {dictionaryName(eventStatuses, value)}
          </Tag>
        );
      },
    },
    {
      title: "错误信息",
      dataIndex: "error",
      width: 260,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
  ];
  const images = product.imageUrls.length
    ? product.imageUrls
    : product.imageUrl
      ? [product.imageUrl]
      : [];

  return (
    <div className={pageStyles.page}>
      <section className={`${pageStyles.surface} ${styles.surface}`}>
        <div className={styles.heading}>
          <div>
            <Link to="/xianyu/products">返回商品列表</Link>
            <div className={styles.title}>
              <h1>{product.title}</h1>
              <Tag color={statusColor}>{statusName}</Tag>
            </div>
          </div>
          {product.status === "on_sale" && (
            <Button type="primary" onClick={() => setEditing(true)}>
              业务设置
            </Button>
          )}
        </div>
        {error && <Alert type="error" showIcon title={error} />}
        <Tabs
          className={styles.tabs}
          items={[
            {
              key: "overview",
              label: "商品概览",
              children: <Descriptions bordered column={2} size="small" items={overview} />,
            },
            {
              key: "images",
              label: "商品图片",
              children: images.length ? (
                <Image.PreviewGroup>
                  <div className={styles.images}>
                    {images.map((src, index) => (
                      <Image key={src} src={src} alt={`${product.title} ${index + 1}`} />
                    ))}
                  </div>
                </Image.PreviewGroup>
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无商品图片" />
              ),
            },
            {
              key: "automation",
              label: "自动化设置",
              children: <Descriptions bordered column={2} size="small" items={automation} />,
            },
            {
              key: "events",
              label: "操作记录",
              children: (
                <Table
                  rowKey="ID"
                  size="small"
                  columns={eventColumns}
                  dataSource={events}
                  pagination={false}
                  locale={{ emptyText: "暂无操作记录" }}
                  scroll={{ x: 760 }}
                />
              ),
            },
          ]}
        />
      </section>
      <ProductSettingModal
        product={editing ? product : null}
        loading={saving}
        onCancel={() => setEditing(false)}
        onSave={saveSetting}
      />
    </div>
  );
}

function AutomationStatus({ enabled, detail }: { enabled: boolean; detail: string }) {
  return (
    <Space size={8}>
      <Tag color={enabled ? "processing" : "default"}>{enabled ? "已开启" : "已关闭"}</Tag>
      {enabled && <span>{detail}</span>}
    </Space>
  );
}
