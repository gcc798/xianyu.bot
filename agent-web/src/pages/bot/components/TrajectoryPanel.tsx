import { useEffect, useMemo, useState } from "react";
import { getToolCall, getTrajectory } from "../../../api/agent";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { AgentEvent, ToolCallDetail } from "../../../models/agent";

type Props = { sessionId: number; onClose: () => void };
type InspectorTab =
  "overview" | "input" | "output" | "prompt" | "schema" | "timing" | "token" | "raw";

const tabs: { id: InspectorTab; label: string }[] = [
  { id: "overview", label: "概览" },
  { id: "input", label: "输入" },
  { id: "output", label: "输出" },
  { id: "prompt", label: "提示词" },
  { id: "schema", label: "Tool Schema" },
  { id: "timing", label: "时间" },
  { id: "token", label: "Token" },
  { id: "raw", label: "Raw JSON" },
];

export default function TrajectoryPanel({ sessionId, onClose }: Props) {
  const { options: eventTypes } = useDictionaryOptions("agent_event_type");
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [selected, setSelected] = useState<AgentEvent | null>(null);
  const [toolCall, setToolCall] = useState<ToolCallDetail | null>(null);
  const [tab, setTab] = useState<InspectorTab>("overview");
  const [filter, setFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [nextAfterId, setNextAfterId] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    getTrajectory(sessionId)
      .then((page) => {
        if (!alive) return;
        setEvents(page.events || []);
        setSelected(page.events?.at(-1) || null);
        setNextAfterId(page.nextAfterId || 0);
        setHasMore(page.hasMore);
      })
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : "轨迹加载失败"))
      .finally(() => setLoading(false));
    return () => {
      alive = false;
    };
  }, [sessionId]);

  useEffect(() => {
    setToolCall(null);
    if (!selected?.callId) return;
    getToolCall(sessionId, selected.callId)
      .then((result) => setToolCall(result.toolCall))
      .catch(() => undefined);
  }, [selected, sessionId]);

  const visible = useMemo(
    () =>
      events.filter((event) => {
        const category = eventCategory(event.type);
        const text = `${event.type} ${JSON.stringify(event.data)}`.toLowerCase();
        return (
          (filter === "all" || filter === category) &&
          (!search || text.includes(search.toLowerCase()))
        );
      }),
    [events, filter, search],
  );

  const groups = useMemo(() => {
    const grouped = new Map<string, AgentEvent[]>();
    for (const event of visible) {
      const key = event.stepNo
        ? `Run #${event.runId || "—"} · Step ${event.stepNo}`
        : event.runId
          ? `Run #${event.runId}`
          : "Session";
      grouped.set(key, [...(grouped.get(key) || []), event]);
    }
    return [...grouped.entries()];
  }, [visible]);

  const timelineStart = events.length ? new Date(events[0].createdAt || 0).getTime() : 0;
  const timelineEnd = events.length ? new Date(events.at(-1)?.createdAt || 0).getTime() : 0;
  const timelineSpan = Math.max(timelineEnd - timelineStart, 1);

  return (
    <div
      className="trajectory-backdrop"
      role="dialog"
      aria-modal="true"
      aria-label="Agent 审计轨迹"
    >
      <section className="trajectory-panel">
        <header className="trajectory-head">
          <div>
            <span className="trace-mark">⌘</span>
            <div>
              <strong>Agent Trajectory</strong>
              <small>
                SESSION #{sessionId} · {events.length} EVENTS
              </small>
            </div>
          </div>
          <button onClick={onClose} aria-label="关闭轨迹">
            ×
          </button>
        </header>

        <div className="timeline-strip" aria-label="运行时间轴">
          <span className="timeline-line" />
          {events
            .filter((event) =>
              ["model", "tool", "approval", "task"].includes(eventCategory(event.type)),
            )
            .map((event) => (
              <button
                key={event.id}
                className={`timeline-node ${eventCategory(event.type)} ${event.type.includes("failed") ? "failed" : ""}`}
                style={{
                  left: `${((new Date(event.createdAt || 0).getTime() - timelineStart) / timelineSpan) * 96 + 2}%`,
                }}
                title={`${dictionaryName(eventTypes, event.type)} · ${formatClock(event.createdAt)}`}
                onClick={() => setSelected(event)}
              />
            ))}
          <div className="timeline-times">
            <span>{formatClock(events[0]?.createdAt)}</span>
            <span>{duration(timelineSpan)}</span>
            <span>{formatClock(events.at(-1)?.createdAt)}</span>
          </div>
        </div>

        <div className="trajectory-body">
          <aside className="event-browser">
            <div className="event-tools">
              <input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="搜索事件或内容"
                aria-label="搜索轨迹"
              />
              <div className="event-filters">
                {["all", "model", "tool", "approval", "error"].map((item) => (
                  <button
                    key={item}
                    className={filter === item ? "active" : ""}
                    onClick={() => setFilter(item)}
                  >
                    {item === "all" ? "全部" : item}
                  </button>
                ))}
              </div>
            </div>
            <div className="event-list">
              {loading && <p className="trace-empty">正在读取持久化轨迹…</p>}
              {error && <p className="trace-error">{error}</p>}
              {groups.map(([name, items]) => (
                <section className="event-group" key={name}>
                  <h3>
                    {name}
                    <span>{items.length}</span>
                  </h3>
                  {items.map((event) => (
                    <button
                      key={event.id}
                      className={`event-item ${selected?.id === event.id ? "active" : ""}`}
                      onClick={() => {
                        setSelected(event);
                        setTab("overview");
                      }}
                    >
                      <span className={`event-kind ${eventCategory(event.type)}`}>
                        {eventIcon(event.type)}
                      </span>
                      <div>
                        <strong>{dictionaryName(eventTypes, event.type)}</strong>
                        <small>{eventSummary(event)}</small>
                      </div>
                      <time>{formatClock(event.createdAt)}</time>
                    </button>
                  ))}
                </section>
              ))}
              {hasMore && (
                <button
                  className="load-more"
                  onClick={() => {
                    setLoading(true);
                    getTrajectory(sessionId, nextAfterId)
                      .then((page) => {
                        setEvents((items) => [...items, ...(page.events || [])]);
                        setNextAfterId(page.nextAfterId || nextAfterId);
                        setHasMore(page.hasMore);
                      })
                      .catch((cause: unknown) =>
                        setError(cause instanceof Error ? cause.message : "加载更多失败"),
                      )
                      .finally(() => setLoading(false));
                  }}
                >
                  加载更多事件
                </button>
              )}
              {!loading && !groups.length && <p className="trace-empty">没有符合条件的事件</p>}
            </div>
          </aside>

          <article className="inspector">
            {selected ? (
              <>
                <header className="inspector-title">
                  <div>
                    <span className={`event-kind ${eventCategory(selected.type)}`}>
                      {eventIcon(selected.type)}
                    </span>
                    <div>
                      <strong>{dictionaryName(eventTypes, selected.type)}</strong>
                      <small>
                        EVENT #{selected.id} · {selected.type}
                      </small>
                    </div>
                  </div>
                  <span className={selected.type.includes("failed") ? "state failed" : "state"}>
                    {selected.type.includes("failed") ? "FAILED" : "RECORDED"}
                  </span>
                </header>
                <nav className="inspector-tabs">
                  {tabs.map((item) => (
                    <button
                      key={item.id}
                      className={tab === item.id ? "active" : ""}
                      onClick={() => setTab(item.id)}
                    >
                      {item.label}
                    </button>
                  ))}
                </nav>
                <div className="inspector-content">{renderTab(tab, selected, toolCall)}</div>
              </>
            ) : (
              <p className="trace-empty">选择一条事件查看完整审计数据</p>
            )}
          </article>
        </div>
      </section>
    </div>
  );
}

