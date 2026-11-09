import { App } from "antd";
import { useCallback, useEffect, useRef, useState } from "react";
import { accountsApi } from "../../../../api/accounts";
import type { QRLogin, XianyuAccount } from "../../../../models/account";

interface UseAccountQRLoginOptions {
  onConfirmed: () => void | Promise<void>;
  onError: (message: string) => void;
}

/** 管理扫码登录的轮询、过期请求隔离和关闭清理。 */
export function useAccountQRLogin({ onConfirmed, onError }: UseAccountQRLoginOptions) {
  const { message } = App.useApp();
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [login, setLogin] = useState<QRLogin | null>(null);
  const timer = useRef<number | undefined>(undefined);
  const request = useRef(0);

  const poll = useCallback(
    async (sessionId: string, requestID: number) => {
      try {
        const response = await accountsApi.qrLoginStatus(sessionId);
        // 新一轮扫码开始后，上一轮迟到的响应不能覆盖当前二维码状态。
        if (requestID !== request.current) return;
        setLogin((current) => ({
          ...response.login,
          qrCodeDataUrl: response.login.qrCodeDataUrl || current?.qrCodeDataUrl,
        }));
        if (
          ["confirmed", "expired", "cancelled", "verification_required", "error"].includes(
            response.login.status,
          )
        ) {
          if (response.login.status === "confirmed") {
            message.success("扫码登录成功");
            void onConfirmed();
          }
          return;
        }
        timer.current = window.setTimeout(() => void poll(sessionId, requestID), 1500);
      } catch (cause) {
        if (requestID === request.current) {
          onError(cause instanceof Error ? cause.message : "二维码状态获取失败");
        }
      }
    },
    [message, onConfirmed, onError],
  );

  const start = async (account: XianyuAccount) => {
    const requestID = ++request.current;
    setOpen(true);
    setLoading(true);
    setLogin(null);
    try {
      const response = await accountsApi.startQRLogin(account.ID);
      if (requestID !== request.current) {
        await accountsApi.cancelQRLogin(response.login.sessionId);
        return;
      }
      setLogin(response.login);
      timer.current = window.setTimeout(() => void poll(response.login.sessionId, requestID), 1500);
    } finally {
      if (requestID === request.current) setLoading(false);
    }
  };

  const close = useCallback(() => {
    request.current += 1;
    if (timer.current) window.clearTimeout(timer.current);
    if (login && ["pending", "scanned"].includes(login.status)) {
      void accountsApi.cancelQRLogin(login.sessionId);
    }
    setOpen(false);
    setLogin(null);
  }, [login]);

  useEffect(
    () => () => {
      request.current += 1;
      if (timer.current) window.clearTimeout(timer.current);
    },
    [],
  );

  return { open, loading, login, start, close };
}
