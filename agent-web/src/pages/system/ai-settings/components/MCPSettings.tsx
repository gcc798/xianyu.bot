import {
  Alert,
  App,
  Button,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { aiConfigApi } from "../../../../api/ai-config";
import { dictionaryName, useDictionaryOptions } from "../../../../hooks/useDictionaryOptions";
import type { MCPServer, MCPServerInput, MCPTool } from "../../../../models/ai-config";
import { formatDateTime } from "../../../../utils/format";
import styles from "../AISettingsPage.module.css";

const emptyServer: MCPServerInput = {
  name: "",
  code: "",
  description: "",
  tags: [],
  baseUrl: "",
  authHeader: "Authorization",
  authToken: "",
  enabled: true,
  enabledTools: [],
  disabledTools: [],
  requestTimeoutSeconds: 15,
  taskTimeoutSeconds: 300,
  catalogTTLSeconds: 3600,
  connectionTTLSeconds: 600,
  connectionScope: "shared",
  maxConcurrentCalls: 0,
};

function statusColor(value: string) {
  if (value === "available") return "success";
  if (value === "unavailable" || value === "failed") return "error";
  if (value === "refreshing" || value === "checking") return "processing";
  if (value === "stale") return "warning";
  return "default";
}

function catalogStatusName(value: string) {
  if (value === "available") return "可用";
  if (value === "refreshing") return "刷新中";
  if (value === "stale") return "待刷新";
  if (value === "failed") return "失败";
  return "未知";
}

export function MCPSettings() {
  const { message, modal } = App.useApp();
  const { options: healthStatuses } = useDictionaryOptions("mcp_health_status");
  const [form] = Form.useForm<MCPServerInput>();
  const [rows, setRows] = useState<MCPServer[]>([]);
  const [tools, setTools] = useState<Record<number, MCPTool[]>>({});
  const [editing, setEditing] = useState<MCPServer | null | undefined>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [checkingAll, setCheckingAll] = useState(false);
  const [checking, setChecking] = useState<Set<number>>(new Set());
  const [refreshing, setRefreshing] = useState<Set<number>>(new Set());
  const [loadingTools, setLoadingTools] = useState<Set<number>>(new Set());

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const servers = await aiConfigApi.mcpServers();
      setRows(servers.list);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "MCP 服务加载失败");
    } finally {
      setLoading(false);
    }
  }, [message]);
  useEffect(() => {
    void load();
  }, [load]);

  const replaceRow = (server: MCPServer) =>
    setRows((current) => current.map((item) => (item.ID === server.ID ? server : item)));

  const loadTools = async (server: MCPServer) => {
    setLoadingTools((current) => new Set(current).add(server.ID));
    try {
      const result = await aiConfigApi.mcpServerTools(server.ID);
      setTools((current) => ({ ...current, [server.ID]: result.list }));
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "缓存 Tool 加载失败");
    } finally {
      setLoadingTools((current) => {
        const next = new Set(current);
        next.delete(server.ID);
        return next;
      });
    }
  };

  const checkHealth = async (server: MCPServer) => {
    setChecking((current) => new Set(current).add(server.ID));
    try {
      const result = await aiConfigApi.checkMCPServerHealth(server.ID);
      replaceRow(result.server);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "健康检查失败");
    } finally {
      setChecking((current) => {
        const next = new Set(current);
        next.delete(server.ID);
        return next;
      });
    }
  };

  const checkAllHealth = async () => {
    setCheckingAll(true);
    const results = await Promise.allSettled(
      rows.map((server) => aiConfigApi.checkMCPServerHealth(server.ID)),
    );
    const succeeded = results.flatMap((result) =>
      result.status === "fulfilled" ? [result.value.server] : [],
    );
    setRows((current) =>
      current.map((server) => succeeded.find((item) => item.ID === server.ID) ?? server),
    );
    const failed = results.length - succeeded.length;
    if (failed) message.warning(`${succeeded.length} 个检查完成，${failed} 个请求失败`);
    else message.success(`${succeeded.length} 个 MCP 健康检查完成`);
    setCheckingAll(false);
  };

  const refreshCatalog = async (server: MCPServer) => {
    setRefreshing((current) => new Set(current).add(server.ID));
    try {
      const result = await aiConfigApi.refreshMCPServerCatalog(server.ID);
      replaceRow(result.server);
      await loadTools(result.server);
      if (result.server.catalogStatus === "available") message.success("MCP Tool 目录已刷新");
      else message.warning(result.server.catalogError || "目录刷新失败，继续使用已有缓存");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "目录刷新失败");
    } finally {
      setRefreshing((current) => {
        const next = new Set(current);
        next.delete(server.ID);
        return next;
      });
    }
  };

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
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "MCP 服务保存失败");
    } finally {
      setSaving(false);
    }
  };
  const remove = (record: MCPServer) =>
    modal.confirm({
      title: "删除 MCP 服务？",
      content: "配置和缓存目录将一并不可用。",
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
    {
      title: "服务",
      dataIndex: "name",
      width: 190,
      render: (value: string, record) => (
        <Space orientation="vertical" size={0}>
          <Typography.Text strong>{value}</Typography.Text>
          <Typography.Text type="secondary">{record.code}</Typography.Text>
        </Space>
      ),
    },
    { title: "服务地址", dataIndex: "baseUrl", ellipsis: true },
    {
      title: "健康",
      dataIndex: "lastStatus",
      width: 150,
      render: (value: string, record) => (
        <Space size={4} wrap>
          <Tag color={statusColor(value)}>
            {value ? dictionaryName(healthStatuses, value) : "未检查"}
          </Tag>
          {record.lastCheckedAt && (
            <Typography.Text type="secondary">
              {record.lastCheckDurationMillis ?? 0} ms
            </Typography.Text>
          )}
        </Space>
      ),
    },
    {
      title: "Tool 目录",
      dataIndex: "catalogStatus",
      width: 160,
      render: (value: string, record) => (
        <Tooltip
          title={
            record.catalogRefreshedAt
              ? `最近刷新：${formatDateTime(record.catalogRefreshedAt)}`
              : "尚未成功刷新"
          }
        >
          <Space size={4}>
            <Tag color={statusColor(value)}>{catalogStatusName(value)}</Tag>
            <Typography.Text type="secondary">{record.catalogToolCount ?? 0} 个</Typography.Text>
          </Space>
        </Tooltip>
      ),
    },
    {
      title: "启用",
      dataIndex: "enabled",
      width: 70,
      render: (value: boolean) => (
        <Tag color={value ? "processing" : "default"}>{value ? "是" : "否"}</Tag>
      ),
    },
    {
      title: "操作",
      width: 270,
      fixed: "right",
      render: (_, record) => (
        <Space size={0}>
          <Button
            type="link"
            loading={checking.has(record.ID)}
            onClick={() => void checkHealth(record)}
          >
            健康检查
          </Button>
          <Button
            type="link"
            loading={refreshing.has(record.ID)}
            onClick={() => void refreshCatalog(record)}
          >
            刷新目录
          </Button>
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
        <Space>
          <Button
            disabled={!rows.length}
            loading={checkingAll}
            onClick={() => void checkAllHealth()}
          >
            全部健康检查
          </Button>
          <Button type="primary" onClick={() => open(null)}>
            新增服务
          </Button>
        </Space>
      </div>
      <Table<MCPServer>
        rowKey="ID"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={rows}
        pagination={false}
        scroll={{ x: 1120 }}
        expandable={{
          onExpand: (expanded, server) => {
            if (expanded && tools[server.ID] === undefined) void loadTools(server);
          },
          expandedRowRender: (server) => (
            <div className={styles.mcpDetails}>
              <Descriptions
                size="small"
                column={3}
                items={[
                  {
                    key: "healthAt",
                    label: "健康检查时间",
                    children: server.lastCheckedAt
                      ? formatDateTime(server.lastCheckedAt)
                      : "未检查",
                  },
                  {
                    key: "scope",
                    label: "连接作用域",
                    children: server.connectionScope === "owner" ? "按用户隔离" : "共享连接",
                  },
                  {
                    key: "ttl",
                    label: "连接空闲回收",
                    children: server.connectionTTLSeconds
                      ? `${server.connectionTTLSeconds} 秒`
                      : "未配置",
                  },
                  {
                    key: "catalogAt",
                    label: "目录刷新时间",
                    children: server.catalogRefreshedAt
                      ? formatDateTime(server.catalogRefreshedAt)
                      : "未刷新",
                  },
                  {
                    key: "catalogTTL",
                    label: "目录刷新周期",
                    children: server.catalogTTLSeconds
                      ? `${server.catalogTTLSeconds} 秒`
                      : "未配置",
                  },
                  {
                    key: "concurrency",
                    label: "调用并发",
                    children: server.maxConcurrentCalls ? `${server.maxConcurrentCalls}` : "不限制",
                  },
                ]}
              />
              {server.lastError && (
                <Alert type="error" showIcon title={`健康检查：${server.lastError}`} />
              )}
              {server.catalogError && (
                <Alert type="warning" showIcon title={`目录刷新：${server.catalogError}`} />
              )}
              <Spin spinning={loadingTools.has(server.ID)}>
                {(tools[server.ID] ?? []).length ? (
                  <div className={styles.mcpToolList}>
                    {(tools[server.ID] ?? []).map((item) => (
                      <div key={item.name}>
                        <Typography.Text strong>{item.name}</Typography.Text>
                        {item.description && (
                          <Typography.Text type="secondary">{` · ${item.description}`}</Typography.Text>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  !loadingTools.has(server.ID) && (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前没有缓存 Tool" />
                  )
                )}
              </Spin>
            </div>
          ),
        }}
      />
      <Modal
        open={editing !== undefined}
        title={editing ? "编辑 MCP 服务" : "新增 MCP 服务"}
        width={820}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnHidden
        onCancel={() => setEditing(undefined)}
        onOk={() => form.submit()}
      >
        <Form
          form={form}
          layout="vertical"
          autoComplete="off"
          onFinish={save}
          className={styles.dialogGrid}
        >
          <Form.Item name="name" label="服务名称" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="服务编码" rules={[{ required: true, whitespace: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="能力描述" className={styles.span2}>
            <Input.TextArea
              rows={2}
              maxLength={1000}
              showCount
              placeholder="用于 MCP 一级路由判断"
            />
          </Form.Item>
          <Form.Item name="tags" label="路由标签" className={styles.span2}>
            <Select
              mode="tags"
              tokenSeparators={[","]}
              placeholder="输入后回车，例如 email、files"
            />
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
          <Form.Item name="enabledTools" label="Tool 允许列表" className={styles.span2}>
            <Select
              mode="tags"
              tokenSeparators={[","]}
              placeholder="留空允许全部；输入 Tool 名称后回车"
            />
          </Form.Item>
          <Form.Item name="disabledTools" label="Tool 禁止列表" className={styles.span2}>
            <Select
              mode="tags"
              tokenSeparators={[","]}
              placeholder="输入禁止使用的 Tool 名称后回车"
            />
          </Form.Item>
          <Form.Item name="connectionScope" label="连接作用域" rules={[{ required: true }]}>
            <Select
              options={[
                { value: "shared", label: "共享连接" },
                { value: "owner", label: "按用户隔离" },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="maxConcurrentCalls"
            label="最大并发调用（0 不限制）"
            rules={[{ required: true }]}
          >
            <InputNumber min={0} max={1000} />
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
          <Form.Item
            name="catalogTTLSeconds"
            label="目录刷新周期（秒）"
            rules={[{ required: true }]}
          >
            <InputNumber min={10} max={86400} />
          </Form.Item>
          <Form.Item
            name="connectionTTLSeconds"
            label="连接空闲回收（秒）"
            rules={[{ required: true }]}
          >
            <InputNumber min={10} max={86400} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