function renderTab(tab: InspectorTab, event: AgentEvent, toolCall: ToolCallDetail | null) {
  const data = event.data;
  if (tab === "overview")
    return (
      <div className="overview-grid">
        <Info label="事件类型" value={event.type} />
        <Info label="Event ID" value={`#${event.id}`} />
        <Info label="Run" value={event.runId ? `#${event.runId}` : "—"} />
        <Info label="Step" value={event.stepNo ? `#${event.stepNo}` : "—"} />
        <Info label="Call ID" value={event.callId || "—"} wide />
        <Info label="记录时间" value={formatDate(event.createdAt)} wide />
        <section className="summary-card">
          <small>事件摘要</small>
          <p>{eventSummary(event)}</p>
        </section>
      </div>
    );
  if (tab === "input")
    return (
      <Code
        value={toolCall?.arguments ?? data.arguments ?? data.messages ?? null}
        empty="该事件没有输入数据"
      />
    );
  if (tab === "output")
    return (
      <Code
        value={
          toolCall?.result ?? data.response ?? data.content ?? data.trace ?? data.error ?? null
        }
        empty="该事件没有输出数据"
      />
    );
  if (tab === "prompt")
    return (
      <Code value={data.content ?? extractSystemMessage(data.messages)} empty="该事件没有 Prompt" />
    );
  if (tab === "schema") return <Code value={data.tools ?? null} empty="该事件没有 Tool Schema" />;
  if (tab === "timing")
    return (
      <div className="overview-grid">
        <Info label="服务端记录" value={formatDate(event.createdAt)} wide />
        <Info
          label="Tool 耗时"
          value={
            toolCall ? `${toolCall.durationMillis} ms` : `${Number(data.durationMillis || 0)} ms`
          }
        />
        <Info label="状态" value={toolCall?.status || String(data.status || "recorded")} />
      </div>
    );
  if (tab === "token")
    return (
      <Code
        value={(data.responseMeta as { usage?: unknown } | undefined)?.usage ?? null}
        empty="模型供应商未返回 Token 数据"
      />
    );
  return <Code value={{ event, toolCall }} />;
}

