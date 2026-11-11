import { Alert, Modal, Spin, Typography } from "antd";
import type { Resource } from "../../../models/resource";
import { formatFileSize } from "../../../utils/format";
import styles from "./ResourcePreviewModal.module.css";

type Props = {
  resource?: Resource;
  url: string;
  loading: boolean;
  error: string;
  onClose: () => void;
};

function previewKind(resource: Resource) {
  const mime = resource.mimeType.toLowerCase();
  const extension = resource.name.split(".").pop()?.toLowerCase() || "";
  if (["text/html", "application/xhtml+xml"].includes(mime)) return "unsupported";
  if (
    mime.startsWith("image/") ||
    ["jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"].includes(extension)
  )
    return "image";
  if (mime.startsWith("video/") || ["mp4", "webm", "mov", "m4v", "ogv"].includes(extension))
    return "video";
  if (mime.startsWith("audio/") || ["mp3", "m4a", "wav", "ogg", "flac", "aac"].includes(extension))
    return "audio";
  if (mime === "application/pdf" || extension === "pdf") return "pdf";
  if (
    [
      "text/plain",
      "text/markdown",
      "text/csv",
      "text/tab-separated-values",
      "application/json",
      "application/xml",
      "text/xml",
    ].includes(mime) ||
    ["txt", "md", "csv", "json", "xml", "log", "yaml", "yml"].includes(extension)
  )
    return "text";
  return "unsupported";
}

export function ResourcePreviewModal({ resource, url, loading, error, onClose }: Props) {
  const kind = resource ? previewKind(resource) : "unsupported";

  return (
    <Modal
      className={styles.modal}
      open={Boolean(resource)}
      title={resource?.name || "资源预览"}
      width="min(1120px, calc(100vw - 48px))"
      footer={null}
      destroyOnHidden
      onCancel={onClose}
    >
      <div className={styles.meta}>
        <Typography.Text type="secondary">{resource?.mimeType || "未知类型"}</Typography.Text>
        <Typography.Text type="secondary">{formatFileSize(resource?.size || 0)}</Typography.Text>
      </div>
      <div className={styles.stage}>
        {loading ? (
          <Spin size="large" />
        ) : error ? (
          <Alert type="error" showIcon title={error} />
        ) : !url ? null : kind === "image" ? (
          <img className={styles.image} src={url} alt={resource?.name} />
        ) : kind === "video" ? (
          <video className={styles.media} src={url} controls />
        ) : kind === "audio" ? (
          <audio className={styles.audio} src={url} controls />
        ) : kind === "pdf" || kind === "text" ? (
          <iframe className={styles.document} src={url} title={resource?.name} />
        ) : (
          <Alert type="info" showIcon title="该文件类型不支持在线预览，请下载查看" />
        )}
      </div>
    </Modal>
  );
}
