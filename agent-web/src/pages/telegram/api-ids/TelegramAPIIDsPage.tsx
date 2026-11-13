import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import { App, Button, Form, Input, InputNumber, Modal, Space, Table, Tag, Typography } from "antd";
import type { TableProps } from "antd";
import { useCallback, useEffect, useState } from "react";
import { telegramAPIConfigApi } from "../../../api/telegram-api";
import type { TelegramAPIConfig, TelegramAPIConfigInput } from "../../../models/telegram-api";
import styles from "../../../styles/page.module.css";

const pageSize = 20;

export function Component() {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<TelegramAPIConfigInput>();
  const [configs, setConfigs] = useState<TelegramAPIConfig[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<TelegramAPIConfig>();
  const [open, setOpen] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await telegramAPIConfigApi.list({ page, pageSize, q: keyword });
      setConfigs(result.list);
      setTotal(result.total);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "Telegram API 配置加载失败");
    } finally {
      setLoading(false);
    }
  }, [keyword, message, page]);

  useEffect(() => {
    void load();
  }, [load]);

  const startCreate = () => {
    setEditing(undefined);
    form.resetFields();
    setOpen(true);
  };

  const startEdit = (config: TelegramAPIConfig) => {
    setEditing(config);
    form.setFieldsValue({ name: config.name, apiId: config.apiId, apiHash: "" });
    setOpen(true);
  };

  const save = async (input: TelegramAPIConfigInput) => {
    setSaving(true);
    try {
      if (editing) await telegramAPIConfigApi.update(editing.ID, input);
      else await telegramAPIConfigApi.create(input);
      message.success(editing ? "Telegram API 配置已更新" : "Telegram API 配置已添加");
      setOpen(false);
      await load();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "Telegram API 配置保存失败");
    } finally {
      setSaving(false);
    }
  };

  const remove = (config: TelegramAPIConfig) => {
    modal.confirm({
      title: `删除 ${config.name}？`,
      content: "若当前 Telegram 服务正在使用此配置，它会在下一个采集任务时切换到其他可用配置。",
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: async () => {
        await telegramAPIConfigApi.delete(config.ID);
        message.success("Telegram API 配置已删除");
        await load();
      },
    });
  };

  const columns: TableProps<TelegramAPIConfig>["columns"] = [
    { title: "名称", dataIndex: "name", width: 180 },
    { title: "API ID", dataIndex: "apiId", width: 150 },
    {
      title: "API Hash",
      dataIndex: "hasApiHash",
      width: 150,
      render: (configured: boolean) =>
        configured ? (
          <Typography.Text type="secondary">•••••••• 已设置</Typography.Text>
        ) : (
          <Tag>未设置</Tag>
        ),
    },
    {
      title: "限流状态",
      dataIndex: "floodWaitUntil",
      width: 220,
      render: (until: string | null) =>
        until && new Date(until).getTime() > Date.now() ? (
          <Space direction="vertical" size={0}>
            <Tag color="warning">限流中</Tag>
            <Typography.Text type="secondary">
              至 {new Date(until).toLocaleString()}
            </Typography.Text>
          </Space>
        ) : (
          <Tag color="success">可用</Tag>
        ),
    },
    {
      title: "最近限流时长",
      dataIndex: "lastFloodWaitSeconds",
      width: 150,
      render: (seconds: number) => (seconds ? `${seconds} 秒` : "-"),
    },
    {
      title: "操作",
      key: "actions",
      width: 180,
      render: (_, config) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => startEdit(config)}>
            编辑
          </Button>
          <Button type="link" danger icon={<DeleteOutlined />} onClick={() => remove(config)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <Space className={styles.sectionToolbar} wrap>
          <Input.Search
            allowClear
            placeholder="搜索名称或 API ID"
            style={{ width: 280 }}
            onSearch={(value) => {
              setPage(1);
              setKeyword(value.trim());
            }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={startCreate}>
            新增 API ID
          </Button>
        </Space>
        <Table<TelegramAPIConfig>
          rowKey="ID"
          columns={columns}
          dataSource={configs}
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showTotal: (count) => `共 ${count} 项`,
            onChange: setPage,
          }}
        />
      </section>
      <Modal
        title={editing ? "编辑 Telegram API ID" : "新增 Telegram API ID"}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" onFinish={save} preserve={false}>
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, whitespace: true, max: 100 }]}
          >
            <Input maxLength={100} />
          </Form.Item>
          <Form.Item
            name="apiId"
            label="API ID"
            rules={[{ required: true, type: "number", min: 1, max: 2147483647 }]}
          >
            <InputNumber precision={0} style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item
            name="apiHash"
            label="API Hash"
            rules={editing ? [] : [{ required: true, len: 32, pattern: /^[a-fA-F0-9]{32}$/ }]}
            extra={
              editing
                ? "留空保留当前 Hash；输入新 Hash 将替换现有凭据。"
                : "32 位十六进制字符串。保存后不会再次明文展示。"
            }
          >
            <Input.Password autoComplete="new-password" maxLength={32} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
