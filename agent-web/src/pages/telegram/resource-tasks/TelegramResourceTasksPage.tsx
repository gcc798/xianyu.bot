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
  Tooltip,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { telegramResourceApi } from "../../../api/telegram";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type {
  TelegramResourceTask,
  TelegramResourceTaskInput,
  TelegramResourceTaskOptions,
} from "../../../models/telegram";
import businessStyles from "../../../styles/business.module.css";
import styles from "../../../styles/page.module.css";
import { formatDateTime } from "../../../utils/format";

const statusColor: Record<string, string> = {
  queued: "processing",
  running: "processing",
  completed: "success",
  failed: "error",
};

function sourceModeLabel(
  task: TelegramResourceTask,
  modes: { code: string; name: string }[],
  strategies: { code: string; name: string }[],
) {
  if (task.sourceMode === "smart") {
    return `${dictionaryName(modes, task.sourceMode)} · ${dictionaryName(strategies, task.sourceStrategy)}`;
  }
  return dictionaryName(modes, task.sourceMode);
}

export function Component() {
  const { message, modal } = App.useApp();
  const { options: taskStatuses } = useDictionaryOptions("telegram_resource_task_status");
  const { options: sourceModes } = useDictionaryOptions("telegram_resource_source_mode");
  const { options: sourceStrategies } = useDictionaryOptions("telegram_resource_source_strategy");
  const [createForm] = Form.useForm<TelegramResourceTaskInput>();
  const sourceMode = Form.useWatch("sourceMode", createForm);
  const [options, setOptions] = useState<TelegramResourceTaskOptions>({
    dialogs: [],
    fileTypes: [],
    maxFileSize: 0,
  });
  const [rows, setRows] = useState<TelegramResourceTask[]>([]);
  const [status, setStatus] = useState<TelegramResourceTask["status"] | undefined>();
  const [createOpen, setCreateOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [syncingSources, setSyncingSources] = useState(false);
  const [error, setError] = useState("");
  const [selectedError, setSelectedError] = useState<{ taskId: number; message: string } | null>(
    null,
  );

  const load = useCallback(
    async (quiet = false) => {
      if (!quiet) setLoading(true);
      setError("");
      try {
        const result = await telegramResourceApi.tasks({ page, pageSize, state: status });
        setRows(result.list);
        setTotal(result.total);
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : "资源任务加载失败");
      } finally {
        if (!quiet) setLoading(false);
      }
    },
    [page, pageSize, status],
  );

  const loadOptions = useCallback(async () => {
    const next = await telegramResourceApi.options();
    if (!next.dialogs.length) {
      const synced = await telegramResourceApi.syncSources();
      next.dialogs = synced.dialogs;
    }
    setOptions(next);
  }, []);
  useEffect(() => {
    void loadOptions().catch((cause) =>
      setError(cause instanceof Error ? cause.message : "资源选项加载失败"),
    );
  }, [loadOptions]);
  useEffect(() => {
    void load();
    const timer = window.setInterval(() => void load(true), 5000);
    return () => window.clearInterval(timer);
  }, [load]);

  const create = async (values: TelegramResourceTaskInput) => {
    setSubmitting(true);
    try {
      await telegramResourceApi.createTask({ ...values, keyword: values.keyword.trim() });
      message.success("任务已进入队列");
      setCreateOpen(false);
      createForm.resetFields();
      setPage(1);
      if (page === 1) await load();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "任务创建失败");
    } finally {
      setSubmitting(false);
    }
  };

  const cancel = (task: TelegramResourceTask) =>
    modal.confirm({
      title: `取消任务 #${task.id}？`,
      content: "正在处理的 Telegram 请求将停止，已登记的资源不会删除。",
      okText: "取消任务",
      okType: "danger",
      cancelText: "返回",
      onOk: async () => {
        await telegramResourceApi.cancelTask(task.id);
        message.success("任务已取消");
        await load();
      },
    });

  const syncSources = async () => {
    setSyncingSources(true);
    setError("");
    try {
      const synced = await telegramResourceApi.syncSources();
      setOptions((current) => ({ ...current, dialogs: synced.dialogs }));
      message.success(`已同步 ${synced.dialogs.length} 个 Telegram 来源`);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Telegram 来源同步失败");
    } finally {
      setSyncingSources(false);
    }
  };

  const columns: TableColumnsType<TelegramResourceTask> = [
    { title: "任务", dataIndex: "id", width: 76, render: (id: number) => `#${id}` },
    { title: "检索主题", dataIndex: "keyword", width: 220, ellipsis: true },
    {
      title: "文件类型",
      dataIndex: "fileTypes",
      width: 150,
      render: (values: string[]) => (values.length ? values.join("、") : "不限"),
    },
    {
      title: "范围",
      width: 180,
      render: (_, task) =>
        `${sourceModeLabel(task, sourceModes, sourceStrategies)} · ${task.sourceDialogIds.length} 个`,
    },
    {
      title: "结果",
      width: 180,
      render: (_, task) =>
        `找到 ${task.filesFound} · 入库 ${task.filesCompleted} / ${task.maxDownloads}`,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value: TelegramResourceTask["status"]) => (
        <Tag color={statusColor[value] ?? "default"}>{dictionaryName(taskStatuses, value)}</Tag>
      ),
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 160,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: "错误",
      dataIndex: "errorMessage",
      ellipsis: true,
      render: (value: string | null, task) =>
        value ? (
          <Tooltip
            title={
              <span
                style={{
                  display: "block",
                  maxWidth: 420,
                  whiteSpace: "pre-wrap",
                  overflowWrap: "anywhere",
                }}
              >
                {value}
              </span>
            }
            placement="topLeft"
          >
            <Button
              type="link"
              className={styles.linkButton}
              aria-label={`查看任务 ${task.id} 错误详情`}
              onClick={() => setSelectedError({ taskId: task.id, message: value })}
            >
              <span
                style={{
                  display: "block",
                  maxWidth: 190,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  textAlign: "left",
                }}
              >
                {value}
              </span>
            </Button>
          </Tooltip>
        ) : (
          "-"
        ),
    },
    {
      title: "操作",
      width: 90,
      fixed: "right",
      render: (_, task) =>
        task.status === "queued" || task.status === "running" ? (
          <Button danger type="link" className={styles.linkButton} onClick={() => cancel(task)}>
            取消
          </Button>
        ) : null,
    },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <Space className={styles.filters} wrap>
          <Select
            allowClear
            aria-label="任务状态"
            placeholder="全部状态"
            style={{ width: 150 }}
            value={status}
            options={taskStatuses.map(({ code, name }) => ({ value: code, label: name }))}
            onChange={(value) => {
              setPage(1);
              setStatus(value);
            }}
          />
          <Button loading={syncingSources} onClick={() => void syncSources()}>
            同步来源
          </Button>
          <Button onClick={() => setCreateOpen(true)}>创建任务</Button>
        </Space>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<TelegramResourceTask>
            rowKey="id"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1280 }}
            locale={{ emptyText: "暂无资源采集任务" }}
          />
        </div>
        <Modal
          title={selectedError ? `任务 #${selectedError.taskId} 错误详情` : "错误详情"}
          open={Boolean(selectedError)}
          footer={null}
          onCancel={() => setSelectedError(null)}
        >
          <Typography.Paragraph
            style={{ marginBottom: 0, whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}
          >
            {selectedError?.message}
          </Typography.Paragraph>
        </Modal>
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
          <Form<TelegramResourceTaskInput>
            form={createForm}
            layout="vertical"
            className={businessStyles.modalGrid}
            initialValues={{
              fileTypes: [],
              sourceMode: "smart",
              sourceDialogIds: [],
              maxDownloads: 10,
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
            <Form.Item name="fileTypes" label="文件类型">
              <Select
                mode="multiple"
                allowClear
                maxTagCount="responsive"
                placeholder="不限"
                options={options.fileTypes.map((value) => ({ value, label: value.toUpperCase() }))}
              />
            </Form.Item>
            <Form.Item name="maxDownloads" label="最大数量" rules={[{ required: true }]}>
              <InputNumber min={1} max={100} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item className="ant-form-item-wide" name="sourceMode" label="搜索方式">
              <Radio.Group
                optionType="button"
                buttonStyle="solid"
                options={sourceModes.map(({ code, name }) => ({ value: code, label: name }))}
              />
            </Form.Item>
            {sourceMode === "manual" && (
              <Form.Item
                className="ant-form-item-wide"
                name="sourceDialogIds"
                label="搜索范围"
                rules={[{ required: true, message: "请选择搜索来源" }]}
              >
                <Select
                  mode="multiple"
                  showSearch
                  optionFilterProp="label"
                  maxTagCount="responsive"
                  placeholder="选择群组或频道"
                  options={options.dialogs.map((dialog) => ({
                    value: dialog.id,
                    label: `${dialog.name}（${dialog.dialogType}）`,
                  }))}
                />
              </Form.Item>
            )}
          </Form>
        </Modal>
      </section>
    </div>
  );
}
