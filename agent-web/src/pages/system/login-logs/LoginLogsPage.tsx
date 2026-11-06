import { Alert, App, Button, Input, Pagination, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { systemApi } from "../../../api/system";
import type { LoginLog } from "../../../models/system";
import styles from "../../../styles/page.module.css";

export function Component() {
  const { message, modal } = App.useApp();
  const [rows, setRows] = useState<LoginLog[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await systemApi.loginLogs(page, 20, query);
      setRows(response.list);
      setTotal(response.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "登录日志加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, query]);

  useEffect(() => { void load(); }, [load]);

  const remove = (record: LoginLog) => modal.confirm({
    title: "删除登录日志？",
    content: "删除后无法恢复。",
    okText: "删除",
    okType: "danger",
    cancelText: "取消",
    onOk: async () => {
      await systemApi.deleteLoginLog(record.ID);
      message.success("登录日志已删除");
      await load();
    },
  });

  const columns: TableColumnsType<LoginLog> = [
    { title: "序号", width: 64, render: (_value, _record, index) => (page - 1) * 20 + index + 1 },
    { title: "用户", dataIndex: "username", width: 160 },
    { title: "IP", dataIndex: "ip", width: 150 },
    { title: "结果", dataIndex: "status", width: 90, render: (value: boolean) => <Tag color={value ? "success" : "error"}>{value ? "成功" : "失败"}</Tag> },
    { title: "说明", dataIndex: "errorMessage", width: 220, ellipsis: true },
    { title: "客户端", dataIndex: "agent", ellipsis: true },
    { title: "时间", dataIndex: "CreatedAt", width: 180, render: (value: string) => new Intl.DateTimeFormat("zh-CN", { dateStyle: "short", timeStyle: "medium" }).format(new Date(value)) },
    { title: "操作", key: "actions", width: 80, render: (_, record) => <Space><Button type="link" danger className={styles.linkButton} onClick={() => remove(record)}>删除</Button></Space> },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Input.Search value={keyword} allowClear placeholder="用户名" style={{ width: 280 }} onChange={(event) => setKeyword(event.target.value)} onSearch={() => { setPage(1); setQuery(keyword.trim()); }} />
          <Button onClick={() => { setKeyword(""); setQuery(""); setPage(1); }}>重置</Button>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<LoginLog> rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1050 }} />
        </div>
        <div className={styles.pagination}><Pagination current={page} pageSize={20} total={total} showSizeChanger={false} onChange={setPage} /></div>
      </section>
    </div>
  );
}
