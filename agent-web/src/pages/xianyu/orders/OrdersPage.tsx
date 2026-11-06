import { Alert, App, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { accountsApi } from "../../../api/accounts";
import { ordersApi } from "../../../api/orders";
import type { XianyuAccount } from "../../../models/account";
import type { Order } from "../../../models/order";
import { formatDateTime, formatMoney } from "../../../utils/format";
import pageStyles from "../../../styles/page.module.css";

const orderStatus: Record<string, { label: string; color: string }> = {
  pending_delivery: { label: "待发货", color: "warning" },
  shipped: { label: "已发货", color: "success" },
  unknown: { label: "待识别", color: "default" },
};

export function Component() {
  const { message } = App.useApp();
  const [rows, setRows] = useState<Order[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number>();
  const [status, setStatus] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await ordersApi.list({ page, pageSize, accountId, status, keyword: query });
      setRows(response.list);
      setTotal(response.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "订单加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, status, query]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => { void accountsApi.list({ page: 1, pageSize: 100 }).then((response) => setAccounts(response.list)).catch((cause) => setError(cause instanceof Error ? cause.message : "账号加载失败")); }, []);

  const syncPending = async () => {
    if (!accountId) { message.warning("请先选择一个闲鱼账号"); return; }
    setSyncing(true);
    try {
      const response = await ordersApi.syncPending(accountId);
      message.success(`已同步 ${response.count} 条待发货订单`);
      await load();
    } finally {
      setSyncing(false);
    }
  };

  const columns: TableColumnsType<Order> = [
    { title: "订单号", dataIndex: "externalOrderId", width: 160, ellipsis: true, render: (value: string, record) => <Link to={`/xianyu/orders/${record.ID}`}>{value}</Link> },
    { title: "商品", dataIndex: "productTitle", width: 190, ellipsis: true },
    { title: "商品编码", dataIndex: "externalProductId", width: 125, ellipsis: true },
    {
      title: "买家", dataIndex: "buyerUserName", width: 110, ellipsis: true,
      render: (value: string, record) => value || record.buyerUserId || "-",
    },
    { title: "账号", dataIndex: "accountLabel", width: 105, ellipsis: true },
    { title: "规格", dataIndex: "skuKey", width: 100, ellipsis: true, render: (value: string) => value || "-" },
    { title: "金额", dataIndex: "amountCent", width: 80, render: formatMoney },
    {
      title: "状态", dataIndex: "status", width: 85,
      render: (value: string) => {
        const item = orderStatus[value] || { label: value || "未知", color: "default" };
        return <Tag color={item.color}>{item.label}</Tag>;
      },
    },
    { title: "更新时间", dataIndex: "UpdatedAt", width: 140, render: formatDateTime },
    { title: "操作", key: "operation", width: 70, fixed: "right", render: (_, record) => <Link to={`/xianyu/orders/${record.ID}`}>查看</Link> },
  ];

  const reset = () => { setAccountId(undefined); setStatus(undefined); setKeyword(""); setQuery(""); setPage(1); };
  return (
    <div className={pageStyles.page}>
      <section className={pageStyles.surface}>
        <div className={pageStyles.filters}>
          <Select
            allowClear value={accountId} placeholder="全部账号" style={{ width: 170 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => { setAccountId(value); setPage(1); }}
          />
          <Select
            allowClear value={status} placeholder="全部状态" style={{ width: 140 }}
            options={[{ value: "pending_delivery", label: "待发货" }, { value: "shipped", label: "已发货" }, { value: "unknown", label: "待识别" }]}
            onChange={(value) => { setStatus(value); setPage(1); }}
          />
          <Input.Search
            value={keyword} allowClear placeholder="订单、商品或买家" style={{ width: 250 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={() => { setQuery(keyword.trim()); setPage(1); }}
          />
          <Space>
            <Button type="primary" onClick={() => { setQuery(keyword.trim()); setPage(1); }}>查询</Button>
            <Button onClick={reset}>重置</Button>
            <Button type="primary" loading={syncing} onClick={() => void syncPending()}>同步待发货</Button>
          </Space>
        </div>
        {error && <Alert className={pageStyles.alert} type="error" showIcon title={error} closable onClose={() => setError("")} />}
        <div className={pageStyles.table}>
          <Table<Order> rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1165 }} />
        </div>
        <div className={pageStyles.pagination}>
          <Pagination
            current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }}
          />
        </div>
      </section>
    </div>
  );
}
