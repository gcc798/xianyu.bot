import { BorderOutlined, MenuUnfoldOutlined } from "@ant-design/icons";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  approveRun,
  cancelRun,
  createSession,
  deleteSession,
  getSession,
  getTrajectory,
  listAgentTools,
  listSessions,
  rejectRun,
  sendMessage,
  streamEvents,
  updateSession,
} from "../../api/agent";
import type {
  AgentEvent,
  AgentToolCatalogEntry,
  Approval,
  ChatMessage,
  Directive,
  Operation,
  Session,
} from "../../models/agent";
import TrajectoryPanel from "./components/TrajectoryPanel";
import SessionSidebar from "./components/SessionSidebar";

const skills = [
  {
    type: "skill" as const,
    code: "resource-discovery",
    name: "resource-discovery",
    description: "理解目标并编排资源检索",
    source: "Skill",
  },
  {
    type: "skill" as const,
    code: "telegram-resource",
    name: "telegram-resource",
    description: "使用当前 MCP 的资源能力",
    source: "Skill",
  },
];
const terminalEvents = new Set(["run.completed", "run.failed", "run.cancelled"]);
type ApprovalFeedback = { status: "completed" | "rejected" | "failed"; message: string };

export default function BotWorkspace() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionId, setSessionId] = useState(0);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [operations, setOperations] = useState<Operation[]>([]);
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [approvalBusyRunId, setApprovalBusyRunId] = useState<number | null>(null);
  const [approvalFeedback, setApprovalFeedback] = useState<ApprovalFeedback | null>(null);
  const [draft, setDraft] = useState("");
  const [directives, setDirectives] = useState<Directive[]>([]);
  const [availableTools, setAvailableTools] = useState<AgentToolCatalogEntry[]>([]);
  const [running, setRunning] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const [runId, setRunId] = useState(0);
  const [lastEventId, setLastEventId] = useState(0);
  const [error, setError] = useState("");
  const [slashOpen, setSlashOpen] = useState(false);
  const [slashIndex, setSlashIndex] = useState(0);
  const [operationsOpen, setOperationsOpen] = useState(false);
  const [trajectoryOpen, setTrajectoryOpen] = useState(false);
  const [historyCollapsed, setHistoryCollapsed] = useState(false);
  const streamRef = useRef<AbortController | null>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  const composingRef = useRef(false);

  const activeSession = sessions.find((item) => item.ID === sessionId);
  const slashQuery =
    draft
      .match(/(?:^|\s)\/([^\n/]*)$/)?.[1]
      ?.trim()
      .toLowerCase() || "";
  const [slashType, ...slashWords] = slashQuery.split(/\s+/);
  const slashTerm =
    slashType === "skill" || slashType === "tool" ? slashWords.join(" ") : slashQuery;
  const capabilities = useMemo(() => [...skills, ...availableTools], [availableTools]);
  const filteredCapabilities = capabilities.filter(
    (item) =>
      (!["skill", "tool"].includes(slashType) || item.type === slashType) &&
      (!slashTerm ||
        `${item.name} ${item.description} ${item.source}`.toLowerCase().includes(slashTerm)),
  );

  useEffect(() => {
    let alive = true;
    listAgentTools()
      .then(({ list }) => {
        if (alive) setAvailableTools(list || []);
      })
      .catch(() => {
        if (alive) setAvailableTools([]);
      });
    return () => {
      alive = false;
    };
  }, []);

  useEffect(() => {
    if (!filteredCapabilities.length) return;
    document
      .querySelector(`[data-slash-index="${slashIndex % filteredCapabilities.length}"]`)
      ?.scrollIntoView({ block: "nearest" });
  }, [slashIndex, filteredCapabilities.length]);

  const refreshSessions = useCallback(async () => {
    const data = await listSessions();
    setSessions(data.list || []);
    return data.list || [];
  }, []);

  const openSession = useCallback(async (id: number, preserveFeedback = false) => {
    // 切换会话前终止旧事件流，避免旧会话增量写入新会话界面。
    streamRef.current?.abort();
    setSessionId(id);
    setError("");
    if (!preserveFeedback) setApprovalFeedback(null);
    setRunning(false);
    setCancelling(false);
    setRunId(0);
    const [{ session }, trajectory] = await Promise.all([getSession(id), getTrajectory(id)]);
    setMessages(session.messages || []);
    setOperations(
      (session.toolTraces || []).map((item) => ({
        callId: item.callId || `stored-${item.id}`,
        runId: item.runId,
        tool: item.toolName,
        status: item.status,
        durationMillis: item.durationMillis,
      })),
    );
    setApprovals(session.pendingApprovals || []);
    setLastEventId(trajectory.latestEventId || 0);
  }, []);

  useEffect(() => {
    let alive = true;
    refreshSessions()
      .then((items) => {
        if (alive && items.length) void openSession(items[0].ID);
      })
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : "加载失败"));
    return () => {
      alive = false;
      streamRef.current?.abort();
    };
  }, [openSession, refreshSessions]);

  useEffect(() => {
    viewportRef.current?.scrollTo({ top: viewportRef.current.scrollHeight, behavior: "smooth" });
  }, [messages, operations, approvals]);

  const reloadActive = useCallback(async () => {
    if (sessionId) await Promise.all([openSession(sessionId, true), refreshSessions()]);
  }, [openSession, refreshSessions, sessionId]);

  const applyEvent = useCallback(
    (event: AgentEvent) => {
      if (event.id) setLastEventId((current) => Math.max(current, event.id));
      const data = event.data;
      if (event.type === "assistant.delta") {
        setMessages((items) =>
          items.map((item) =>
            item.runId === Number(data.runId) && item.role === "assistant"
              ? { ...item, content: item.content + String(data.content || "") }
              : item,
          ),
        );
      } else if (event.type === "tool.started") {
        setOperations((items) => [
          ...items,
          {
            callId: String(data.callId || `event-${event.id}`),
            runId: Number(data.runId),
            tool: String(data.tool),
            server: String(data.server || ""),
            status: "running",
          },
        ]);
      } else if (event.type === "tool.completed" || event.type === "tool.failed") {
        setOperations((items) =>
          items.map((item) =>
            item.callId === String(data.callId)
              ? {
                  ...item,
                  status:
                    event.type === "tool.failed" ? "failed" : String(data.status || "completed"),
                  durationMillis: Number(data.durationMillis || 0),
                  summary: String((data.error as { message?: string } | undefined)?.message || ""),
                }
              : item,
          ),
        );
      } else if (event.type === "approval.required") {
        setApprovals((items) => [
          ...items,
          {
            runId: Number(data.runId),
            toolCode: String(data.toolCode || data.tool),
            toolName: String(data.toolName || data.tool),
            toolSource: String(data.server ? "mcp" : "local"),
            createdAt: new Date().toISOString(),
          },
        ]);
      }
      if (terminalEvents.has(event.type)) {
        const failureMessage =
          event.type === "run.failed" ? String(data.message || "Agent 运行失败") : "";
        setRunning(false);
        setCancelling(false);
        setMessages((items) =>
          items.map((item) =>
            item.runId === Number(data.runId) ? { ...item, streaming: false } : item,
          ),
        );
        streamRef.current?.abort();
        // 终态后以服务端持久化结果为准，补齐流断开前可能尚未到达的事件。
        void reloadActive().finally(() => {
          if (failureMessage) setError(failureMessage);
        });
      }
    },
    [reloadActive],
  );

  const connectStream = useCallback(
    (id: number, cursor: number) => {
      streamRef.current?.abort();
      const controller = new AbortController();
      streamRef.current = controller;
      void streamEvents(id, cursor, controller.signal, applyEvent).catch((cause: unknown) => {
        if (!controller.signal.aborted)
          setError(cause instanceof Error ? cause.message : "事件流已断开");
      });
    },
    [applyEvent],
  );

  const newConversation = async () => {
    try {
      const { session } = await createSession();
      await refreshSessions();
      await openSession(session.ID);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "创建失败");
    }
  };

  const removeConversation = async (id: number) => {
    if (!window.confirm("删除这条对话及其审计轨迹？")) return;
    try {
      await deleteSession(id);
      const items = await refreshSessions();
      if (id === sessionId) {
        if (items.length) await openSession(items[0].ID);
        else {
          setSessionId(0);
          setMessages([]);
          setOperations([]);
        }
      }
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "删除失败");
    }
  };

  const renameConversation = async (id: number, title: string) => {
    try {
      await updateSession(id, title);
      setSessions((items) =>
        items.map((session) => (session.ID === id ? { ...session, title } : session)),
      );
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "修改会话名称失败");
      throw cause;
    }
  };

  const submit = async () => {
    const content = draft.trim();
    if (!content || running) return;
    try {
      let targetSession = sessionId;
      if (!targetSession) {
        const created = await createSession();
        targetSession = created.session.ID;
        await refreshSessions();
        setSessionId(targetSession);
      }
      setDraft("");
      setSlashOpen(false);
      setError("");
      setCancelling(false);
      const response = await sendMessage(targetSession, content, directives);
      const nextRunId = response.run.id;
      setRunId(nextRunId);
      setRunning(true);
      setMessages((items) => [
        ...items,
        { localId: `user-${nextRunId}`, runId: nextRunId, role: "user", content },
        {
          localId: `assistant-${nextRunId}`,
          runId: nextRunId,
          role: "assistant",
          content: "",
          streaming: true,
        },
      ]);
      setDirectives([]);
      connectStream(targetSession, lastEventId);
      void refreshSessions().catch((cause: unknown) =>
        setError(cause instanceof Error ? cause.message : "刷新会话名称失败"),
      );
    } catch (cause) {
      setRunning(false);
      setError(cause instanceof Error ? cause.message : "发送失败");
    }
  };

  const stopRun = async () => {
    if (!sessionId || !runId || cancelling) return;
    setCancelling(true);
    setError("");
    try {
      await cancelRun(sessionId, runId);
    } catch (cause) {
      setCancelling(false);
      setError(cause instanceof Error ? cause.message : "停止运行失败");
    }
  };

  const selectCapability = (type: "skill" | "tool", name: string) => {
    setDirectives((items) => [...items.filter((item) => item.type !== type), { type, name }]);
    setDraft((value) => value.replace(/(?:^|\s)\/[^\n/]*$/, "").trimStart());
    setSlashOpen(false);
  };
  const approve = async (approval: Approval) => {
    setApprovalBusyRunId(approval.runId);
    setError("");
    try {
      const result = await approveRun(approval.runId);
      setApprovalFeedback({
        status: result.status === "completed" ? "completed" : "failed",
        message:
          result.status === "completed"
            ? `${approval.toolName} 已执行成功。`
            : `${approval.toolName} 未成功执行，当前状态：${result.status}。`,
      });
      if (result.status === "completed") {
        setApprovals((items) => items.filter((item) => item !== approval));
      }
      try {
        await reloadActive();
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : "会话刷新失败");
        setApprovalFeedback((current) =>
          current
            ? { ...current, message: `${current.message} 会话刷新失败，请稍后重新打开会话核对。` }
            : current,
        );
      }
    } catch (cause) {
      const failure = cause instanceof Error ? cause.message : "确认失败";
      try {
        await reloadActive();
      } catch {
        setError("审批状态刷新失败，请重新打开会话核对。");
      }
      setApprovalFeedback({
        status: "failed",
        message: `${approval.toolName} 执行未成功：${failure}`,
      });
    } finally {
      setApprovalBusyRunId(null);
    }
  };
  const reject = async (approval: Approval) => {
    setApprovalBusyRunId(approval.runId);
    setError("");
    try {
      const result = await rejectRun(approval.runId);
      setApprovalFeedback({
        status: result.status === "rejected" ? "rejected" : "failed",
        message:
          result.status === "rejected"
            ? `${approval.toolName} 已拒绝，未执行。`
            : `${approval.toolName} 未能完成拒绝处理，当前状态：${result.status}。`,
      });
      if (result.status === "rejected") {
        setApprovals((items) => items.filter((item) => item !== approval));
      }
      try {
        await reloadActive();
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : "会话刷新失败");
        setApprovalFeedback((current) =>
          current
            ? { ...current, message: `${current.message} 会话刷新失败，请稍后重新打开会话核对。` }
            : current,
        );
      }
    } catch (cause) {
      const failure = cause instanceof Error ? cause.message : "拒绝失败";
      try {
        await reloadActive();
      } catch {
        setError("审批状态刷新失败，请重新打开会话核对。");
      }
      setApprovalFeedback({
        status: "failed",
        message: `${approval.toolName} 未能完成拒绝：${failure}`,
      });
    } finally {
      setApprovalBusyRunId(null);
    }
  };

  const groupedOperations = useMemo(
    () => operations.filter((item) => !runId || item.runId === runId),
    [operations, runId],
  );
  return (
    <main className={`agent-app ${historyCollapsed ? "history-collapsed" : ""}`}>
      <SessionSidebar
        sessions={sessions}
        activeSessionId={sessionId}
        onCreate={() => void newConversation()}
        onOpen={(id) => void openSession(id)}
        onRename={renameConversation}
        onDelete={(id) => void removeConversation(id)}
        onCollapse={() => setHistoryCollapsed(true)}
      />

      <section className="chat-stage">
        <header className="chat-header">
          <div className="chat-title">
            {historyCollapsed && (
              <button
                className="history-toggle"
                title="展开历史对话"
                aria-label="展开历史对话"
                onClick={() => setHistoryCollapsed(false)}
              >
                <MenuUnfoldOutlined />
              </button>
            )}
            <div>
              <strong title={activeSession?.title}>{activeSession?.title || "新对话"}</strong>
              {running && <small>正在执行</small>}
            </div>
          </div>
          <div className="header-actions">
            <button
              className="trace-button"
              disabled={!sessionId}
              onClick={() => setTrajectoryOpen(true)}
            >
              轨迹 <span>⌘</span>
            </button>
          </div>
        </header>
        <div className="message-scroll" ref={viewportRef} aria-live="polite">
          {!messages.length && (
            <section className="welcome">
              <h1>今天想完成什么？</h1>
              <div className="prompt-grid">
                {[
                  "帮我检索斗破苍穹相关资源，并展示完整检索过程",
                  "列出当前可访问的来源，优先找名称包含“搜”的群组",
                  "检查当前 MCP 能力，只说明不可用项",
                ].map((text) => (
                  <button key={text} onClick={() => setDraft(text)}>
                    {text}
                    <span>↗</span>
                  </button>
                ))}
              </div>
            </section>
          )}
          <div className="message-column">
            {messages.map((message) => (
              <article key={message.ID || message.localId} className={`message ${message.role}`}>
                <div>
                  {message.role === "assistant" && message.content ? (
                    <div className="markdown-content">
                      <Markdown
                        remarkPlugins={[remarkGfm]}
                        skipHtml
                        components={{
                          a: ({ href, children }) => (
                            <a href={href} target="_blank" rel="noreferrer">
                              {children}
                            </a>
                          ),
                          table: ({ children }) => (
                            <div className="markdown-table">
                              <table>{children}</table>
                            </div>
                          ),
                        }}
                      >
                        {message.content}
                      </Markdown>
                    </div>
                  ) : (
                    message.content || (message.streaming ? " " : "")
                  )}
                  {message.streaming && <span className="typing-cursor" />}
                </div>
              </article>
            ))}
            {!!groupedOperations.length && (
              <section className="operation-group">
                <button
                  className="operation-summary"
                  onClick={() => setOperationsOpen((value) => !value)}
                >
                  <span className="operation-icon">⌘</span>
                  <span>执行了 {groupedOperations.length} 个操作</span>
                  <small>
                    {groupedOperations.some((item) => item.status === "failed")
                      ? "包含失败"
                      : running
                        ? "执行中"
                        : "已完成"}
                  </small>
                  <b>{operationsOpen ? "⌃" : "⌄"}</b>
                </button>
                {operationsOpen && (
                  <div className="operation-list">
                    {groupedOperations.map((operation) => (
                      <div className="operation-row" key={operation.callId}>
                        <span className={`op-state ${operation.status}`} />
                        <div>
                          <strong>{operation.tool}</strong>
                          <small>
                            {operation.server || "MCP"} · {operation.callId}
                          </small>
                        </div>
                        <em>
                          {operation.durationMillis
                            ? `${operation.durationMillis} ms`
                            : operation.status}
                        </em>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            )}
            {approvalFeedback && (
              <section
                className={`approval-feedback ${approvalFeedback.status}`}
                role={approvalFeedback.status === "failed" ? "alert" : "status"}
                aria-live="polite"
              >
                <strong>
                  {approvalFeedback.status === "completed"
                    ? "操作已执行"
                    : approvalFeedback.status === "rejected"
                      ? "操作已拒绝"
                      : "审批处理失败"}
                </strong>
                <p>{approvalFeedback.message}</p>
              </section>
            )}
            {approvals.slice(0, 1).map((approval) => (
              <section
                className="approval"
                key={`${approval.runId}-${approval.toolName}-${approval.createdAt}`}
              >
                <div>
                  <strong>{approval.toolName} 需要确认</strong>
                  <p>
                    该操作尚未执行，确认后会继续。
                    {approvals.length > 1 && ` 另有 ${approvals.length - 1} 项待审批。`}
                  </p>
                </div>
                <div className="approval-actions">
                  <button
                    className="approval-reject"
                    disabled={approvalBusyRunId === approval.runId}
                    onClick={() => void reject(approval)}
                  >
                    拒绝
                  </button>
                  <button
                    disabled={approvalBusyRunId === approval.runId}
                    onClick={() => void approve(approval)}
                  >
                    {approvalBusyRunId === approval.runId ? "处理中…" : "确认执行"}
                  </button>
                </div>
              </section>
            ))}
          </div>
        </div>

        <footer className="composer-wrap">
          {error && (
            <div className="error-banner">
              <span>!</span>
              {error}
              <button onClick={() => setError("")}>×</button>
            </div>
          )}
          {!!directives.length && (
            <div className="directive-list">
              {directives.map((directive) => (
                <span key={directive.type} className={directive.type}>
                  <b>{directive.type}</b>
                  {directive.type === "tool"
                    ? `${availableTools.find((item) => item.code === directive.name)?.name || directive.name} · ${availableTools.find((item) => item.code === directive.name)?.source || "Tool"}`
                    : directive.name}
                  <button
                    onClick={() =>
                      setDirectives((items) => items.filter((item) => item.type !== directive.type))
                    }
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          )}
          {slashOpen && (
            <div className="slash-picker" role="listbox" aria-label="Skill 和 Tool">
              {filteredCapabilities.map((capability, index) => (
                <button
                  key={`${capability.type}-${capability.code}`}
                  className={`${capability.type}${slashIndex % filteredCapabilities.length === index ? " selected" : ""}`}
                  data-slash-index={index}
                  onMouseDown={(event) => {
                    event.preventDefault();
                    selectCapability(
                      capability.type,
                      capability.type === "tool" ? capability.code : capability.name,
                    );
                  }}
                  onMouseEnter={() => setSlashIndex(index)}
                  role="option"
                  aria-selected={slashIndex % filteredCapabilities.length === index}
                >
                  <div>
                    <strong>{capability.name}</strong>
                    <small>{capability.description}</small>
                  </div>
                  <em>{capability.type === "skill" ? "Skill" : capability.source}</em>
                </button>
              ))}
              {!filteredCapabilities.length && (
                <p className="slash-empty">没有匹配的 Skill 或 Tool</p>
              )}
            </div>
          )}
          <div className="composer">
            <textarea
              value={draft}
              rows={2}
              disabled={running}
              placeholder="输入运营问题或任务，输入 / 选择 Skill 或 Tool"
              onChange={(event) => {
                setDraft(event.target.value);
                setSlashIndex(0);
                setSlashOpen(/(?:^|\s)\/[^\n/]*$/.test(event.target.value));
              }}
              onCompositionStart={() => {
                composingRef.current = true;
              }}
              onCompositionEnd={() => {
                composingRef.current = false;
              }}
              onKeyDown={(event) => {
                if (event.nativeEvent.isComposing || composingRef.current) return;
                if (slashOpen && event.key === "ArrowDown" && filteredCapabilities.length) {
                  event.preventDefault();
                  setSlashIndex((index) => {
                    return (index + 1) % filteredCapabilities.length;
                  });
                } else if (slashOpen && event.key === "ArrowUp" && filteredCapabilities.length) {
                  event.preventDefault();
                  setSlashIndex((index) => {
                    return (index - 1 + filteredCapabilities.length) % filteredCapabilities.length;
                  });
                } else if (slashOpen && event.key === "Enter" && filteredCapabilities.length) {
                  event.preventDefault();
                  const capability = filteredCapabilities[slashIndex % filteredCapabilities.length];
                  selectCapability(
                    capability.type,
                    capability.type === "tool" ? capability.code : capability.name,
                  );
                } else if (event.key === "Escape") setSlashOpen(false);
                if (event.key === "Enter" && !event.shiftKey && !slashOpen) {
                  event.preventDefault();
                  void submit();
                }
              }}
            />
            <div className="composer-meta">
              <small>Enter 发送 · Shift + Enter 换行</small>
              <button
                className={`send-button${running ? " stop-run-button" : ""}`}
                disabled={running ? cancelling : !draft.trim()}
                onClick={() => void (running ? stopRun() : submit())}
                aria-label={running ? (cancelling ? "正在停止" : "停止生成") : "发送"}
                title={running ? (cancelling ? "正在停止" : "停止生成") : "发送"}
              >
                {running ? <BorderOutlined /> : "↑"}
              </button>
            </div>
          </div>
          <p className="disclaimer">
            AI 生成内容可能有误，重要操作请核实。完整调用记录保存在审计轨迹中。
          </p>
        </footer>
      </section>
      {trajectoryOpen && sessionId > 0 && (
        <TrajectoryPanel sessionId={sessionId} onClose={() => setTrajectoryOpen(false)} />
      )}
    </main>
  );
}
