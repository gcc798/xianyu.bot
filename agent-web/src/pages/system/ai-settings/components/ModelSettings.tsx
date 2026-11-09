import { App, Button, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { aiConfigApi } from "../../../../api/ai-config";
import type { AIModel, AIModelInput } from "../../../../models/ai-config";
import styles from "../AISettingsPage.module.css";

const emptyModel: AIModelInput = {
  name: "",
  provider: "openai",
  baseUrl: "",
  apiKey: "",
  model: "",
  enabled: true,
  priority: 100,
  supportsTools: true,
  supportsStreaming: true,
  contextWindow: 128000,
  maxOutputTokens: 2000,
  requestTimeoutSeconds: 60,
  extraConfig: "",
};

export function ModelSettings() {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<AIModelInput>();
  const [rows, setRows] = useState<AIModel[]>([]);
  const [editing, setEditing] = useState<AIModel | null | undefined>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setRows((await aiConfigApi.models()).list);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "模型配置加载失败");
    } finally {
      setLoading(false);
    }
  }, [message]);
  useEffect(() => {
    void load();
  }, [load]);

  const open = (record: AIModel | null) => {
    setEditing(record);
    form.setFieldsValue(record ?? emptyModel);
  };
  const save = async (input: AIModelInput) => {
    setSaving(true);
    try {
      if (editing) await aiConfigApi.updateModel(editing.ID, input);
      else await aiConfigApi.createModel(input);
      message.success("模型配置已保存");
      setEditing(undefined);
      await load();
    } finally {
      setSaving(false);
    }
  };
  const remove = (record: AIModel) =>
    modal.confirm({
      title: "删除模型配置？",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await aiConfigApi.deleteModel(record.ID);
        message.success("模型已删除");
        await load();
      },
    });
  const columns: TableColumnsType<AIModel> = [
    { title: "优先级", dataIndex: "priority", width: 80 },
    { title: "名称", dataIndex: "name", width: 160, ellipsis: true },
    { title: "提供商", dataIndex: "provider", width: 110 },
    { title: "模型标识", dataIndex: "model", width: 180, ellipsis: true },
    { title: "服务地址", dataIndex: "baseUrl", ellipsis: true },
    {
      title: "凭据",
      dataIndex: "hasApiKey",
      width: 90,
      render: (value: boolean) => (
        <Tag color={value ? "success" : "warning"}>{value ? "已配置" : "未配置"}</Tag>
      ),
    },
    {
      title: "状态",
      width: 90,
      render: (_, record) => (
        <Tag color={record.enabled ? "processing" : "default"}>
          {record.enabled ? "启用" : "停用"}
        </Tag>
      ),
    },
    {
      title: "操作",
      width: 110,
      fixed: "right",
      render: (_, record) => (
        <Space size={2}>
          <Button type="link" onClick={() => open(record)}>
            编辑
          </Button>
          <Button type="link" danger onClick={() => remove(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div className={styles.toolbar}>
        <Button type="primary" onClick={() => open(null)}>
          新增模型
        </Button>
      </div>
      <Table<AIModel>
        rowKey="ID"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={rows}
        pagination={false}
        scroll={{ x: 1050 }}
      />
      <Modal
        open={editing !== undefined}
        title={editing ? "编辑模型" : "新增模型"}
        width={720}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnHidden
        onCancel={() => setEditing(undefined)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={save} className={styles.dialogGrid}>
          <Form.Item name="name" label="显示名称" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="provider" label="提供商" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
            <InputNumber min={1} precision={0} />
          </Form.Item>
          <Form.Item
            name="baseUrl"
            label="服务地址"
            className={styles.span2}
            rules={[{ required: true, type: "url" }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="model" label="模型标识" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="apiKey" label="API Key" rules={[{ required: true, whitespace: true }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item name="contextWindow" label="上下文窗口" rules={[{ required: true }]}>
            <InputNumber min={1} />
          </Form.Item>
          <Form.Item name="maxOutputTokens" label="最大输出 Token" rules={[{ required: true }]}>
            <InputNumber min={1} />
          </Form.Item>
          <Form.Item
            name="requestTimeoutSeconds"
            label="请求超时（秒）"
            rules={[{ required: true }]}
          >
            <InputNumber min={1} max={600} />
          </Form.Item>
          <Form.Item label="能力开关">
            <div className={styles.switches}>
              <Form.Item name="enabled" valuePropName="checked" noStyle>
                <Switch checkedChildren="启用" unCheckedChildren="停用" />
              </Form.Item>
              <Form.Item name="supportsTools" valuePropName="checked" noStyle>
                <Switch checkedChildren="Tool" unCheckedChildren="Tool" />
              </Form.Item>
              <Form.Item name="supportsStreaming" valuePropName="checked" noStyle>
                <Switch checkedChildren="Stream" unCheckedChildren="Stream" />
              </Form.Item>
            </div>
          </Form.Item>
          <Form.Item name="extraConfig" label="扩展配置 JSON" className={styles.span2}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
