import { MenuFoldOutlined } from "@ant-design/icons";
import type { Session } from "../../../models/agent";

interface SessionSidebarProps {
  sessions: Session[];
  activeSessionId: number;
  onCreate: () => void;
  onOpen: (sessionId: number) => void;
  onDelete: (sessionId: number) => void;
  onCollapse: () => void;
}

/** Bot 历史会话导航；数据加载和删除确认仍由工作区统一处理。 */
export default function SessionSidebar({
  sessions,
  activeSessionId,
  onCreate,
  onOpen,
  onDelete,
  onCollapse,
}: SessionSidebarProps) {
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
            <button className="session-open" onClick={() => onOpen(session.ID)}>
              <strong>{session.title || "新对话"}</strong>
              <small>{formatTime(session.lastMessageAt || session.CreatedAt)}</small>
            </button>
            <button
              className="session-delete"
              aria-label="删除对话"
              onClick={(event) => {
                event.stopPropagation();
                onDelete(session.ID);
              }}
            >
              ×
            </button>
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
