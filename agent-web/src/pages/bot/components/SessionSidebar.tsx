import { CheckOutlined, DeleteOutlined, EditOutlined, MenuFoldOutlined } from "@ant-design/icons";
import { useState } from "react";
import type { Session } from "../../../models/agent";

interface SessionSidebarProps {
  sessions: Session[];
  activeSessionId: number;
  onCreate: () => void;
  onOpen: (sessionId: number) => void;
  onRename: (sessionId: number, title: string) => Promise<void>;
  onDelete: (sessionId: number) => void;
  onCollapse: () => void;
}

/** Bot 历史会话导航；数据加载和删除确认仍由工作区统一处理。 */
export default function SessionSidebar({
  sessions,
  activeSessionId,
  onCreate,
  onOpen,
  onRename,
  onDelete,
  onCollapse,
}: SessionSidebarProps) {
  const [editingId, setEditingId] = useState(0);
  const [title, setTitle] = useState("");
  const [saving, setSaving] = useState(false);

  const saveTitle = async (session: Session) => {
    const nextTitle = title.trim();
    if (!nextTitle || saving) return;
    setSaving(true);
    try {
      if (nextTitle !== session.title) await onRename(session.ID, nextTitle);
      setEditingId(0);
    } finally {
      setSaving(false);
    }
  };

  return (
    <aside className="session-sidebar">
      <div className="session-toolbar">
        <button className="new-chat" onClick={onCreate}>
          <span>＋</span> 新对话
        </button>
        <button
          className="history-toggle"
          title="收起历史对话"
          aria-label="收起历史对话"
          onClick={onCollapse}
        >
          <MenuFoldOutlined />
        </button>
      </div>
      <nav className="session-list" aria-label="对话列表">
        {sessions.map((session) => (
          <div
            key={session.ID}
            className={`session-row ${session.ID === activeSessionId ? "active" : ""}`}
          >
            {editingId === session.ID ? (
              <form
                className="session-edit"
                onSubmit={(event) => {
                  event.preventDefault();
                  void saveTitle(session);
                }}
              >
                <input
                  autoFocus
                  maxLength={200}
                  aria-label="会话名称"
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Escape") setEditingId(0);
                  }}
                />
                <button type="submit" disabled={!title.trim() || saving} aria-label="保存会话名称">
                  <CheckOutlined />
                </button>
              </form>
            ) : (
              <>
                <button className="session-open" onClick={() => onOpen(session.ID)}>
                  <strong title={session.title}>{session.title || "新对话"}</strong>
                  <small>{formatTime(session.lastMessageAt || session.CreatedAt)}</small>
                </button>
                <span className="session-actions">
                  <button
                    aria-label="编辑会话名称"
                    title="编辑名称"
                    onClick={() => {
                      setEditingId(session.ID);
                      setTitle(session.title || "新对话");
                    }}
                  >
                    <EditOutlined />
                  </button>
                  <button
                    className="session-delete"
                    aria-label="删除对话"
                    title="删除对话"
                    onClick={() => onDelete(session.ID)}
                  >
                    <DeleteOutlined />
                  </button>
                </span>
              </>
            )}
          </div>
        ))}
      </nav>
    </aside>
  );
}

function formatTime(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      }).format(new Date(value))
    : "";
}
