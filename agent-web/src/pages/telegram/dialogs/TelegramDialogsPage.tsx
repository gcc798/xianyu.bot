import { Alert, App, Button, Input, InputNumber, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import type { Key } from "react";
import { telegramApi } from "../../../api/telegram";
import type { TelegramDialog } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

export function Component() {
  const { message, modal } = App.useApp();
  const [rows, setRows] = useState<TelegramDialog[]>([]);
  const [selected, setSelected] = useState<Key[]>([]);
  const [page, setPage] = useState(1), [pageSize, setPageSize] = useState(30), [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState(""), [query, setQuery] = useState(""), [filter, setFilter] = useState<string>();
  const [membersMin, setMembersMin] = useState<number>(), [membersMax, setMembersMax] = useState<number>();
  const [loading, setLoading] = useState(false), [error, setError] = useState("");
  const load = useCallback(async () => {
    setLoading(true); setError("");
    try { const data = await telegramApi.dialogs({ page, page_size: pageSize, q: query, filter, members_min: membersMin, members_max: membersMax }); setRows(data.list); setTotal(data.total); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Telegram 会话加载失败"); }
    finally { setLoading(false); }
  }, [page, pageSize, query, filter, membersMin, membersMax]);
  useEffect(() => { void load(); }, [load]);

  const leave = () => modal.confirm({
    title: `退出选中的 ${selected.length} 个会话？`, content: "退出后将无法继续扫描这些群组或频道。",
    okText: "退出", okType: "danger", cancelText: "取消",
    onOk: async () => { const data = await telegramApi.leaveDialogs(selected.map(Number)); message.success(`成功 ${data.success}，失败 ${data.failed}`); setSelected([]); await load(); },
  });
  const columns: TableColumnsType<TelegramDialog> = [
    { title: "会话名称", dataIndex: "name", width: 220, ellipsis: true },
    { title: "用户名", dataIndex: "username", width: 150, ellipsis: true, render: (value: string | null) => value ? `@${value}` : "-" },
    { title: "类型", dataIndex: "dialog_type", width: 100, render: (value: string) => <Tag>{value}</Tag> },
    { title: "成员", dataIndex: "member_count", width: 90, render: (value: number | null) => value ?? "-" },
    { title: "未读", dataIndex: "unread_count", width: 80 },
    { title: "状态", dataIndex: "status", width: 90, render: (value: string) => <Tag color={value === "normal" ? "success" : "warning"}>{value === "normal" ? "正常" : value}</Tag> },
    { title: "最后消息", dataIndex: "last_message_at", width: 150, render: (value: string | null) => value ? formatDateTime(value) : "-" },
  ];
  const search = () => { setQuery(keyword.trim()); setPage(1); };
  const reset = () => { setKeyword(""); setQuery(""); setFilter(undefined); setMembersMin(undefined); setMembersMax(undefined); setPage(1); };
  return <div className={styles.page}><section className={styles.surface}>
    <div className={styles.filters}>
      <Input.Search value={keyword} allowClear placeholder="会话名称" style={{ width: 210 }} onChange={(event) => setKeyword(event.target.value)} onSearch={search} />
      <Select allowClear value={filter} placeholder="全部会话" style={{ width: 140 }} options={[{ value: "groups", label: "群组" }, { value: "channels", label: "频道" }, { value: "normal", label: "正常" }, { value: "invalid", label: "异常" }, { value: "muted", label: "已静音" }, { value: "archived", label: "已归档" }, { value: "creator", label: "我创建的" }]} onChange={(value) => { setFilter(value); setPage(1); }} />
      <InputNumber value={membersMin} min={0} placeholder="最少成员" onChange={(value) => setMembersMin(value ?? undefined)} />
      <InputNumber value={membersMax} min={0} placeholder="最多成员" onChange={(value) => setMembersMax(value ?? undefined)} />
      <Space><Button type="primary" onClick={search}>查询</Button><Button onClick={reset}>重置</Button><Button danger disabled={!selected.length} onClick={leave}>退出会话</Button></Space>
    </div>
    {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
    <div className={styles.table}><Table<TelegramDialog> rowKey="id" size="small" loading={loading} rowSelection={{ selectedRowKeys: selected, onChange: setSelected }} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1050 }} /></div>
    <div className={styles.pagination}><Pagination current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[30, 50, 100]} onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }} /></div>
  </section></div>;
}
