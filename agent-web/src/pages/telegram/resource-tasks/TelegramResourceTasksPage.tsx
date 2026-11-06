import { Alert, App, Button, Form, Input, InputNumber, Pagination, Radio, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { telegramApi } from "../../../api/telegram";
import type { TelegramResourceOptions, TelegramResourceTask, TelegramResourceTaskInput } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

const statusMeta: Record<TelegramResourceTask["status"], { label: string; color?: string }> = {
  queued: { label: "排队中", color: "processing" },
  running: { label: "执行中", color: "processing" },
  completed: { label: "已完成", color: "success" },
  failed: { label: "失败", color: "error" },
  cancelled: { label: "已取消" },
};

const sourceModeLabel = (task: TelegramResourceTask) => {
  if (task.source_mode === "all") return "全部来源";
  if (task.source_mode === "manual") return "手动选择";
  return task.source_strategy === "ai" ? "智能 · AI" : "智能 · 本地";
};

export function Component() {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<TelegramResourceTaskInput>();
  const sourceMode = Form.useWatch("source_mode", form);
  const [options, setOptions] = useState<TelegramResourceOptions>({ dialogs: [], file_types: [], max_file_size: 0 });
  const [rows, setRows] = useState<TelegramResourceTask[]>([]);
  const [page, setPage] = useState(1), [pageSize, setPageSize] = useState(20), [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false), [submitting, setSubmitting] = useState(false), [error, setError] = useState("");

  const load = useCallback(async (quiet = false) => {
    if (!quiet) setLoading(true);
    setError("");
    try {
      const data = await telegramApi.resourceTasks({ page, page_size: pageSize });
      setRows(data.list); setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "资源任务加载失败");
    } finally {
      if (!quiet) setLoading(false);
    }
  }, [page, pageSize]);

  useEffect(() => {
    void telegramApi.resourceOptions().then(setOptions).catch((cause) => setError(cause instanceof Error ? cause.message : "资源选项加载失败"));
  }, []);
  useEffect(() => {
    void load();
    const timer = window.setInterval(() => void load(true), 5000);
    return () => window.clearInterval(timer);
  }, [load]);

  const create = async (values: TelegramResourceTaskInput) => {
    setSubmitting(true);
    try {
      await telegramApi.createResourceTask({ ...values, keyword: values.keyword.trim() });
      message.success("任务已进入队列");
      form.resetFields(["keyword"]);
      setPage(1);
      await load();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "任务创建失败");
    } finally {
      setSubmitting(false);
    }
  };

  const runAction = (task: TelegramResourceTask, action: "cancel" | "retry") => modal.confirm({
    title: action === "cancel" ? `取消任务 #${task.id}？` : `重新执行任务 #${task.id}？`,
    content: action === "cancel" ? "正在下载或上传的文件将停止处理。" : "失败或已取消的任务会重新进入队列。",
    okText: action === "cancel" ? "取消任务" : "重新执行",
    okType: action === "cancel" ? "danger" : "primary",
    cancelText: "返回",
    onOk: async () => {
      if (action === "cancel") await telegramApi.cancelResourceTask(task.id); else await telegramApi.retryResourceTask(task.id);
      message.success(action === "cancel" ? "任务已取消" : "任务已重新排队");
      await load();
    },
  });

  const columns: TableColumnsType<TelegramResourceTask> = [
    { title: "任务", dataIndex: "id", width: 76, render: (value: number) => `#${value}` },
    { title: "检索主题", dataIndex: "keyword", width: 210, ellipsis: true },
    { title: "文件类型", dataIndex: "file_types", width: 150, ellipsis: true, render: (value: string[]) => value.length ? value.join("、") : "不限" },
    { title: "范围", width: 150, render: (_, row) => `${sourceModeLabel(row)} · ${row.source_dialog_ids.length} 个` },
    { title: "进度", width: 160, render: (_, row) => `${row.files_completed} / ${row.max_downloads}（发现 ${row.files_found}）` },
    { title: "状态", dataIndex: "status", width: 100, render: (value: TelegramResourceTask["status"]) => <Tag color={statusMeta[value].color}>{statusMeta[value].label}</Tag> },
    { title: "创建时间", dataIndex: "created_at", width: 150, render: (value: string) => formatDateTime(value) },
    { title: "错误", dataIndex: "error_message", width: 180, ellipsis: true, render: (value: string | null) => value || "-" },
    { title: "操作", fixed: "right", width: 90, render: (_, row) => <>
      {(["queued", "running"] as const).includes(row.status as "queued" | "running") && <Button danger type="link" className={styles.linkButton} onClick={() => runAction(row, "cancel")}>取消</Button>}
      {(["failed", "cancelled"] as const).includes(row.status as "failed" | "cancelled") && <Button type="link" className={styles.linkButton} onClick={() => runAction(row, "retry")}>重试</Button>}
    </> },
  ];

  return <div className={styles.page}><section className={styles.surface}>
    <Form form={form} layout="inline" className={styles.filters} initialValues={{ file_types: [], source_mode: "smart", source_dialog_ids: [], max_downloads: 10 }} onFinish={create}>
      <Form.Item name="keyword" label="资源名称" rules={[{ required: true, whitespace: true, min: 2, max: 200 }]}><Input maxLength={200} placeholder="输入资源主题" style={{ width: 220 }} /></Form.Item>
      <Form.Item name="file_types" label="文件类型"><Select mode="multiple" allowClear maxTagCount="responsive" placeholder="不限" style={{ width: 190 }} options={options.file_types.map((value) => ({ value, label: value }))} /></Form.Item>
      <Form.Item name="max_downloads" label="最大数量" rules={[{ required: true }]}><InputNumber min={1} max={100} style={{ width: 100 }} /></Form.Item>
      <Form.Item name="source_mode" label="搜索方式"><Radio.Group optionType="button" buttonStyle="solid" options={[{ value: "smart", label: "智能" }, { value: "all", label: "全部" }, { value: "manual", label: "手动" }]} /></Form.Item>
      {sourceMode === "manual" && <Form.Item name="source_dialog_ids" label="搜索范围" rules={[{ required: true, message: "请选择搜索来源" }]}><Select mode="multiple" showSearch optionFilterProp="label" maxTagCount="responsive" placeholder="选择群组或频道" style={{ width: 260 }} options={options.dialogs.map((item) => ({ value: item.id, label: `${item.name}（${item.dialog_type}）` }))} /></Form.Item>}
      <Form.Item><Button type="primary" htmlType="submit" loading={submitting}>创建任务</Button></Form.Item>
    </Form>
    {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
    <div className={styles.table}><Table<TelegramResourceTask> rowKey="id" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1360 }} /></div>
    <div className={styles.pagination}><Pagination current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[20, 50, 100]} onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }} /><Space><Button onClick={() => void load()}>刷新</Button></Space></div>
  </section></div>;
}
