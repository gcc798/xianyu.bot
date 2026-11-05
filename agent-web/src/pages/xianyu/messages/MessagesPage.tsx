import { Alert, App, Button, DatePicker, Form, Input, Modal, Pagination, Select, Space, Table, Tag } from "antd";
import type { GetProps, TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { accountsApi } from "../../../api/accounts";
import { messagesApi } from "../../../api/messages";
import type { XianyuAccount } from "../../../models/account";
import type { XianyuMessage } from "../../../models/message";
import { formatDateTime } from "../../../utils/format";
import businessStyles from "../../../styles/business.module.css";
import pageStyles from "../../../styles/page.module.css";

type RangeValue = GetProps<typeof DatePicker.RangePicker>["value"];

export function Component() {
  const { message, modal } = App.useApp();
  const [searchParams] = useSearchParams();
  const initialAccountID = Number(searchParams.get("accountId")) || undefined;
  const initialStart = Number(searchParams.get("startTime")) || undefined;
  const initialEnd = Number(searchParams.get("endTime")) || undefined;
  const [rows, setRows] = useState<XianyuMessage[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number | undefined>(initialAccountID);
  const [direction, setDirection] = useState<string>();
  const [externalProductId, setExternalProductId] = useState("");
  const [buyerName, setBuyerName] = useState("");
  const [content, setContent] = useState("");
  const [timeRange, setTimeRange] = useState<RangeValue>();
  const [applied, setApplied] = useState({ externalProductId: "", buyerName: "", content: "", startTime: initialStart, endTime: initialEnd });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [replyTarget, setReplyTarget] = useState<XianyuMessage | null>(null);
  const [sending, setSending] = useState(false);
  const [replyForm] = Form.useForm<{ text: string }>();

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await messagesApi.list({ page, pageSize, accountId, direction, ...applied });
      setRows(response.list);
      setTotal(response.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "消息加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, direction, applied]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => { void accountsApi.list({ page: 1, pageSize: 100 }).then((response) => setAccounts(response.list)).catch((cause) => setError(cause instanceof Error ? cause.message : "账号加载失败")); }, []);

  const applyFilters = () => {
    setApplied({
      externalProductId: externalProductId.trim(),
      buyerName: buyerName.trim(),
      content: content.trim(),
      startTime: timeRange?.[0]?.valueOf(),
      endTime: timeRange?.[1]?.valueOf(),
    });
    setPage(1);
  };

  const reset = () => {
    setAccountId(undefined); setDirection(undefined); setExternalProductId(""); setBuyerName(""); setContent(""); setTimeRange(undefined);
    setApplied({ externalProductId: "", buyerName: "", content: "", startTime: undefined, endTime: undefined }); setPage(1);
  };

  const sendReply = async ({ text }: { text: string }) => {
    if (!replyTarget) return;
    setSending(true);
    try {
      await messagesApi.reply(replyTarget.ID, text.trim());
      message.success("消息已发送，会话已转为人工接管");
      setReplyTarget(null);
      replyForm.resetFields();
      await load();
    } finally {
      setSending(false);
    }
  };

  const toggleTakeover = (record: XianyuMessage) => {
    const active = !record.takeoverActive;
    modal.confirm({
      title: active ? "确认人工接管？" : "确认释放接管？",
      content: active ? "接管后该会话将暂停自动回复。" : "释放后该会话可以再次触发自动回复。",
      okText: active ? "接管" : "释放",
      cancelText: "取消",
      onOk: async () => { await messagesApi.updateTakeover(record.ID, active); message.success(active ? "已人工接管" : "已释放接管"); await load(); },
    });
  };

  const renderActions = (record: XianyuMessage) => (
    <Space size={2}>
      <Button
        type="link"
        className={pageStyles.linkButton}
        disabled={record.direction !== "inbound" || record.contentType === 25}
        onClick={() => { setReplyTarget(record); replyForm.setFieldsValue({ text: "" }); }}
      >回复</Button>
      <Button type="link" className={pageStyles.linkButton} onClick={() => toggleTakeover(record)}>
        {record.takeoverActive ? "释放" : "接管"}
      </Button>
    </Space>
  );

  const columns: TableColumnsType<XianyuMessage> = [
    {
      title: "时间", dataIndex: "messageTime", width: 130,
      render: (value: number, record) => formatDateTime(value || record.CreatedAt),
    },
    { title: "账号", dataIndex: "accountLabel", width: 95, ellipsis: true },
    {
      title: "发送方", dataIndex: "senderUserName", width: 105, ellipsis: true,
      render: (value: string, record) => value || record.senderUserId || "-",
    },
    {
      title: "方向", dataIndex: "direction", width: 75,
      render: (value: string) => <Tag color={value === "inbound" ? "processing" : "default"}>{value === "inbound" ? "买家发来" : "卖家发出"}</Tag>,
    },
    {
      title: "消息内容", dataIndex: "content", width: 280,
      render: (value: string, record) => <div className={businessStyles.messageContent}>{record.contentType === 25 ? "平台消息" : value || "-"}</div>,
    },
    { title: "商品编码", dataIndex: "externalProductId", width: 110, ellipsis: true, render: (value: string) => value || "-" },
    { title: "订单号", dataIndex: "externalOrderId", width: 120, ellipsis: true, render: (value: string) => value || "-" },
    {
      title: "接管状态", dataIndex: "takeoverActive", width: 85,
      render: (value: boolean) => <Tag color={value ? "warning" : "default"}>{value ? "人工接管" : "自动处理"}</Tag>,
    },
    { title: "操作", key: "actions", width: 110, fixed: "right", render: (_, record) => renderActions(record) },
  ];

  return (
    <div className={pageStyles.page}>
      <section className={pageStyles.surface}>
        <div className={pageStyles.filters} style={{ flexWrap: "wrap" }}>
          <Select
            allowClear value={accountId} placeholder="全部账号" style={{ width: 160 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => { setAccountId(value); setPage(1); }}
          />
          <DatePicker.RangePicker showTime value={timeRange} onChange={setTimeRange} />
          <Select
            allowClear value={direction} placeholder="全部方向" style={{ width: 130 }}
            options={[{ value: "inbound", label: "买家发来" }, { value: "outbound", label: "卖家发出" }]}
            onChange={(value) => { setDirection(value); setPage(1); }}
          />
          <Input value={externalProductId} allowClear placeholder="商品编码" style={{ width: 160 }} onChange={(event) => setExternalProductId(event.target.value)} />
          <Input value={buyerName} allowClear placeholder="买家名称" style={{ width: 140 }} onChange={(event) => setBuyerName(event.target.value)} />
          <Input value={content} allowClear placeholder="消息内容" style={{ width: 180 }} onChange={(event) => setContent(event.target.value)} />
          <Space><Button type="primary" onClick={applyFilters}>查询</Button><Button onClick={reset}>重置</Button></Space>
        </div>
        {error && <Alert className={pageStyles.alert} type="error" showIcon title={error} closable onClose={() => setError("")} />}
        <div className={pageStyles.table}>
          <Table<XianyuMessage> rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 1110 }} />
        </div>
        <div className={pageStyles.pagination}>
          <Pagination
            current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }}
          />
        </div>
      </section>
      <Modal
        open={Boolean(replyTarget)}
        title="人工回复"
        okText="发送"
        cancelText="取消"
        confirmLoading={sending}
        destroyOnHidden
        onCancel={() => setReplyTarget(null)}
        onOk={() => replyForm.submit()}
      >
        <Alert type="warning" showIcon title="发送成功后，该会话将进入人工接管，自动回复暂停。" style={{ marginBottom: 16 }} />
        <Form form={replyForm} layout="vertical" onFinish={sendReply}>
          <Form.Item
            name="text"
            label="回复内容"
            rules={[{ required: true, whitespace: true, message: "请输入回复内容" }, { max: 4000, message: "最多 4000 个字符" }]}
          >
            <Input.TextArea rows={5} maxLength={4000} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
