import { Alert, App, Button, Pagination, Progress, Space, Table, Tabs, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { resourcesApi } from "../../api/resources";
import type { AgentResourceTask } from "../../models/resource";
import styles from "../../styles/page.module.css";
import { formatDateTime } from "../../utils/format";
import { Component as TelegramResourceTasksPage } from "../telegram/resource-tasks/TelegramResourceTasksPage";

const stateMeta: Record<AgentResourceTask["state"], { label: string; color?: string }> = {
  queued: { label: "排队中", color: "processing" },
  running: { label: "执行中", color: "processing" },
  completed: { label: "已完成", color: "success" },
  failed: { label: "失败", color: "error" },
  cancelled: { label: "已取消" },
};

function AgentTasks() {
  const { message, modal } = App.useApp();
  const [rows, setRows] = useState<AgentResourceTask[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await resourcesApi.tasks({ page, pageSize });
      setRows(result.list);
      setTotal(result.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Agent 采集任务加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  useEffect(() => void load(), [load]);

  const cancel = (task: AgentResourceTask) =>
    modal.confirm({
      title: `取消采集任务 #${task.ID}？`,
      content: "外部 MCP 任务将收到取消请求。",
      okText: "取消任务",
      okType: "danger",
      cancelText: "返回",
      onOk: async () => {
        await resourcesApi.cancelTask(task.ID);
        message.success("任务已取消");
        await load();
      },
    });

  const columns: TableColumnsType<AgentResourceTask> = [
    { title: "任务", dataIndex: "ID", width: 80, render: (ID: number) => `#${ID}` },
    { title: "来源", dataIndex: "mcpServerCode", width: 160 },
    {
      title: "阶段",
      dataIndex: "stage",
      width: 150,
      render: (stage: string, task) =>
        task.state === "running" ? (
          <Space size={8}>
            <span>{stage || "执行中"}</span>
            <Progress percent={task.progressPercent} size="small" style={{ width: 100 }} />
          </Space>
        ) : (
          stage || "-"
        ),
    },
    {
      title: "状态",
      dataIndex: "state",
      width: 100,
      render: (state: AgentResourceTask["state"]) => (
        <Tag color={stateMeta[state].color}>{stateMeta[state].label}</Tag>
      ),
    },
    { title: "创建时间", dataIndex: "CreatedAt", width: 160, render: formatDateTime },
    {
      title: "错误",
      dataIndex: "errorMessage",
      ellipsis: true,
      render: (errorMessage: string) => errorMessage || "-",
    },
    {
      title: "操作",
      width: 90,
      fixed: "right",
      render: (_, task) =>
        task.state === "queued" || task.state === "running" ? (
          <Button type="link" danger onClick={() => cancel(task)}>
            取消
          </Button>
        ) : null,
    },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<AgentResourceTask>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 900 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => {
              setPage(nextSize === pageSize ? nextPage : 1);
              setPageSize(nextSize);
            }}
          />
        </div>
      </section>
    </div>
  );
}

export function Component() {
  return (
    <Tabs
      items={[
        { key: "telegram", label: "Telegram 批量扫描", children: <TelegramResourceTasksPage /> },
        { key: "agent", label: "Agent / MCP 获取", children: <AgentTasks /> },
      ]}
    />
  );
}
