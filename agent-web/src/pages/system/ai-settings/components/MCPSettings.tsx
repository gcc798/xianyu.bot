import {
  App,
  Button,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { aiConfigApi } from "../../../../api/ai-config";
import { dictionaryName, useDictionaryOptions } from "../../../../hooks/useDictionaryOptions";
import { listAgentTools } from "../../../../api/agent";
import type { MCPServer, MCPServerInput } from "../../../../models/ai-config";
import type { AgentToolCatalogEntry } from "../../../../models/agent";
import styles from "../AISettingsPage.module.css";

const emptyServer: MCPServerInput = {
  name: "",
  code: "",
  baseUrl: "",
  authHeader: "Authorization",
  authToken: "",
  enabled: true,
  requestTimeoutSeconds: 15,
  taskTimeoutSeconds: 300,
};

export function MCPSettings() {
  const { message, modal } = App.useApp();
  const { options: healthStatuses } = useDictionaryOptions("mcp_health_status");
  const [form] = Form.useForm<MCPServerInput>();
  const [rows, setRows] = useState<MCPServer[]>([]);
  const [tools, setTools] = useState<AgentToolCatalogEntry[]>([]);
  const [editing, setEditing] = useState<MCPServer | null | undefined>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const catalog = await listAgentTools();
      const servers = await aiConfigApi.mcpServers();
      setRows(servers.list);
      setTools(catalog.list.filter((item) => item.providerCode));
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "MCP 服务加载失败");
    } finally {
      setLoading(false);
    }
  }, [message]);
  useEffect(() => {
    void load();
  }, [load]);

  const open = (record: MCPServer | null) => {
    setEditing(record);
    form.setFieldsValue(record ? { ...record, authToken: "" } : emptyServer);
  };
  const save = async (input: MCPServerInput) => {
    setSaving(true);
    try {
      if (editing) await aiConfigApi.updateMCPServer(editing.ID, input);
      else await aiConfigApi.createMCPServer(input);
      message.success("MCP 服务已保存");
      setEditing(undefined);
      await load();
    } finally {
      setSaving(false);
    }
  };
  const remove = (record: MCPServer) =>
    modal.confirm({
      title: "删除 MCP 服务？",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await aiConfigApi.deleteMCPServer(record.ID);
        message.success("服务已删除");
        await load();
      },
    });

  const columns: TableColumnsType<MCPServer> = [
    { title: "服务名称", dataIndex: "name", width: 160, ellipsis: true },
    { title: "编码", dataIndex: "code", width: 140, ellipsis: true },
    { title: "服务地址", dataIndex: "baseUrl", ellipsis: true },
    {
      title: "认证",
      dataIndex: "hasAuthToken",
      width: 90,
      render: (value: boolean) => (
        <Tag color={value ? "success" : "default"}>{value ? "已配置" : "无"}</Tag>
      ),
    },
    {
      title: "运行状态",
      dataIndex: "lastStatus",
      width: 110,
      render: (value: string) => (
        <Tag
          color={value === "available" ? "success" : value === "unavailable" ? "error" : "default"}
        >
          {value ? dictionaryName(healthStatuses, value) : "未检查"}
        </Tag>
      ),
    },
    {
      title: "启用",
      dataIndex: "enabled",
      width: 80,
      render: (value: boolean) => (
        <Tag color={value ? "processing" : "default"}>{value ? "启用" : "停用"}</Tag>
      ),
    },
    {
      title: "操作",
      width: 100,
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
          新增服务
        </Button>
      </div>
      <Table<MCPServer>
        rowKey="ID"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={rows}
        pagination={false}
        scroll={{ x: 1000 }}
        expandable={{
          expandedRowRender: (server) => {
            const serverTools = tools.filter((item) => item.providerCode === server.code);
            return serverTools.length ? (
              <Space direction="vertical" size={4}>
                {serverTools.map((item) => (
                  <div key={item.code}>
                    <Typography.Text strong>{item.name}</Typography.Text>
                    <Typography.Text type="secondary">{` · ${item.description}`}</Typography.Text>
                  </div>
                ))}
              </Space>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前没有可用 Tool" />
            );
          },
          rowExpandable: (server) => server.enabled,
        }}
      />
      <Modal
        open={editing !== undefined}
        title={editing ? "编辑 MCP 服务" : "新增 MCP 服务"}
        width={700}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnHidden
        onCancel={() => setEditing(undefined)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={save} className={styles.dialogGrid}>
          <Form.Item name="name" label="服务名称" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="服务编码" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="baseUrl"
            label="服务地址"
            className={styles.span2}
            rules={[
              { required: true, message: "请输入服务地址" },
              {
                validator: async (_, value: string | undefined) => {
                  if (!value) return;
                  try {
                    const url = new URL(value);
                    if (!/^https?:\/\//i.test(value) || !url.hostname) throw new Error();
                  } catch {
                    throw new Error("请输入有效的 HTTP 或 HTTPS 地址");
                  }
                },
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="authHeader" label="认证 Header">
            <Input />
          </Form.Item>
          <Form.Item name="authToken" label="认证 Token">
            <Input.Password
              autoComplete="new-password"
              placeholder={editing?.hasAuthToken ? "留空则保留现有凭据" : "可选"}
            />
          </Form.Item>
          <Form.Item
            name="requestTimeoutSeconds"
            label="请求超时（秒）"
            rules={[{ required: true }]}
          >
            <InputNumber min={1} max={600} />
          </Form.Item>
          <Form.Item name="taskTimeoutSeconds" label="任务超时（秒）" rules={[{ required: true }]}>
            <InputNumber min={1} max={86400} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
