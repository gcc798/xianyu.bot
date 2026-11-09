import {
  Alert,
  App,
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Pagination,
  Radio,
  Select,
  Space,
  Table,
  Tag,
} from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { telegramApi } from "../../../api/telegram";
import type {
  TelegramResourceOptions,
  TelegramResourceTask,
  TelegramResourceTaskInput,
} from "../../../models/telegram";
import businessStyles from "../../../styles/business.module.css";
import styles from "../../../styles/page.module.css";
import { formatDateTime, formatFileSize } from "../../../utils/format";

const statusMeta: Record<TelegramResourceTask["status"], { label: string; color?: string }> = {
  queued: { label: "排队中", color: "processing" },
  running: { label: "执行中", color: "processing" },
  completed: { label: "已完成", color: "success" },
  failed: { label: "失败", color: "error" },
  cancelled: { label: "已取消" },
};

const stageMeta: Record<TelegramResourceTask["current_stage"], { label: string; color?: string }> =
  {
    queued: { label: "等待执行" },
    searching: { label: "检索中", color: "processing" },
    downloading: { label: "下载中", color: "processing" },
    uploading: { label: "上传中", color: "processing" },
    completed: { label: "入库完成", color: "success" },
    no_result: { label: "未找到文件" },
    failed: { label: "处理失败", color: "error" },
    cancelled: { label: "已终止" },
  };

const sourceModeLabel = (task: TelegramResourceTask) => {
  if (task.source_mode === "all") return "全部来源";
  if (task.source_mode === "manual") return "手动选择";
  return task.source_strategy === "ai" ? "智能 · AI" : "智能 · 本地";
};

