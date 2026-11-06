import { Alert, Button, Modal, Space, Spin, Tag } from "antd";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { telegramApi } from "../../../../api/telegram";
import type { TelegramFileTypeOption, TelegramResourceFile } from "../../../../models/telegram";
import { formatFileSize } from "../../../../utils/format";
import previewStyles from "./ResourcePreviewModal.module.css";

type Props = {
  files: TelegramResourceFile[];
  fileTypes: TelegramFileTypeOption[];
  selectedId: number | null;
  onClose: () => void;
};

export function ResourcePreviewModal({ files, fileTypes, selectedId, onClose }: Props) {
  const typeByValue = useMemo(
    () => new Map(fileTypes.map((item) => [item.value, item])),
    [fileTypes],
  );
  const previewFiles = useMemo(
    () =>
      files.filter(
        (file) => file.upload_status === "success" && typeByValue.get(file.file_type)?.preview,
      ),
    [files, typeByValue],
  );
  const [url, setUrl] = useState("");
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [scale, setScale] = useState(1);
  const pointerStart = useRef<number | null>(null);

  const [activeId, setActiveId] = useState<number | null>(selectedId);
  useEffect(() => setActiveId(selectedId), [selectedId]);
  const activeIndex = previewFiles.findIndex((item) => item.id === (activeId ?? selectedId));
  const activeFile = activeIndex >= 0 ? previewFiles[activeIndex] : null;
  const activeKind = activeFile ? typeByValue.get(activeFile.file_type)?.preview : null;

  const move = useCallback(
    (offset: number) => {
      if (!previewFiles.length) return;
      const current = activeIndex >= 0 ? activeIndex : 0;
      const next = (current + offset + previewFiles.length) % previewFiles.length;
      setActiveId(previewFiles[next]?.id ?? null);
    },
    [activeIndex, previewFiles],
  );

  useEffect(() => {
    if (!activeFile || !activeKind) return;
    let disposed = false;
    let objectUrl = "";
    setLoading(true);
    setError("");
    setUrl("");
    setText("");
    setScale(1);
    void telegramApi
      .resourceFile(activeFile.id, "preview")
      .then(async (blob) => {
        if (activeKind === "text") {
          const content = await blob.text();
          if (!disposed) setText(content);
          return;
        }
        objectUrl = URL.createObjectURL(blob);
        if (disposed) URL.revokeObjectURL(objectUrl);
        else setUrl(objectUrl);
      })
      .catch((cause) => {
        if (!disposed) setError(cause instanceof Error ? cause.message : "文件预览失败");
      })
      .finally(() => {
        if (!disposed) setLoading(false);
      });
    return () => {
      disposed = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [activeFile, activeKind]);

  useEffect(() => {
    if (selectedId === null) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "ArrowLeft") move(-1);
      if (event.key === "ArrowRight") move(1);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [move, selectedId]);

  const renderContent = () => {
    if (loading) return <Spin size="large" />;
    if (error) return <Alert type="error" showIcon title={error} />;
    if (!activeFile || !activeKind)
      return <Alert type="info" showIcon title="该文件不支持在线预览" />;
    if (activeKind === "text") return <pre className={previewStyles.text}>{text}</pre>;
    if (!url) return null;
    if (activeKind === "image")
      return (
        <img
          className={previewStyles.media}
          src={url}
          alt={activeFile.original_filename}
          style={{ transform: `scale(${scale})` }}
        />
      );
    if (activeKind === "video")
      return (
        <video
          className={previewStyles.media}
          src={url}
          controls
          style={{ transform: `scale(${scale})` }}
        />
      );
    if (activeKind === "audio") return <audio className={previewStyles.audio} src={url} controls />;
    return (
      <iframe className={previewStyles.document} src={url} title={activeFile.original_filename} />
    );
  };

  const zoomable = activeKind === "image" || activeKind === "video";
  return (
    <Modal
      className={previewStyles.modal}
      open={selectedId !== null}
      title={activeFile?.original_filename || "文件预览"}
      width="min(1120px, calc(100vw - 48px))"
      footer={null}
      destroyOnHidden
      onCancel={onClose}
    >
      <div className={previewStyles.meta}>
        <Space size={8}>
          <Tag>{typeByValue.get(activeFile?.file_type || "")?.label || activeFile?.file_type}</Tag>
          <span>{formatFileSize(activeFile?.file_size || 0)}</span>
        </Space>
        <Space size={6}>
          {zoomable && (
            <>
              <Button
                aria-label="缩小"
                disabled={scale <= 0.5}
                onClick={() => setScale((value) => Math.max(0.5, value - 0.25))}
              >
                −
              </Button>
              <span className={previewStyles.scale}>{Math.round(scale * 100)}%</span>
              <Button
                aria-label="放大"
                disabled={scale >= 3}
                onClick={() => setScale((value) => Math.min(3, value + 0.25))}
              >
                ＋
              </Button>
              <Button onClick={() => setScale(1)}>还原</Button>
            </>
          )}
          <span>
            {activeIndex + 1} / {previewFiles.length}
          </span>
        </Space>
      </div>
      <div
        className={previewStyles.stage}
        onPointerDown={(event) => {
          pointerStart.current = event.clientX;
        }}
        onPointerUp={(event) => {
          if (pointerStart.current === null) return;
          const distance = event.clientX - pointerStart.current;
          pointerStart.current = null;
          if (Math.abs(distance) > 60) move(distance > 0 ? -1 : 1);
        }}
        onWheel={(event) => {
          if (!zoomable || !event.ctrlKey) return;
          event.preventDefault();
          setScale((value) =>
            Math.min(3, Math.max(0.5, value + (event.deltaY < 0 ? 0.25 : -0.25))),
          );
        }}
      >
        {previewFiles.length > 1 && (
          <Button
            className={`${previewStyles.arrow} ${previewStyles.previous}`}
            aria-label="上一个文件"
            shape="circle"
            onClick={() => move(-1)}
          >
            ‹
          </Button>
        )}
        {renderContent()}
        {previewFiles.length > 1 && (
          <Button
            className={`${previewStyles.arrow} ${previewStyles.next}`}
            aria-label="下一个文件"
            shape="circle"
            onClick={() => move(1)}
          >
            ›
          </Button>
        )}
      </div>
    </Modal>
  );
}
