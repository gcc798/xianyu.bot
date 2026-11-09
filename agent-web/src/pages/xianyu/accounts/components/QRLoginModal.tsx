import { Alert, Modal, Result, Spin } from "antd";
import type { QRLogin } from "../../../../models/account";
import styles from "../../../../styles/business.module.css";

type Props = {
  open: boolean;
  loading: boolean;
  login: QRLogin | null;
  onCancel: () => void;
};

export default function QRLoginModal({ open, loading, login, onCancel }: Props) {
  return (
    <Modal
      open={open}
      title="闲鱼扫码登录"
      width={460}
      footer={null}
      mask={{ closable: false }}
      onCancel={onCancel}
    >
      <div className={styles.qrBody}>
        {loading ? (
          <Spin size="large" />
        ) : login?.qrCodeDataUrl ? (
          <div>
            <img src={login.qrCodeDataUrl} alt="闲鱼登录二维码" className={styles.qrCode} />
            <div className={styles.qrStatus}>{login.message || "请使用闲鱼 App 扫码"}</div>
          </div>
        ) : (
          <Result
            status={
              login?.status === "confirmed"
                ? "success"
                : login?.status === "verification_required"
                  ? "warning"
                  : "info"
            }
            title={login?.message || "二维码会话已结束"}
          />
        )}
        {login?.status === "verification_required" && (
          <Alert type="warning" showIcon title="请先在闲鱼 App 完成安全验证，再重新扫码登录。" />
        )}
      </div>
    </Modal>
  );
}