export function Component() {
  const { message, modal } = App.useApp();
  const navigate = useNavigate();
  const [createForm] = Form.useForm<TelegramResourceTaskInput>();
  const [filterForm] = Form.useForm<{ q?: string; status?: TelegramResourceTask["status"] }>();
  const sourceMode = Form.useWatch("source_mode", createForm);
  const [options, setOptions] = useState<TelegramResourceOptions>({
    dialogs: [],
    file_types: [],
    max_file_size: 0,
  });
  const [rows, setRows] = useState<TelegramResourceTask[]>([]);
  const [filters, setFilters] = useState<{ q?: string; status?: TelegramResourceTask["status"] }>(
    {},
  );
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteTask, setDeleteTask] = useState<TelegramResourceTask | null>(null);
  const [page, setPage] = useState(1),
    [pageSize, setPageSize] = useState(20),
    [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false),
    [submitting, setSubmitting] = useState(false),
    [deleting, setDeleting] = useState(false),
    [error, setError] = useState("");

  const load = useCallback(
    async (quiet = false) => {
      if (!quiet) setLoading(true);
      setError("");
      try {
        const data = await telegramApi.resourceTasks({ page, page_size: pageSize, ...filters });
        setRows(data.list);
        setTotal(data.total);
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : "资源任务加载失败");
      } finally {
        if (!quiet) setLoading(false);
      }
    },
    [filters, page, pageSize],
  );

  useEffect(() => {
    void telegramApi
      .resourceOptions()
      .then(setOptions)
      .catch((cause) => setError(cause instanceof Error ? cause.message : "资源选项加载失败"));
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
      setCreateOpen(false);
      createForm.resetFields();
      if (page === 1) await load();
      else setPage(1);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "任务创建失败");
    } finally {
      setSubmitting(false);
    }
  };

  const runAction = (task: TelegramResourceTask, action: "cancel" | "retry") =>
    modal.confirm({
      title: action === "cancel" ? `取消任务 #${task.id}？` : `重新执行任务 #${task.id}？`,
      content:
        action === "cancel"
          ? "正在下载或上传的文件将停止处理。"
          : "失败或已取消的任务会重新进入队列。",
      okText: action === "cancel" ? "取消任务" : "重新执行",
      okType: action === "cancel" ? "danger" : "primary",
      cancelText: "返回",
      onOk: async () => {
        if (action === "cancel") await telegramApi.cancelResourceTask(task.id);
        else await telegramApi.retryResourceTask(task.id);
        message.success(action === "cancel" ? "任务已取消" : "任务已重新排队");
        await load();
      },
    });

  const removeTask = async (deleteFiles: boolean) => {
    if (!deleteTask) return;
    setDeleting(true);
    try {
      await telegramApi.deleteResourceTask(deleteTask.id, deleteFiles);
      message.success(deleteFiles ? "任务和关联文件已删除" : "任务已删除，关联文件已保留");
      setDeleteTask(null);
      if (rows.length === 1 && page > 1) setPage(page - 1);
      else await load();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "资源任务删除失败");
    } finally {
      setDeleting(false);
    }
  };

  const columns: TableColumnsType<TelegramResourceTask> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    { title: "任务", dataIndex: "id", width: 76, render: (value: number) => `#${value}` },
    { title: "检索主题", dataIndex: "keyword", width: 210, ellipsis: true },
    {
      title: "文件类型",
      dataIndex: "file_types",
      width: 150,
      ellipsis: true,
      render: (value: string[]) => (value.length ? value.join("、") : "不限"),
    },
    {
      title: "范围",
      width: 150,
      render: (_, row) => `${sourceModeLabel(row)} · ${row.source_dialog_ids.length} 个`,
    },
    {
      title: "结果",
      width: 180,
      render: (_, row) =>
        `找到 ${row.files_found} · 入库 ${row.files_completed} / ${row.max_downloads}`,
    },
    {
      title: "当前阶段",
      width: 210,
      render: (_, row) => (
        <Space size={6}>
          <Tag color={stageMeta[row.current_stage].color}>{stageMeta[row.current_stage].label}</Tag>
          {["downloading", "uploading"].includes(row.current_stage) && row.current_file_name ? (
            <span title={row.current_file_name}>
              {row.current_file_name} · {formatFileSize(row.current_file_size || 0)}
            </span>
          ) : null}
        </Space>
      ),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value: TelegramResourceTask["status"]) => (
        <Tag color={statusMeta[value].color}>{statusMeta[value].label}</Tag>
      ),
    },
    {
      title: "创建时间",
      dataIndex: "created_at",
      width: 150,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: "错误",
      dataIndex: "error_message",
      width: 180,
      ellipsis: true,
      render: (value: string | null) => value || "-",
    },
    {
      title: "操作",
      fixed: "right",
      width: 190,
      render: (_, row) => (
        <Space size={2}>
          {row.files_found > 0 && (
            <Button
              type="link"
              className={styles.linkButton}
              onClick={() => navigate(`/telegram/resource-files?task_id=${row.id}`)}
            >
              查看文件
            </Button>
          )}
          {(["queued", "running"] as const).includes(row.status as "queued" | "running") && (
            <Button
              danger
              type="link"
              className={styles.linkButton}
              onClick={() => runAction(row, "cancel")}
            >
              取消
            </Button>
          )}
          {(["failed", "cancelled"] as const).includes(row.status as "failed" | "cancelled") && (
            <Button
              type="link"
              className={styles.linkButton}
              onClick={() => runAction(row, "retry")}
            >
              重试
            </Button>
          )}
          {!(["queued", "running"] as const).includes(row.status as "queued" | "running") && (
            <Button
              danger
              type="link"
              className={styles.linkButton}
              onClick={() => setDeleteTask(row)}
            >
              删除
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <Form
          form={filterForm}
          layout="inline"
          className={styles.filters}
          onFinish={(values) => {
            setPage(1);
            setFilters({ ...values, q: values.q?.trim() || undefined });
          }}
        >
          <Form.Item name="q" label="资源名称">
            <Input allowClear maxLength={200} placeholder="输入资源名称" style={{ width: 220 }} />
          </Form.Item>
          <Form.Item name="status" label="任务状态">
            <Select
              allowClear
              placeholder="全部状态"
              style={{ width: 150 }}
              options={Object.entries(statusMeta).map(([value, meta]) => ({
                value,
                label: meta.label,
              }))}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                查询
              </Button>
              <Button
                onClick={() => {
                  filterForm.resetFields();
                  setPage(1);
                  setFilters({});
                }}
              >
                重置
              </Button>
              <Button onClick={() => setCreateOpen(true)}>创建任务</Button>
            </Space>
          </Form.Item>
        </Form>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<TelegramResourceTask>
            rowKey="id"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1620 }}
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
        <Modal
          title="创建资源任务"
          open={createOpen}
          width={720}
          okText="创建任务"
          cancelText="取消"
          confirmLoading={submitting}
          onOk={() => createForm.submit()}
          onCancel={() => setCreateOpen(false)}
        >
          <Form
            form={createForm}
            layout="vertical"
            className={businessStyles.modalGrid}
            initialValues={{
              file_types: [],
              source_mode: "smart",
              source_dialog_ids: [],
              max_downloads: 10,
            }}
            onFinish={create}
          >
            <Form.Item
              className="ant-form-item-wide"
              name="keyword"
              label="资源名称"
              rules={[{ required: true, whitespace: true, min: 2, max: 200 }]}
            >
              <Input maxLength={200} placeholder="输入资源主题" />
            </Form.Item>
            <Form.Item name="file_types" label="文件类型">
              <Select
                mode="multiple"
                allowClear
                maxTagCount="responsive"
                placeholder="不限"
                options={options.file_types.map((item) => ({
                  value: item.value,
                  label: item.label,
                }))}
              />
            </Form.Item>
            <Form.Item name="max_downloads" label="最大数量" rules={[{ required: true }]}>
              <InputNumber min={1} max={100} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item className="ant-form-item-wide" name="source_mode" label="搜索方式">
              <Radio.Group
                optionType="button"
                buttonStyle="solid"
                options={[
                  { value: "smart", label: "智能" },
                  { value: "all", label: "全部" },
                  { value: "manual", label: "手动" },
                ]}
              />
            </Form.Item>
            {sourceMode === "manual" && (
              <Form.Item
                className="ant-form-item-wide"
                name="source_dialog_ids"
                label="搜索范围"
                rules={[{ required: true, message: "请选择搜索来源" }]}
              >
                <Select
                  mode="multiple"
                  showSearch
                  optionFilterProp="label"
                  maxTagCount="responsive"
                  placeholder="选择群组或频道"
                  options={options.dialogs.map((item) => ({
                    value: item.id,
                    label: `${item.name}（${item.dialog_type}）`,
                  }))}
                />
              </Form.Item>
            )}
          </Form>
        </Modal>
        <Modal
          title={`删除任务 #${deleteTask?.id || ""}？`}
          open={Boolean(deleteTask)}
          closable={!deleting}
          maskClosable={false}
          footer={
            <Space>
              <Button disabled={deleting} onClick={() => setDeleteTask(null)}>
                取消
              </Button>
              <Button loading={deleting} onClick={() => void removeTask(false)}>
                仅删除任务
              </Button>
              <Button
                danger
                type="primary"
                loading={deleting}
                onClick={() => void removeTask(true)}
              >
                删除任务和文件
              </Button>
            </Space>
          }
          onCancel={() => setDeleteTask(null)}
        >
          <p>
            是否同时删除关联资源文件？选择“删除任务和文件”会一并删除文件数据库记录及 RustFS
            文件，且无法恢复。
          </p>
        </Modal>
      </section>
    </div>
  );
}
