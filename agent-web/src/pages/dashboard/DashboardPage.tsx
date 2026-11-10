import { Alert, Button, DatePicker, Empty, Segmented, Select, Space, Table, Tabs, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useMemo, useState } from "react";
import { accountsApi } from "../../api/accounts";
import { dashboardApi } from "../../api/dashboard";
import { dictionaryName, useDictionaryOptions } from "../../hooks/useDictionaryOptions";
import type { XianyuAccount } from "../../models/account";
import type { BusinessEvent } from "../../models/automation";
import type {
  DashboardMessagePoint,
  DashboardOrderPoint,
  DashboardOverview,
} from "../../models/dashboard";
import pageStyles from "../../styles/page.module.css";
import { formatDateTime } from "../../utils/format";
import styles from "./DashboardPage.module.css";

type DatePreset = "today" | "7d" | "30d" | "custom";
const DAY = 24 * 60 * 60 * 1000;

function presetRange(preset: Exclude<DatePreset, "custom">): [number, number] {
  const end = Date.now();
  const start = new Date(end);
  start.setHours(0, 0, 0, 0);
  if (preset === "7d") start.setTime(start.getTime() - 6 * DAY);
  if (preset === "30d") start.setTime(start.getTime() - 29 * DAY);
  return [start.getTime(), end];
}

const eventStatusColor: Record<string, string> = {
  success: "success",
  failed: "error",
  unknown: "warning",
};

