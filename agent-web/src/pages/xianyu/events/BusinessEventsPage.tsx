import { Alert, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { accountsApi } from "../../../api/accounts";
import { automationApi } from "../../../api/automation";
import type { XianyuAccount } from "../../../models/account";
import type { BusinessEvent } from "../../../models/automation";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

const eventTypes: Record<string, string> = {
  message: "消息接收",
  delivery: "自动发货",
  relist: "自动续售",
  order_sync: "订单同步",
  product_sync: "商品同步",
  login_check: "登录检查",
};

const eventStatuses: Record<string, { label: string; color: string }> = {
  success: { label: "成功", color: "success" },
  failed: { label: "失败", color: "error" },
  unknown: { label: "结果未知", color: "warning" },
};

export function Component() {
  const [rows, setRows] = useState<BusinessEvent[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number>();
  const [type, setType] = useState<string>();
  const [status, setStatus] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await automationApi.businessEvents({ page, pageSize, accountId, type, status, keyword: query });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "业务事件加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, type, status, query]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => {
    void accountsApi.list({ page: 1, pageSize: 100 })
      .then((data) => setAccounts(data.list))
      .catch(() => undefined);
  }, []);

  const columns: TableColumnsType<BusinessEvent> = [
    { title: "发生时间", dataIndex: "CreatedAt", width: 150, render: formatDateTime },
    { title: "账号", dataIndex: "accountLabel", width: 120, ellipsis: true, render: (value: string) => value || "-" },
    {
      title: "类型", dataIndex: "type", width: 110,
      render: (value: string) => <Tag>{eventTypes[value] ?? value}</Tag>,
    },
    {
      title: "状态", dataIndex: "status", width: 100,
      render: (value: string) => {
        const item = eventStatuses[value] ?? { label: value, color: "default" };
        return <Tag color={item.color}>{item.label}</Tag>;
      },
    },
    { title: "摘要", dataIndex: "summary", width: 250, ellipsis: true },
    { title: "商品", dataIndex: "productTitle", width: 180, ellipsis: true, render: (value: string) => value || "-" },
    { title: "订单号", dataIndex: "externalOrderId", width: 160, ellipsis: true, render: (value: string) => value || "-" },
    { title: "错误信息", dataIndex: "error", width: 240, ellipsis: true, render: (value: string) => value || "-" },
  ];

  const search = () => { setQuery(keyword.trim()); setPage(1); };
  const reset = () => {
    setAccountId(undefined);
    setType(undefined);
    setStatus(undefined);
    setKeyword("");
    setQuery("");
    setPage(1);
  };

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Select
            allowClear value={accountId} placeholder="全部账号" style={{ width: 160 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => { setAccountId(value); setPage(1); }}
          />
          <Select
            allowClear value={type} placeholder="全部类型" style={{ width: 140 }}
            options={Object.entries(eventTypes).map(([value, label]) => ({ value, label }))}
            onChange={(value) => { setType(value); setPage(1); }}
          />
          <Select
            allowClear value={status} placeholder="全部状态" style={{ width: 130 }}
            options={Object.entries(eventStatuses).map(([value, item]) => ({ value, label: item.label }))}
            onChange={(value) => { setStatus(value); setPage(1); }}
          />
          <Input.Search
            value={keyword} allowClear placeholder="摘要、错误或关联数据" style={{ width: 240 }}
            onChange={(event) => setKeyword(event.target.value)} onSearch={search}
          />
          <Space>
            <Button type="primary" onClick={search}>查询</Button>
            <Button onClick={reset}>重置</Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<BusinessEvent>
            rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows}
            pagination={false} scroll={{ x: 1310 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }}
          />
        </div>
      </section>
    </div>
  );
}
