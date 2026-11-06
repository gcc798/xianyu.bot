import { Alert, App, Button, Descriptions, Input, Modal, Progress, Space, Tag } from "antd";
import { useCallback, useEffect, useRef, useState } from "react";
import { telegramApi } from "../../../api/telegram";
import type { TelegramQR, TelegramScan, TelegramStatus } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";

const emptyScan: TelegramScan = { status: "idle", current: 0, total: 0, name: "", message: "" };

export function Component() {
  const { message } = App.useApp();
  const [status, setStatus] = useState<TelegramStatus>();
  const [qr, setQR] = useState<TelegramQR>();
  const [scan, setScan] = useState<TelegramScan>(emptyScan);
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const timer = useRef<number | undefined>(undefined);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setStatus(await telegramApi.status());
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Telegram 状态加载失败");
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load();
    return () => window.clearInterval(timer.current);
  }, [load]);

  const pollLogin = () => {
    window.clearInterval(timer.current);
    timer.current = window.setInterval(async () => {
      try {
        const next = await telegramApi.loginStatus();
        setQR(next);
        if (["success", "expired", "error", "2fa"].includes(next.status)) {
          window.clearInterval(timer.current);
          if (next.status === "success") {
            setQR(undefined);
            await load();
          }
        }
      } catch (cause) {
        window.clearInterval(timer.current);
        setError(cause instanceof Error ? cause.message : "登录状态查询失败");
      }
    }, 1500);
  };
  const startLogin = async () => {
    try {
      setQR(await telegramApi.startQR());
      pollLogin();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "二维码登录启动失败");
    }
  };
  const submitPassword = async () => {
    if (!password) return;
    try {
      setQR(await telegramApi.submitPassword(password));
      setPassword("");
      pollLogin();
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "两步验证失败");
    }
  };
  const startScan = async () => {
    try {
      const result = await telegramApi.startScan();
      setScan(result.progress);
      window.clearInterval(timer.current);
      timer.current = window.setInterval(async () => {
        try {
          const next = await telegramApi.scanStatus();
          setScan(next);
          if (["completed", "failed"].includes(next.status)) {
            window.clearInterval(timer.current);
            if (next.status === "completed") await load();
          }
        } catch (cause) {
          window.clearInterval(timer.current);
          setError(cause instanceof Error ? cause.message : "扫描状态查询失败");
        }
      }, 1500);
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "会话扫描启动失败");
    }
  };

  const summary = status?.summary;
  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.sectionToolbar}>
          <Space>
            <Tag color={status?.authorized ? "success" : "default"}>
              {status?.authorized ? "已登录" : "未登录"}
            </Tag>
            <span>{status?.account || "未绑定账号"}</span>
          </Space>
          <Space>
            {!status?.authorized && (
              <Button type="primary" loading={loading} onClick={() => void startLogin()}>
                扫码登录
              </Button>
            )}
            <Button
              type="primary"
              disabled={!status?.authorized || scan.status === "running"}
              onClick={() => void startScan()}
            >
              扫描会话
            </Button>
            <Button onClick={() => void load()}>刷新</Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <Descriptions
          bordered
          column={4}
          size="small"
          style={{ margin: "0 20px" }}
          items={[
            { key: "total", label: "会话总数", children: summary?.total ?? "-" },
            { key: "groups", label: "群组", children: summary?.groups_count ?? "-" },
            { key: "channels", label: "频道", children: summary?.channels_count ?? "-" },
            { key: "unread", label: "未读消息", children: summary?.unread ?? "-" },
            { key: "abnormal", label: "异常会话", children: summary?.abnormal ?? "-" },
            { key: "muted", label: "已静音", children: summary?.muted ?? "-" },
            { key: "archived", label: "已归档", children: summary?.archived ?? "-" },
            { key: "created", label: "我创建的", children: summary?.created_by_me ?? "-" },
          ]}
        />
        {scan.status !== "idle" && (
          <div style={{ padding: 20 }}>
            <Progress
              percent={scan.total ? Math.round((scan.current / scan.total) * 100) : 0}
              status={
                scan.status === "failed"
                  ? "exception"
                  : scan.status === "completed"
                    ? "success"
                    : "active"
              }
            />
            <span>
              {scan.message}
              {scan.name ? ` · ${scan.name}` : ""}
            </span>
          </div>
        )}
      </section>
      <Modal
        open={Boolean(qr)}
        title="Telegram 扫码登录"
        footer={null}
        onCancel={() => {
          window.clearInterval(timer.current);
          setQR(undefined);
        }}
      >
        {qr?.image && (
          <img
            src={qr.image}
            alt="Telegram 登录二维码"
            style={{ display: "block", width: 280, margin: "0 auto" }}
          />
        )}
        {qr?.status === "connecting" && <p>正在连接 Telegram…</p>}
        {qr?.status === "2fa" && (
          <Space.Compact block>
            <Input.Password
              value={password}
              placeholder="两步验证密码"
              onChange={(event) => setPassword(event.target.value)}
            />
            <Button type="primary" onClick={() => void submitPassword()}>
              提交
            </Button>
          </Space.Compact>
        )}
        {qr?.error && <Alert type="error" showIcon title={qr.error} style={{ marginTop: 16 }} />}
      </Modal>
    </div>
  );
}
