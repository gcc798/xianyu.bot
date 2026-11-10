import { Alert, Button, Input, InputNumber, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { accountsApi } from "../../../api/accounts";
import { automationApi } from "../../../api/automation";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { XianyuAccount } from "../../../models/account";
import type { BusinessEvent } from "../../../models/automation";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

const eventStatusColor: Record<string, string> = {
  success: "success",
  failed: "error",
  unknown: "warning",
};

export function Component() {
  const { options: eventTypes } = useDictionaryOptions("business_event_type");
  const { options: eventStatuses } = useDictionaryOptions("business_event_status");
  const [rows, setRows] = useState<BusinessEvent[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number>();
  const [productId, setProductId] = useState<number>();
  const [orderId, setOrderId] = useState<number>();
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
      const data = await automationApi.businessEvents({
        page,
        pageSize,
        accountId,
        productId,
        orderId,
        type,
        status,
        keyword: query,
      });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "业务事件加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, productId, orderId, type, status, query]);

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    void accountsApi
      .list({ page: 1, pageSize: 100 })
      .then((data) => setAccounts(data.list))
      .catch(() => undefined);
  }, []);

  const columns: TableColumnsType<BusinessEvent> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    { title: "发生时间", dataIndex: "CreatedAt", width: 150, render: formatDateTime },
    {
      title: "账号",
      dataIndex: "accountLabel",
      width: 120,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "类型",
      dataIndex: "type",
      width: 110,
      render: (value: string) => <Tag>{dictionaryName(eventTypes, value)}</Tag>,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value: string) => (
        <Tag color={eventStatusColor[value] ?? "default"}>
          {dictionaryName(eventStatuses, value)}
        </Tag>
      ),
    },
    { title: "摘要", dataIndex: "summary", width: 250, ellipsis: true },
    {
      title: "商品",
      dataIndex: "productTitle",
      width: 180,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "订单号",
      dataIndex: "externalOrderId",
      width: 160,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "错误信息",
      dataIndex: "error",
      width: 240,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
  ];

  const search = () => {
    setQuery(keyword.trim());
    setPage(1);
  };
  const reset = () => {
    setAccountId(undefined);
    setProductId(undefined);
    setOrderId(undefined);
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
            allowClear
            value={accountId}
            placeholder="全部账号"
            style={{ width: 160 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => {
              setAccountId(value);
              setPage(1);
            }}
          />
          <InputNumber
            value={productId}
            min={1}
            placeholder="商品 ID"
            style={{ width: 120 }}
            onChange={(value) => {
              setProductId(value ?? undefined);
              setPage(1);
            }}
          />
          <InputNumber
            value={orderId}
            min={1}
            placeholder="订单 ID"
            style={{ width: 120 }}
            onChange={(value) => {
              setOrderId(value ?? undefined);
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={type}
            placeholder="全部类型"
            style={{ width: 140 }}
            options={eventTypes.map(({ code, name }) => ({ value: code, label: name }))}
            onChange={(value) => {
              setType(value);
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={status}
            placeholder="全部状态"
            style={{ width: 130 }}
            options={eventStatuses.map(({ code, name }) => ({ value: code, label: name }))}
            onChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
          />
          <Input.Search
            value={keyword}
            allowClear
            placeholder="摘要、错误或关联数据"
            style={{ width: 240 }}
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
          <Table<BusinessEvent>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1310 }}
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
