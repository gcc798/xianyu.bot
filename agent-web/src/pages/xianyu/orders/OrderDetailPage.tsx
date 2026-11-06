import { Alert, Descriptions, Spin, Table, Tabs, Tag } from "antd";
import type { DescriptionsProps, TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { automationApi } from "../../../api/automation";
import { ordersApi } from "../../../api/orders";
import type { BusinessEvent } from "../../../models/automation";
import type { Order } from "../../../models/order";
import styles from "../../../styles/detail.module.css";
import pageStyles from "../../../styles/page.module.css";
import { formatDateTime, formatMoney } from "../../../utils/format";

const statusMeta: Record<string, { label: string; color: string }> = {
  pending_delivery: { label: "待发货", color: "warning" },
  shipped: { label: "已发货", color: "success" },
  unknown: { label: "待识别", color: "default" },
};

const eventStatusMeta: Record<string, { label: string; color: string }> = {
  success: { label: "成功", color: "success" },
  failed: { label: "失败", color: "error" },
  unknown: { label: "结果未知", color: "warning" },
};

export function Component() {
  const { id } = useParams();
  const orderID = Number(id);
  const [order, setOrder] = useState<Order>();
  const [events, setEvents] = useState<BusinessEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!Number.isInteger(orderID) || orderID <= 0) {
      setError("订单 ID 无效");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    try {
      const [detail, eventPage] = await Promise.all([
        ordersApi.detail(orderID),
        automationApi.businessEvents({ page: 1, pageSize: 20, orderId: orderID }),
      ]);
      setOrder(detail);
      setEvents(eventPage.list);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "订单详情加载失败");
    } finally {
      setLoading(false);
    }
  }, [orderID]);

  useEffect(() => { void load(); }, [load]);

  if (loading) return <div className="app-loading"><Spin size="large" /></div>;
  if (!order) {
    return <div className={pageStyles.page}><Alert type="error" showIcon title={error || "订单不存在"} action={<Link to="/xianyu/orders">返回订单列表</Link>} /></div>;
  }

  const status = statusMeta[order.status] ?? { label: order.status || "未知", color: "default" };
  const overview: DescriptionsProps["items"] = [
    { key: "externalOrderId", label: "闲鱼订单号", children: order.externalOrderId },
    { key: "status", label: "订单状态", children: <Tag color={status.color}>{status.label}</Tag> },
    { key: "product", label: "成交商品", children: order.productTitle },
    { key: "productId", label: "闲鱼商品 ID", children: order.externalProductId || "-" },
    { key: "sku", label: "成交规格", children: order.skuKey || "-" },
    { key: "amount", label: "成交金额", children: formatMoney(order.amountCent) },
    { key: "account", label: "所属账号", children: order.accountLabel },
    { key: "created", label: "首次同步", children: formatDateTime(order.CreatedAt) },
    { key: "updated", label: "最后更新", children: formatDateTime(order.UpdatedAt) },
  ];
  const buyer: DescriptionsProps["items"] = [
    { key: "buyerName", label: "买家名称", children: order.buyerUserName || "-" },
    { key: "buyerId", label: "买家 ID", children: order.buyerUserId || "-" },
    { key: "session", label: "会话 ID", children: order.sessionId || "-", span: 2 },
  ];
  const eventColumns: TableColumnsType<BusinessEvent> = [
    { title: "发生时间", dataIndex: "CreatedAt", width: 160, render: formatDateTime },
    { title: "摘要", dataIndex: "summary", ellipsis: true },
    { title: "状态", dataIndex: "status", width: 110, render: (value: string) => { const item = eventStatusMeta[value] ?? { label: value, color: "default" }; return <Tag color={item.color}>{item.label}</Tag>; } },
    { title: "错误信息", dataIndex: "error", width: 260, ellipsis: true, render: (value: string) => value || "-" },
  ];

  return (
    <div className={pageStyles.page}>
      <section className={`${pageStyles.surface} ${styles.surface}`}>
        <div className={styles.heading}>
          <div>
            <Link to="/xianyu/orders">返回订单列表</Link>
            <div className={styles.title}><h1>{order.externalOrderId}</h1><Tag color={status.color}>{status.label}</Tag></div>
          </div>
        </div>
        {error && <Alert type="error" showIcon title={error} />}
        <Tabs
          className={styles.tabs}
          items={[
            { key: "overview", label: "订单信息", children: <Descriptions bordered column={2} size="small" items={overview} /> },
            { key: "buyer", label: "买家与会话", children: <Descriptions bordered column={2} size="small" items={buyer} /> },
            { key: "events", label: "操作记录", children: <Table rowKey="ID" size="small" columns={eventColumns} dataSource={events} pagination={false} locale={{ emptyText: "暂无操作记录" }} scroll={{ x: 760 }} /> },
          ]}
        />
      </section>
    </div>
  );
}