function Info({ label, value, wide = false }: { label: string; value: string; wide?: boolean }) {
  return (
    <div className={`info-card ${wide ? "wide" : ""}`}>
      <small>{label}</small>
      <strong>{value}</strong>
    </div>
  );
}

function Code({ value, empty = "暂无数据" }: { value: unknown; empty?: string }) {
  if (value === null || value === undefined || value === "")
    return <p className="trace-empty">{empty}</p>;
  return (
    <pre className="json-view">
      <code>{typeof value === "string" ? value : JSON.stringify(value, null, 2)}</code>
    </pre>
  );
}

function extractSystemMessage(messages: unknown) {
  if (!Array.isArray(messages)) return null;
  return (
    messages.find(
      (message) =>
        message && typeof message === "object" && (message as { role?: string }).role === "system",
    ) || null
  );
}

function eventCategory(type: string) {
  if (type.includes("failed") || type.includes("cancelled")) return "error";
  if (
    type.startsWith("model.") ||
    type === "request.snapshot" ||
    type === "system.prompt" ||
    type === "assistant.message"
  )
    return "model";
  if (type.startsWith("tool.")) return "tool";
  if (type.startsWith("approval.")) return "approval";
  if (type.startsWith("task.")) return "task";
  return "system";
}

function eventIcon(type: string) {
  return { model: "AI", tool: "⌘", approval: "!", task: "↓", error: "×", system: "·" }[
    eventCategory(type)
  ];
}

function eventSummary(event: AgentEvent) {
  const data = event.data;
  return String(
    data.tool ||
      data.model ||
      data.skill ||
      data.outcome ||
      data.reason ||
      data.message ||
      data.content ||
      event.type,
  ).slice(0, 90);
}

function formatClock(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        fractionalSecondDigits: 3,
      }).format(new Date(value))
    : "—";
}
function formatDate(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "medium" }).format(
        new Date(value),
      )
    : "—";
}
function duration(ms: number) {
  return ms < 1000 ? `${ms} ms` : `${(ms / 1000).toFixed(2)} s`;
}
