import { Alert, App, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { accountsApi } from "../../../api/accounts";
import { automationApi } from "../../../api/automation";
import type { XianyuAccount } from "../../../models/account";
import type { DeliveryTask } from "../../../models/automation";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

const taskStatuses: Record<string, { label: string; color: string }> = {
  pending: { label: "等待执行", color: "default" },
  sending: { label: "发送中", color: "processing" },
  sent_unconfirmed: { label: "待确认", color: "warning" },
  failed: { label: "失败", color: "error" },
  unknown: { label: "结果未知", color: "warning" },
  confirmed: { label: "已确认", color: "success" },
};

export function Component() {
  const { message, modal } = App.useApp();
  const [rows, setRows] = useState<DeliveryTask[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number>();
  const [status, setStatus] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await automationApi.deliveryTasks({
        page,
        pageSize,
        accountId,
        status,
        keyword: query,
      });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "发货任务加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, status, query]);

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    void accountsApi
      .list({ page: 1, pageSize: 100 })
      .then((data) => setAccounts(data.list))
      .catch(() => undefined);
  }, []);

  const retry = (record: DeliveryTask) => {
    modal.confirm({
      title: "重试发货任务？",
      content: "任务将重新进入等待队列。",
      okText: "重试",
      cancelText: "取消",
      onOk: async () => {
        await automationApi.retryDeliveryTask(record.ID);
        message.success("任务已进入等待队列");
        await load();
      },
    });
  };

  const columns: TableColumnsType<DeliveryTask> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    { title: "订单号", dataIndex: "externalOrderId", width: 160, ellipsis: true },
    { title: "商品", dataIndex: "productTitle", width: 200, ellipsis: true },
    { title: "买家", dataIndex: "buyerUserName", width: 110, ellipsis: true },
    { title: "账号", dataIndex: "accountLabel", width: 120, ellipsis: true },
    {
      title: "模式",
      dataIndex: "mode",
      width: 90,
      render: (value: string) => (
        <Tag color={value === "card" ? "processing" : "default"}>
          {value === "card" ? "卡密" : "文本"}
        </Tag>
      ),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value: string) => {
        const item = taskStatuses[value] ?? { label: value, color: "default" };
        return <Tag color={item.color}>{item.label}</Tag>;
      },
    },
    { title: "尝试次数", dataIndex: "attempts", width: 90 },
    { title: "下次执行", dataIndex: "nextAttemptAt", width: 150, render: formatDateTime },
    {
      title: "失败原因",
      dataIndex: "lastError",
      width: 200,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "操作",
      width: 70,
      fixed: "right",
      render: (_, record) =>
        record.status === "failed" ? (
          <Button type="link" className={styles.linkButton} onClick={() => retry(record)}>
            重试
          </Button>
        ) : (
          "-"
        ),
    },
  ];

  const search = () => {
    setQuery(keyword.trim());
    setPage(1);
  };
  const reset = () => {
    setAccountId(undefined);
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
            allowClear
            value={accountId}
            placeholder="全部账号"
            style={{ width: 170 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => {
              setAccountId(value);
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={status}
            placeholder="全部状态"
            style={{ width: 140 }}
            options={Object.entries(taskStatuses).map(([value, item]) => ({
              value,
              label: item.label,
            }))}
            onChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
          />
          <Input.Search
            value={keyword}
            allowClear
            placeholder="订单、商品或买家"
            style={{ width: 250 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={search}
          />
          <Space>
            <Button type="primary" onClick={search}>
              查询
            </Button>
            <Button onClick={reset}>重置</Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<DeliveryTask>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1380 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => {
              setPage(nextSize === pageSize ? nextPage : 1);
              setPageSize(nextSize);
            }}
          />
        </div>
      </section>
    </div>
  );
}