export function Component() {
  const { options: eventStatuses } = useDictionaryOptions("business_event_status");
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [accountId, setAccountId] = useState<number>();
  const [preset, setPreset] = useState<DatePreset>("7d");
  const [range, setRange] = useState<[number, number]>(() => presetRange("7d"));
  const [overview, setOverview] = useState<DashboardOverview>();
  const [loading, setLoading] = useState(false),
    [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setOverview(
        await dashboardApi.overview({ accountId, startTime: range[0], endTime: range[1] }),
      );
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "运营总览加载失败");
    } finally {
      setLoading(false);
    }
  }, [accountId, range]);

  useEffect(() => {
    void accountsApi
      .list({ page: 1, pageSize: 100 })
      .then((data) => setAccounts(data.list))
      .catch((cause) => setError(cause instanceof Error ? cause.message : "账号加载失败"));
  }, []);
  useEffect(() => {
    void load();
  }, [load]);

  const selectPreset = (value: string | number) => {
    const next = value as DatePreset;
    setPreset(next);
    if (next !== "custom") setRange(presetRange(next));
  };

  const metrics = overview?.metrics;
  const metricItems = [
    { label: "可用账号", value: metrics ? `${metrics.readyAccounts} / ${metrics.accounts}` : "-" },
    { label: "在售商品", value: metrics?.onSaleProducts ?? "-" },
    { label: "入站消息", value: metrics?.inboundMessages ?? "-" },
    { label: "自动回复率", value: metrics ? `${metrics.autoReplyRate.toFixed(1)}%` : "-" },
    {
      label: "新增 / 待发货",
      value: metrics ? `${metrics.newOrders} / ${metrics.pendingOrders}` : "-",
    },
    {
      label: "自动化异常",
      value: metrics?.failedAutomationTasks ?? "-",
      alert: Boolean(metrics?.failedAutomationTasks),
    },
  ];

  const pointTime = useCallback(
    (timestamp: number) => {
      const date = new Date(timestamp);
      return overview && overview.bucketMillis < DAY
        ? new Intl.DateTimeFormat("zh-CN", {
            month: "numeric",
            day: "numeric",
            hour: "2-digit",
          }).format(date)
        : new Intl.DateTimeFormat("zh-CN", { month: "numeric", day: "numeric" }).format(date);
    },
    [overview],
  );

  const messageColumns: TableColumnsType<DashboardMessagePoint> = [
    { title: "时间", dataIndex: "timestamp", width: 140, render: pointTime },
    { title: "收到", dataIndex: "inbound", width: 120 },
    { title: "发出", dataIndex: "outbound", width: 120 },
    { title: "自动回复", dataIndex: "autoReplies", width: 120 },
  ];
  const orderColumns: TableColumnsType<DashboardOrderPoint> = [
    { title: "时间", dataIndex: "timestamp", width: 140, render: pointTime },
    { title: "新增订单", dataIndex: "orders", width: 120 },
    { title: "成功发货", dataIndex: "delivered", width: 120 },
    { title: "发货失败", dataIndex: "failed", width: 120 },
  ];
  const eventColumns: TableColumnsType<BusinessEvent> = [
    { title: "时间", dataIndex: "CreatedAt", width: 150, render: formatDateTime },
    {
      title: "账号",
      dataIndex: "accountLabel",
      width: 140,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    { title: "摘要", dataIndex: "summary", width: 280, ellipsis: true },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value: string) => {
        return (
          <Tag color={eventStatusColor[value] ?? "default"}>
            {dictionaryName(eventStatuses, value)}
          </Tag>
        );
      },
    },
    {
      title: "错误",
      dataIndex: "error",
      width: 240,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
  ];

  const automationRows = useMemo(
    () =>
      overview
        ? [
            {
              label: "自动回复",
              success: overview.automation.replySuccess,
              failed: overview.automation.replyFailed,
            },
            {
              label: "自动发货",
              success: overview.automation.deliverySuccess,
              failed: overview.automation.deliveryFailed,
            },
            {
              label: "自动续售",
              success: overview.automation.relistSuccess,
              failed: overview.automation.relistFailed,
            },
          ]
        : [],
    [overview],
  );

  const tabs = [
    {
      key: "messages",
      label: "消息趋势",
      children: (
        <div className={styles.section}>
          <div className={styles.sourceStrip}>
            {overview?.messageSources.map((item) => (
              <span key={item.name}>
                {item.name}
                <b>{item.value}</b>
              </span>
            ))}
          </div>
          <Table<DashboardMessagePoint>
            rowKey="timestamp"
            size="small"
            loading={loading}
            columns={messageColumns}
            dataSource={overview?.messageTrend || []}
            pagination={false}
            locale={{
              emptyText: (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="所选时间内暂无消息" />
              ),
            }}
          />
        </div>
      ),
    },
    {
      key: "orders",
      label: "订单与发货",
      children: (
        <div className={styles.section}>
          <Table<DashboardOrderPoint>
            rowKey="timestamp"
            size="small"
            loading={loading}
            columns={orderColumns}
            dataSource={overview?.orderTrend || []}
            pagination={false}
            locale={{
              emptyText: (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="所选时间内暂无订单" />
              ),
            }}
          />
        </div>
      ),
    },
    {
      key: "automation",
      label: "自动化与事件",
      children: (
        <div className={styles.section}>
          <div className={styles.automation}>
            {automationRows.map((item) => (
              <div className={styles.automationItem} key={item.label}>
                <span>{item.label}</span>
                <div className={styles.automationValues}>
                  <b>成功 {item.success}</b>
                  <em>异常 {item.failed}</em>
                </div>
              </div>
            ))}
          </div>
          <h2 className={styles.sectionTitle}>最近业务事件</h2>
          <Table<BusinessEvent>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={eventColumns}
            dataSource={overview?.recentEvents || []}
            pagination={false}
            scroll={{ x: 910 }}
          />
        </div>
      ),
    },
  ];

  return (
    <div className={pageStyles.page}>
      <section className={pageStyles.surface}>
        <div className={pageStyles.filters}>
          <Select
            allowClear
            value={accountId}
            placeholder="全部账号"
            style={{ width: 170 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={setAccountId}
          />
          <Segmented
            value={preset}
            options={[
              { value: "today", label: "今天" },
              { value: "7d", label: "近 7 天" },
              { value: "30d", label: "近 30 天" },
              { value: "custom", label: "自定义" },
            ]}
            onChange={selectPreset}
          />
          {preset === "custom" && (
            <DatePicker.RangePicker
              showTime
              format="YYYY-MM-DD HH:mm"
              onChange={(value) => {
                if (value?.[0] && value[1]) setRange([value[0].valueOf(), value[1].valueOf()]);
              }}
            />
          )}
          <Space>
            <Button onClick={() => void load()} loading={loading}>
              刷新
            </Button>
          </Space>
        </div>
        {error && <Alert className={pageStyles.alert} type="error" showIcon title={error} />}
        <div className={styles.metrics}>
          {metricItems.map((item) => (
            <div className={styles.metric} data-alert={item.alert} key={item.label}>
              <span>{item.label}</span>
              <strong>{item.value}</strong>
            </div>
          ))}
        </div>
        <div className={styles.content}>
          <Tabs items={tabs} />
        </div>
      </section>
    </div>
  );
}
