import { Alert, App, Button, Input, InputNumber, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { telegramApi } from "../../../api/telegram";
import type { TelegramResourceFile, TelegramResourceOptions } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";
import { formatDateTime, formatFileSize } from "../../../utils/format";

const fileStatusLabel: Record<string, string> = {
  queued: "排队中", downloading: "下载中", success: "成功", failed: "失败", cancelled: "已取消", duplicate: "重复",
  pending: "待上传", uploading: "上传中",
};

const statusColor = (value: string) => value === "success" ? "success" : value === "failed" ? "error" : value === "cancelled" || value === "duplicate" ? undefined : "processing";
const previewable = (file: TelegramResourceFile) => file.file_type === "pdf" || Boolean(file.mime_type?.startsWith("image/") || file.mime_type?.startsWith("text/"));

export function Component() {
  const { message } = App.useApp();
  const [options, setOptions] = useState<TelegramResourceOptions>({ dialogs: [], file_types: [], max_file_size: 0 });
  const [rows, setRows] = useState<TelegramResourceFile[]>([]);
  const [page, setPage] = useState(1), [pageSize, setPageSize] = useState(20), [total, setTotal] = useState(0);
  const [taskId, setTaskId] = useState<number>(), [filename, setFilename] = useState(""), [fileType, setFileType] = useState<string>(), [status, setStatus] = useState<string>(), [sourceId, setSourceId] = useState<number>();
  const [queryFilename, setQueryFilename] = useState("");
  const [loading, setLoading] = useState(false), [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true); setError("");
    try {
      const data = await telegramApi.resourceFiles({ page, page_size: pageSize, task_id: taskId, filename: queryFilename, file_type: fileType, status, source_id: sourceId });
      setRows(data.list); setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "资源文件加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, taskId, queryFilename, fileType, status, sourceId]);

  useEffect(() => { void telegramApi.resourceOptions().then(setOptions).catch((cause) => setError(cause instanceof Error ? cause.message : "资源选项加载失败")); }, []);
  useEffect(() => { void load(); }, [load]);

  const openFile = async (file: TelegramResourceFile, action: "preview" | "download") => {
    try {
      const blob = await telegramApi.resourceFile(file.id, action);
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      if (action === "preview") link.target = "_blank"; else link.download = file.original_filename;
      link.click();
      window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "文件读取失败");
    }
  };

  const columns: TableColumnsType<TelegramResourceFile> = [
    { title: "文件名", dataIndex: "original_filename", width: 280, ellipsis: true },
    { title: "类型", dataIndex: "file_type", width: 90, render: (value: string) => value || "-" },
    { title: "大小", dataIndex: "file_size", width: 110, render: formatFileSize },
    { title: "来源", dataIndex: "source_name", width: 190, ellipsis: true },
    { title: "任务", dataIndex: "task_id", width: 80, render: (value: number) => `#${value}` },
    { title: "完成时间", dataIndex: "completed_at", width: 150, render: (value: string | null) => value ? formatDateTime(value) : "-" },
    { title: "下载", dataIndex: "download_status", width: 100, render: (value: string) => <Tag color={statusColor(value)}>{fileStatusLabel[value] || value}</Tag> },
    { title: "存储", dataIndex: "upload_status", width: 100, render: (value: string) => <Tag color={statusColor(value)}>{fileStatusLabel[value] || value}</Tag> },
    { title: "操作", fixed: "right", width: 120, render: (_, file) => <Space size={2}>
      {previewable(file) && file.upload_status === "success" && <Button type="link" className={styles.linkButton} onClick={() => void openFile(file, "preview")}>预览</Button>}
      {file.upload_status === "success" && <Button type="link" className={styles.linkButton} onClick={() => void openFile(file, "download")}>下载</Button>}
    </Space> },
  ];

  const search = () => { setQueryFilename(filename.trim()); setPage(1); };
  const reset = () => { setTaskId(undefined); setFilename(""); setQueryFilename(""); setFileType(undefined); setStatus(undefined); setSourceId(undefined); setPage(1); };

  return <div className={styles.page}><section className={styles.surface}>
    <div className={styles.filters}>
      <InputNumber value={taskId} min={1} placeholder="任务 ID" style={{ width: 120 }} onChange={(value) => setTaskId(value ?? undefined)} />
      <Input value={filename} allowClear placeholder="文件名" style={{ width: 200 }} onChange={(event) => setFilename(event.target.value)} onPressEnter={search} />
      <Select allowClear value={fileType} placeholder="全部类型" style={{ width: 130 }} options={options.file_types.map((value) => ({ value, label: value }))} onChange={(value) => { setFileType(value); setPage(1); }} />
      <Select allowClear value={status} placeholder="全部状态" style={{ width: 140 }} options={["queued", "downloading", "success", "failed", "cancelled", "duplicate"].map((value) => ({ value, label: fileStatusLabel[value] }))} onChange={(value) => { setStatus(value); setPage(1); }} />
      <Select allowClear showSearch optionFilterProp="label" value={sourceId} placeholder="全部来源" style={{ width: 190 }} options={options.dialogs.map((item) => ({ value: item.id, label: item.name }))} onChange={(value) => { setSourceId(value); setPage(1); }} />
      <Space><Button type="primary" onClick={search}>查询</Button><Button onClick={reset}>重置</Button><Button onClick={() => void load()}>刷新</Button></Space>
    </div>
    {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
    <div className={styles.table}><Table<TelegramResourceFile> rowKey="id" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1220 }} /></div>
    <div className={styles.pagination}><Pagination current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[20, 50, 100]} onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }} /></div>
  </section></div>;
}
