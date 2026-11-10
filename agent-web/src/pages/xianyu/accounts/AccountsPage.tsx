import { Alert, App, Button, Input, Pagination, Space, Switch, Table, Tag, Tooltip } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { accountsApi } from "../../../api/accounts";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type {
  AccountCredentialInput,
  WebSocketStatus,
  XianyuAccount,
} from "../../../models/account";
import pageStyles from "../../../styles/page.module.css";
import businessStyles from "../../../styles/business.module.css";
import AccountEditorModal from "./components/AccountEditorModal";
import type { AccountEditorInput } from "./components/AccountEditorModal";
import CredentialModal from "./components/CredentialModal";
import QRLoginModal from "./components/QRLoginModal";
import { useAccountQRLogin } from "./hooks/useAccountQRLogin";

const accountStatusColor: Record<string, string> = {
  ready: "success",
  expired: "error",
  verification_required: "warning",
  unconfigured: "default",
};

const connectionStatusColor: Record<string, string> = {
  connected: "success",
  connecting: "processing",
  reconnecting: "warning",
  failed: "error",
};

export function Component() {
  const { message, modal } = App.useApp();
  const { options: accountStatuses } = useDictionaryOptions("account_status");
  const { options: connectionStatuses } = useDictionaryOptions("account_connection_status");
  const [rows, setRows] = useState<XianyuAccount[]>([]);
  const [connections, setConnections] = useState<Record<number, WebSocketStatus>>({});
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [editorAccount, setEditorAccount] = useState<XianyuAccount | null | undefined>();
  const [credentialAccount, setCredentialAccount] = useState<XianyuAccount | null>(null);
  const [saving, setSaving] = useState(false);
  const [busyID, setBusyID] = useState(0);

  const refreshConnections = useCallback(async (accounts: XianyuAccount[]) => {
    const results = await Promise.allSettled(
      accounts.map((account) => accountsApi.connectionStatus(account.ID)),
    );
    setConnections((current) => {
      const next = { ...current };
      results.forEach((result, index) => {
        if (result.status === "fulfilled") next[accounts[index].ID] = result.value.connection;
      });
      return next;
    });
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await accountsApi.list({ page, pageSize, keyword: query });
      setRows(response.list);
      setTotal(response.total);
      void refreshConnections(response.list);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "账号加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, query, refreshConnections]);

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    const timer = window.setInterval(() => {
      if (rows.length) void refreshConnections(rows);
    }, 10000);
    return () => window.clearInterval(timer);
  }, [rows, refreshConnections]);
  const qrLogin = useAccountQRLogin({ onConfirmed: load, onError: setError });

  const saveAccount = async (input: AccountEditorInput) => {
    setSaving(true);
    try {
      if (editorAccount) {
        await accountsApi.update({
          ID: editorAccount.ID,
          label: input.label.trim(),
          note: input.note.trim(),
          enabled: editorAccount.enabled,
        });
      } else {
        await accountsApi.create({ label: input.label.trim(), note: input.note.trim() });
      }
      message.success("账号已保存");
      setEditorAccount(undefined);
      await load();
    } finally {
      setSaving(false);
    }
  };

  const saveCredential = async (input: AccountCredentialInput) => {
    setSaving(true);
    try {
      await accountsApi.saveCredential({
        ...input,
        cookie: input.cookie.trim(),
        token: input.token.trim(),
      });
      message.success("账号凭证已保存");
      setCredentialAccount(null);
      await load();
    } finally {
      setSaving(false);
    }
  };

  const updateEnabled = async (record: XianyuAccount, enabled: boolean) => {
    setBusyID(record.ID);
    try {
      await accountsApi.update({ ID: record.ID, label: record.label, note: record.note, enabled });
      setRows((current) =>
        current.map((item) => (item.ID === record.ID ? { ...item, enabled } : item)),
      );
      message.success(enabled ? "账号已启用" : "账号已停用");
    } finally {
      setBusyID(0);
    }
  };

  const checkLogin = async (record: XianyuAccount) => {
    setBusyID(record.ID);
    try {
      const response = await accountsApi.checkLogin(record.ID);
      setRows((current) =>
        current.map((item) =>
          item.ID === record.ID ? { ...item, status: response.login.status } : item,
        ),
      );
      message.info(dictionaryName(accountStatuses, response.login.status));
    } finally {
      setBusyID(0);
    }
  };

  const changeConnection = async (record: XianyuAccount, running: boolean) => {
    setBusyID(record.ID);
    try {
      if (running) await accountsApi.startConnection(record.ID);
      else await accountsApi.stopConnection(record.ID);
      message.success(running ? "消息连接正在启动" : "消息连接已停止");
      const response = await accountsApi.connectionStatus(record.ID);
      setConnections((current) => ({ ...current, [record.ID]: response.connection }));
    } finally {
      setBusyID(0);
    }
  };

  const remove = (record: XianyuAccount) =>
    modal.confirm({
      title: "删除这个账号？",
      content: "账号及其关联配置将无法继续使用。",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await accountsApi.delete(record.ID);
        message.success("账号已删除");
        await load();
      },
    });

  const stopConnection = (record: XianyuAccount) =>
    modal.confirm({
      title: "停止消息连接？",
      content: "停止后不会继续接收该账号的新消息。",
      okText: "停止",
      cancelText: "取消",
      onOk: () => changeConnection(record, false),
    });

  const renderConnection = (record: XianyuAccount) => {
    const connection = connections[record.ID];
    const statusName = connection ? dictionaryName(connectionStatuses, connection.state) : "未连接";
    return (
      <div>
        <Tag color={connectionStatusColor[connection?.state ?? ""] ?? "default"}>{statusName}</Tag>
        {connection?.lastError && (
          <Tooltip title={connection.lastError}>
            <div className={businessStyles.errorText}>{connection.lastError}</div>
          </Tooltip>
        )}
      </div>
    );
  };

  const renderActions = (record: XianyuAccount) => {
    const running = connections[record.ID]?.running;
    return (
      <Space size={2}>
        <Button
          type="link"
          className={pageStyles.linkButton}
          onClick={() => setEditorAccount(record)}
        >
          编辑
        </Button>
        <Button
          type="link"
          className={pageStyles.linkButton}
          onClick={() => setCredentialAccount(record)}
        >
          凭证
        </Button>
        <Button
          type="link"
          className={pageStyles.linkButton}
          onClick={() => void qrLogin.start(record)}
        >
          扫码登录
        </Button>
        <Button
          type="link"
          className={pageStyles.linkButton}
          disabled={!record.hasCredential}
          loading={busyID === record.ID}
          onClick={() => void checkLogin(record)}
        >
          检查登录
        </Button>
        {running ? (
          <Button
            type="link"
            danger
            className={pageStyles.linkButton}
            loading={busyID === record.ID}
            onClick={() => stopConnection(record)}
          >
            停止连接
          </Button>
        ) : (
          <Button
            type="link"
            className={pageStyles.linkButton}
            disabled={record.status !== "ready" || !record.enabled}
            loading={busyID === record.ID}
            onClick={() => void changeConnection(record, true)}
          >
            启动连接
          </Button>
        )}
        <Button type="link" danger className={pageStyles.linkButton} onClick={() => remove(record)}>
          删除
        </Button>
      </Space>
    );
  };

  const columns: TableColumnsType<XianyuAccount> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    { title: "账号名称", dataIndex: "label", width: 125, ellipsis: true },
    {
      title: "闲鱼用户 ID",
      dataIndex: "externalUserId",
      width: 125,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "凭证",
      dataIndex: "hasCredential",
      width: 76,
      render: (value: boolean) => (
        <Tag color={value ? "success" : "default"}>{value ? "已配置" : "未配置"}</Tag>
      ),
    },
    {
      title: "登录态",
      dataIndex: "status",
      width: 76,
      render: (value: string) => (
        <Tag color={accountStatusColor[value] ?? "default"}>
          {dictionaryName(accountStatuses, value) || "未知"}
        </Tag>
      ),
    },
    {
      title: "消息连接",
      key: "connection",
      width: 125,
      render: (_, record) => renderConnection(record),
    },
    {
      title: "启用",
      dataIndex: "enabled",
      width: 62,
      render: (value: boolean, record) => (
        <Switch
          size="small"
          checked={value}
          loading={busyID === record.ID}
          onChange={(checked) => void updateEnabled(record, checked)}
        />
      ),
    },
    {
      title: "备注",
      dataIndex: "note",
      width: 110,
      ellipsis: true,
      render: (value: string) => value || "-",
    },
    {
      title: "操作",
      key: "actions",
      width: 340,
      fixed: "right",
      render: (_, record) => renderActions(record),
    },
  ];

  return (
    <div className={pageStyles.page}>
      <section className={pageStyles.surface}>
        <div className={pageStyles.filters}>
          <Input.Search
            value={keyword}
            allowClear
            placeholder="名称、备注或闲鱼用户 ID"
            style={{ width: 300 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={() => {
              setPage(1);
              setQuery(keyword.trim());
            }}
          />
          <Button
            onClick={() => {
              setKeyword("");
              setQuery("");
              setPage(1);
            }}
          >
            重置
          </Button>
          <Button type="primary" onClick={() => setEditorAccount(null)}>
            新增账号
          </Button>
        </div>
        {error && (
          <Alert
            className={pageStyles.alert}
            type="error"
            showIcon
            title={error}
            closable
            onClose={() => setError("")}
          />
        )}
        <div className={pageStyles.table}>
          <Table<XianyuAccount>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1040 }}
          />
        </div>
        <div className={pageStyles.pagination}>
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
      <AccountEditorModal
        account={editorAccount}
        loading={saving}
        onCancel={() => setEditorAccount(undefined)}
        onSave={saveAccount}
      />
      <CredentialModal
        account={credentialAccount}
        loading={saving}
        onCancel={() => setCredentialAccount(null)}
        onSave={saveCredential}
      />
      <QRLoginModal
        open={qrLogin.open}
        loading={qrLogin.loading}
        login={qrLogin.login}
        onCancel={qrLogin.close}
      />
    </div>
  );
}
