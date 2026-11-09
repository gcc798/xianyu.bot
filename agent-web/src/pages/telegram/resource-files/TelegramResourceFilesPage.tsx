import {
  Alert,
  App,
  Button,
  Input,
  InputNumber,
  Pagination,
  Select,
  Space,
  Table,
  Tag,
} from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { telegramApi } from "../../../api/telegram";
import type { TelegramResourceFile, TelegramResourceOptions } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";
import { formatDateTime, formatFileSize } from "../../../utils/format";
import { ResourcePreviewModal } from "./components/ResourcePreviewModal";

const fileStatusLabel: Record<string, string> = {
  queued: "排队中",
  downloading: "下载中",
  success: "成功",
  failed: "失败",
  cancelled: "已取消",
  duplicate: "重复",
  pending: "待上传",
  uploading: "上传中",
};

const statusColor = (value: string) =>
  value === "success"
    ? "success"
    : value === "failed"
      ? "error"
      : value === "cancelled" || value === "duplicate"
        ? undefined
        : "processing";
export function Component() {
  const { message, modal } = App.useApp();
  const [searchParams] = useSearchParams();
  const initialTaskId = Number(searchParams.get("task_id"));
  const [options, setOptions] = useState<TelegramResourceOptions>({
    dialogs: [],
    file_types: [],
    max_file_size: 0,
  });
  const [rows, setRows] = useState<TelegramResourceFile[]>([]);
  const [page, setPage] = useState(1),
    [pageSize, setPageSize] = useState(20),
    [total, setTotal] = useState(0);
  const [taskId, setTaskId] = useState<number | undefined>(
      initialTaskId > 0 ? initialTaskId : undefined,
    ),
    [filename, setFilename] = useState(""),
    [fileType, setFileType] = useState<string>(),
    [status, setStatus] = useState<string>(),
    [sourceId, setSourceId] = useState<number>();
  const [queryFilename, setQueryFilename] = useState("");
  const [loading, setLoading] = useState(false),
    [error, setError] = useState("");
  const [previewId, setPreviewId] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await telegramApi.resourceFiles({
        page,
        page_size: pageSize,
        task_id: taskId,
        filename: queryFilename,
        file_type: fileType,
        status,
        source_id: sourceId,
      });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "资源文件加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, taskId, queryFilename, fileType, status, sourceId]);

  useEffect(() => {
    void telegramApi
      .resourceOptions()
      .then(setOptions)
      .catch((cause) => setError(cause instanceof Error ? cause.message : "资源选项加载失败"));
  }, []);
  useEffect(() => {
    void load();
  }, [load]);

  const downloadFile = async (file: TelegramResourceFile) => {
    try {
      const blob = await telegramApi.resourceFile(file.id, "download");
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = file.original_filename;
      link.click();
      window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "文件读取失败");
    }
  };

  const deleteFile = (file: TelegramResourceFile) =>
    modal.confirm({
      title: `删除文件“${file.original_filename}”？`,
      content: "文件数据库记录及 RustFS 文件将同步删除，且无法恢复。",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await telegramApi.deleteResourceFile(file.id);
        message.success("资源文件已删除");
        if (rows.length === 1 && page > 1) setPage(page - 1);
        else await load();
      },
    });

  const columns: TableColumnsType<TelegramResourceFile> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    {
      title: "任务",
      dataIndex: "task_id",
      width: 80,
      render: (value: number | null) => (value ? `#${value}` : "-"),
    },
    {
      title: "文件名",
      dataIndex: "original_filename",
      width: 280,
      ellipsis: true,
      render: (value: string, file) =>
        options.file_types.find((item) => item.value === file.file_type)?.preview &&
        file.upload_status === "success" ? (
          <Button
            type="link"
            className={styles.linkButton}
            title={value}
            onClick={() => setPreviewId(file.id)}
          >
            {value}
          </Button>
        ) : (
          value
        ),
    },
    {
      title: "类型",
      dataIndex: "file_type",
      width: 120,
      render: (value: string) =>
        options.file_types.find((item) => item.value === value)?.label || value || "-",
    },
    { title: "大小", dataIndex: "file_size", width: 110, render: formatFileSize },
    { title: "来源", dataIndex: "source_name", width: 190, ellipsis: true },
    {
      title: "完成时间",
      dataIndex: "completed_at",
      width: 150,
      render: (value: string | null) => (value ? formatDateTime(value) : "-"),
    },
    {
      title: "下载",
      dataIndex: "download_status",
      width: 100,
      render: (value: string) => (
        <Tag color={statusColor(value)}>{fileStatusLabel[value] || value}</Tag>
      ),
    },
    {
      title: "存储",
      dataIndex: "upload_status",
      width: 100,
      render: (value: string) => (
        <Tag color={statusColor(value)}>{fileStatusLabel[value] || value}</Tag>
      ),
    },
    {
      title: "操作",
      fixed: "right",
      width: 120,
      render: (_, file) => (
        <Space size={2}>
          {file.upload_status === "success" && (
            <Button
              type="link"
              className={styles.linkButton}
              onClick={() => void downloadFile(file)}
            >
              下载
            </Button>
          )}
          {!(["queued", "downloading"] as const).includes(
            file.download_status as "queued" | "downloading",
          ) &&
            !(["pending", "uploading"] as const).includes(
              file.upload_status as "pending" | "uploading",
            ) && (
              <Button
                danger
                type="link"
                className={styles.linkButton}
                onClick={() => deleteFile(file)}
              >
                删除
              </Button>
            )}
        </Space>
      ),
    },
  ];

  const search = () => {
    setQueryFilename(filename.trim());
    setPage(1);
  };
  const reset = () => {
    setTaskId(undefined);
    setFilename("");
    setQueryFilename("");
    setFileType(undefined);
    setStatus(undefined);
    setSourceId(undefined);
    setPage(1);
  };

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <InputNumber
            value={taskId}
            min={1}
            placeholder="任务 ID"
            style={{ width: 120 }}
            onChange={(value) => setTaskId(value ?? undefined)}
          />
          <Input
            value={filename}
            allowClear
            placeholder="文件名"
            style={{ width: 200 }}
            onChange={(event) => setFilename(event.target.value)}
            onPressEnter={search}
          />
          <Select
            allowClear
            value={fileType}
            placeholder="全部类型"
            style={{ width: 160 }}
            options={options.file_types.map((item) => ({ value: item.value, label: item.label }))}
            onChange={(value) => {
              setFileType(value);
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={status}
            placeholder="全部状态"
            style={{ width: 140 }}
            options={["queued", "downloading", "success", "failed", "cancelled", "duplicate"].map(
              (value) => ({ value, label: fileStatusLabel[value] }),
            )}
            onChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
          />
          <Select
            allowClear
            showSearch
            optionFilterProp="label"
            value={sourceId}
            placeholder="全部来源"
            style={{ width: 190 }}
            options={options.dialogs.map((item) => ({ value: item.id, label: item.name }))}
            onChange={(value) => {
              setSourceId(value);
              setPage(1);
            }}
          />
          <Space>
            <Button type="primary" onClick={search}>
              查询
            </Button>
            <Button onClick={reset}>重置</Button>
            <Button onClick={() => void load()}>刷新</Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<TelegramResourceFile>
            rowKey="id"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1260 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            pageSizeOptions={[20, 50, 100]}
            onChange={(nextPage, nextSize) => {
              setPage(nextSize === pageSize ? nextPage : 1);
              setPageSize(nextSize);
            }}
          />
        </div>
        <ResourcePreviewModal
          files={rows}
          fileTypes={options.file_types}
          selectedId={previewId}
          onClose={() => setPreviewId(null)}
        />
      </section>
    </div>
  );
}
