import {
  CloseCircleOutlined,
  CloseOutlined,
  ExclamationCircleOutlined,
  FileTextOutlined,
  RobotOutlined,
  ToolOutlined,
  UnorderedListOutlined,
} from "@ant-design/icons";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getToolCall, getTrajectory } from "../../../api/agent";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { AgentEvent, AgentTrajectoryOrder, ToolCallDetail } from "../../../models/agent";

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
const eventFilters = [
  { id: "all", label: "全部" },
  { id: "model", label: "模型" },
  { id: "tool", label: "工具" },
  { id: "approval", label: "审批" },
  { id: "task", label: "任务" },
  { id: "error", label: "异常" },
];
const eventNames: Record<string, string> = {
  "run.started": "开始运行",
  "run.completed": "运行完成",
  "run.failed": "运行失败",
  "run.cancelled": "运行取消",
  "turn.started": "Turn 开始",
  "turn.completed": "Turn 完成",
  "turn.failed": "Turn 失败",
  "step.started": "Step 开始",
  "step.completed": "Step 完成",
  "step.failed": "Step 失败",
  "system.prompt": "系统提示词",
  "request.snapshot": "模型请求快照",
  "model.attempt.started": "模型请求",
  "model.attempt.completed": "模型返回",
  "model.attempt.failed": "模型失败",
  "model.fallback": "切换备用模型",
  "model.unavailable": "模型不可用",
  "assistant.message": "助手回复",
  "tool.started": "Tool 调用",
  "tool.completed": "Tool 返回",
  "tool.failed": "Tool 失败",
  "approval.required": "等待审批",
  "approval.rejected": "审批已拒绝",
  "skill.loaded": "加载 Skill",
  "task.created": "创建任务",
  "task.updated": "任务进度",
  "task.completed": "任务完成",
};
const timezone = "Asia/Shanghai";
const shanghaiOffset = 8 * 60 * 60 * 1000;

export default function TrajectoryPanel({ sessionId, onClose }: Props) {
  const { options: eventTypes } = useDictionaryOptions("agent_event_type");
  const [raw, setEvents] = useState<AgentEvent[]>([]);
  const [selected, setSelected] = useState<AgentEvent | null>(null);
  const [toolCall, setToolCall] = useState<ToolCallDetail | null>(null);
  const [tab, setTab] = useState<InspectorTab>("overview");
  const [filter, setFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [order, setOrder] = useState<AgentTrajectoryOrder>("desc");
  const events = useMemo(() => groupCalls(raw, order), [raw, order]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [nextCursorId, setNextCursorId] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [timelineViewport, setTimelineViewport] = useState<{ left: number; right: number } | null>(
    null,
  );
  const eventListRef = useRef<HTMLDivElement>(null);
  const eventRowRefs = useRef(new Map<number, HTMLButtonElement>());
  const pendingScrollEventId = useRef<number | null>(null);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError("");
    setEvents([]);
    setSelected(null);
    setTimelineViewport(null);
    getTrajectory(sessionId, { order })
      .then((page) => {
        if (!alive) return;
        setEvents(page.events || []);
        setNextCursorId(page.nextCursorId || 0);
        setHasMore(page.hasMore);
      })
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : "轨迹加载失败"))
      .finally(() => setLoading(false));
    return () => {
      alive = false;
    };
  }, [sessionId, order]);

  useEffect(() => {
    setToolCall(null);
    if (!selected?.callId) return;
    let active = true;
    getToolCall(sessionId, selected.callId)
      .then((result) => {
        if (active) setToolCall(result.toolCall);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [selected, sessionId]);

  const visibleEvents = useMemo(
    () =>
      events.filter((event) => {
        const category = eventCategory(event.type);
        const text =
          `${eventName(event.type, eventTypes)} ${eventSummary(event)} ${event.type} ${JSON.stringify(event.data)}`.toLowerCase();
        return (
          (filter === "all" || filter === category) &&
          (!search || text.includes(search.toLowerCase()))
        );
      }),
    [events, eventTypes, filter, search],
  );

  const groups = useMemo(() => {
    const grouped = new Map<string, AgentEvent[]>();
    for (const event of visibleEvents) {
      const key = dateKey(event.createdAt);
      grouped.set(key, [...(grouped.get(key) || []), event]);
    }
    return [...grouped.entries()];
  }, [visibleEvents]);

  const timestamps = events
    .map((event) => Date.parse(event.createdAt || ""))
    .filter(Number.isFinite);
  const timelineStart = timestamps.length ? Math.min(...timestamps) : 0;
  const timelineEnd = timestamps.length ? Math.max(...timestamps) : 0;
  const timelineSpan = Math.max(timelineEnd - timelineStart, 1);
  const timelinePosition = useCallback(
    (timestamp: number) =>
      timelineEnd === timelineStart ? 50 : 2 + ((timestamp - timelineStart) / timelineSpan) * 96,
    [timelineEnd, timelineSpan, timelineStart],
  );
  const timelineTicks = useMemo(
    () => makeTimelineTicks(timelineStart, timelineEnd),
    [timelineEnd, timelineStart],
  );
  const visibleEventById = useMemo(
    () => new Map(visibleEvents.map((event) => [event.id, event])),
    [visibleEvents],
  );

  const updateTimelineViewport = useCallback(() => {
    const list = eventListRef.current;
    if (!list || !visibleEventById.size) return;
    const bounds = list.getBoundingClientRect();
    let firstTimestamp = Number.POSITIVE_INFINITY;
    let lastTimestamp = Number.NEGATIVE_INFINITY;
    for (const row of list.querySelectorAll<HTMLElement>("[data-event-id]")) {
      const rowBounds = row.getBoundingClientRect();
      if (rowBounds.bottom <= bounds.top || rowBounds.top >= bounds.bottom) continue;
      const event = visibleEventById.get(Number(row.dataset.eventId));
      const timestamp = Date.parse(event?.createdAt || "");
      if (!Number.isFinite(timestamp)) continue;
      firstTimestamp = Math.min(firstTimestamp, timestamp);
      lastTimestamp = Math.max(lastTimestamp, timestamp);
    }
    if (Number.isFinite(firstTimestamp) && Number.isFinite(lastTimestamp)) {
      setTimelineViewport({
        left: timelinePosition(firstTimestamp),
        right: timelinePosition(lastTimestamp),
      });
    }
  }, [timelinePosition, visibleEventById]);

  useEffect(() => {
    const next = (selected && visibleEventById.get(selected.id)) || visibleEvents[0] || null;
    if (selected !== next) setSelected(next);
  }, [selected, visibleEventById, visibleEvents]);

  useEffect(() => {
    updateTimelineViewport();
  }, [visibleEvents, events, updateTimelineViewport]);

  useEffect(() => {
    const pendingEventId = pendingScrollEventId.current;
    if (pendingEventId === null) return;
    const row = eventRowRefs.current.get(pendingEventId);
    if (!row) return;
    row.scrollIntoView({ behavior: "smooth", block: "center" });
    pendingScrollEventId.current = null;
  }, [visibleEvents]);

  const selectEvent = (event: AgentEvent) => {
    setSelected(event);
    setTab("overview");
    if (visibleEventById.has(event.id)) {
      eventRowRefs.current.get(event.id)?.scrollIntoView({ behavior: "smooth", block: "center" });
      return;
    }
    pendingScrollEventId.current = event.id;
    setFilter("all");
    setSearch("");
  };

  const loadMore = async () => {
    if (loadingMore || !hasMore) return;
    setLoadingMore(true);
    setError("");
    try {
      const page = await getTrajectory(sessionId, {
        order,
        cursorId: nextCursorId,
      });
      setEvents((current) => [...current, ...(page.events || [])]);
      setNextCursorId(page.nextCursorId || nextCursorId);
      setHasMore(page.hasMore);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "加载更多事件失败");
    } finally {
      setLoadingMore(false);
    }
  };

  return (
    <div className="trajectory-backdrop" role="dialog" aria-modal="true" aria-label="运行轨迹">
      <section className="trajectory-panel">
        <header className="trajectory-head">
          <div>
            <span className="trace-mark">
              <RobotOutlined />
            </span>
            <div>
              <strong>运行轨迹</strong>
              <small>
                会话 #{sessionId} · {events.length} 条事件 · 北京时间
              </small>
            </div>
          </div>
          <button onClick={onClose} aria-label="关闭轨迹" title="关闭轨迹">
            <CloseOutlined />
          </button>
        </header>

        <div className="timeline-strip" aria-label="运行时间轴">
          <div className="timeline-meta">
            <strong>会话时间轴</strong>
            <div className="timeline-legend" aria-label="事件状态图例">
              <span>
                <i className="status-recorded" />
                已记录
              </span>
              <span>
                <i className="status-running" />
                进行中
              </span>
              <span>
                <i className="status-pending" />
                待审批
              </span>
              <span>
                <i className="status-success" />
                成功
              </span>
              <span>
                <i className="status-failed" />
                失败/取消
              </span>
            </div>
          </div>
          <div className="timeline-track">
            <span className="timeline-line" />
            {timelineViewport && (
              <span
                className="timeline-viewport"
                aria-hidden="true"
                style={{
                  left: `${timelineViewport.left}%`,
                  width: `${Math.max(timelineViewport.right - timelineViewport.left, 1)}%`,
                }}
              />
            )}
            {timelineTicks.map((tick) => (
              <span
                className="timeline-tick"
                key={tick.timestamp}
                style={{ left: `${timelinePosition(tick.timestamp)}%` }}
              >
                <i />
                <time>{tick.label}</time>
              </span>
            ))}
            {events
              .filter((event) => eventCategory(event.type) !== "system")
              .map((event) => (
                <button
                  key={event.id}
                  type="button"
                  className={`timeline-node status-${eventStatus(event)} ${selected?.id === event.id ? "selected" : ""}`}
                  style={{ left: `${timelinePosition(Date.parse(event.createdAt || ""))}%` }}
                  title={`${eventName(event.type, eventTypes)} · ${formatDate(event.createdAt)} · ${eventSummary(event)}`}
                  aria-label={`${eventName(event.type, eventTypes)}，${formatDate(event.createdAt)}`}
                  onClick={() => selectEvent(event)}
                />
              ))}
          </div>
          <div className="timeline-times">
            <time>{formatAxisTime(timelineStart)}</time>
            <span>{events.length ? formatDuration(timelineSpan) : "无事件"}</span>
            <time>{formatAxisTime(timelineEnd)}</time>
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
                {eventFilters.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    className={filter === item.id ? "active" : ""}
                    aria-pressed={filter === item.id}
                    onClick={() => setFilter(item.id)}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
              <div className="trace-controls">
                <div className="trace-order" role="group" aria-label="时间顺序">
                  <span>顺序</span>
                  <button
                    type="button"
                    className={order === "desc" ? "active" : ""}
                    aria-pressed={order === "desc"}
                    onClick={() => setOrder("desc")}
                  >
                    最新优先
                  </button>
                  <button
                    type="button"
                    className={order === "asc" ? "active" : ""}
                    aria-pressed={order === "asc"}
                    onClick={() => setOrder("asc")}
                  >
                    最早优先
                  </button>
                </div>
              </div>
              {error && (
                <p className="trace-error trace-toolbar-error" role="alert">
                  {error}
                </p>
              )}
            </div>
            <div className="event-list" ref={eventListRef} onScroll={updateTimelineViewport}>
              {loading && <p className="trace-empty">正在读取持久化轨迹…</p>}
              {groups.map(([day, dayEvents]) => (
                <section className="event-group" key={day}>
                  <h3>
                    <time dateTime={day}>{formatDayLabel(day)}</time>
                    <span>{dayEvents.length}</span>
                  </h3>
                  {dayEvents.map((event) => (
                    <button
                      key={event.id}
                      ref={(element) => {
                        if (element) eventRowRefs.current.set(event.id, element);
                        else eventRowRefs.current.delete(event.id);
                      }}
                      data-event-id={event.id}
                      type="button"
                      className={`event-item status-${eventStatus(event)} ${selected?.id === event.id ? "active" : ""}`}
                      aria-pressed={selected?.id === event.id}
                      onClick={() => selectEvent(event)}
                    >
                      <span className={`event-kind status-${eventStatus(event)}`}>
                        {eventIcon(event.type)}
                      </span>
                      <div>
                        <strong>{eventName(event.type, eventTypes)}</strong>
                        <small>{eventSummary(event)}</small>
                      </div>
                      <time dateTime={event.createdAt} title={formatDate(event.createdAt)}>
                        {formatClock(event.createdAt)}
                      </time>
                    </button>
                  ))}
                </section>
              ))}
              {hasMore && (
                <button
                  type="button"
                  className="load-more"
                  disabled={loadingMore}
                  onClick={() => void loadMore()}
                >
                  {loadingMore ? "正在加载…" : order === "desc" ? "加载更早事件" : "加载后续事件"}
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
                    <span className={`event-kind status-${eventStatus(selected)}`}>
                      {eventIcon(selected.type)}
                    </span>
                    <div>
                      <strong>{eventName(selected.type, eventTypes)}</strong>
                      <small>
                        事件 #{selected.id} · {selected.type}
                      </small>
                    </div>
                  </div>
                  <span className={`state status-${eventStatus(selected)}`}>
                    {eventStatusLabel(selected)}
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

// 把同一次调用的审计记录合成一行，后续分页加载时重新配对。
function groupCalls(raw: AgentEvent[], order: AgentTrajectoryOrder) {
  const calls = new Map<string, AgentEvent>();
  const rows: AgentEvent[] = [];
  for (const event of [...raw].sort((a, b) => a.id - b.id)) {
    if (["turn.started", "turn.completed", "step.started", "step.completed"].includes(event.type))
      continue;
    const { type, data } = event;
    const run = event.runId ?? data.runId;
    const step = event.stepNo ?? data.stepNo;
    const call = event.callId || data.callId;
    let key = "";
    if (run && call && ["tool.started", "tool.completed", "tool.failed"].includes(type))
      key = JSON.stringify([run, "tool", call]);
    if (run && step && (type === "request.snapshot" || type.startsWith("model.attempt.")))
      key = JSON.stringify([run, data.turnNo ?? 1, step, data.attemptNo ?? 1]);
    if (!key) {
      rows.push(event);
      continue;
    }
    const row = calls.get(key);
    if (row) {
      row.data = { ...row.data, ...data };
      row.type = type;
      row.id = event.id;
      row.createdAt = event.createdAt;
    } else {
      const first = { ...event, data: { ...data } };
      calls.set(key, first);
      rows.push(first);
    }
  }
  return rows.sort((a, b) => (order === "asc" ? a.id - b.id : b.id - a.id));
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
          toolCall?.result ??
          data.response ??
          data.content ??
          data.trace ??
          data.error ??
          (event.type.endsWith(".failed") ? data.message : null)
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
  const icons = {
    model: <RobotOutlined />,
    tool: <ToolOutlined />,
    approval: <ExclamationCircleOutlined />,
    task: <UnorderedListOutlined />,
    error: <CloseCircleOutlined />,
    system: <FileTextOutlined />,
  };
  return icons[eventCategory(type) as keyof typeof icons];
}

function eventName(type: string, eventTypes: ReturnType<typeof useDictionaryOptions>["options"]) {
  if (["tool.started", "tool.completed", "tool.failed"].includes(type)) return "Tool 调用";
  if (type === "request.snapshot" || type.startsWith("model.attempt.")) return "模型请求";
  const localizedName = dictionaryName(eventTypes, type);
  return localizedName && localizedName !== type ? localizedName : eventNames[type] || type;
}

function eventSummary(event: AgentEvent) {
  const data = event.data;
  if (event.type === "system.prompt") return "系统提示词已记录";
  const error = data.error;
  const errorMessage =
    error && typeof error === "object" && "message" in error
      ? (error as { message?: unknown }).message
      : error;
  const tool = firstText(data.toolName, data.tool, data.name);
  if (event.type.includes("failed") || event.type.includes("cancelled")) {
    return firstText(errorMessage, data.message, data.reason) || "操作未能完成";
  }
  if (event.type === "approval.required") return `等待确认：${tool || "待审批操作"}`;
  if (event.type === "approval.rejected") return `已拒绝：${tool || "操作未执行"}`;
  if (event.type.startsWith("tool.")) {
    return firstText(tool, data.message, data.summary) || "工具调用状态已记录";
  }
  if (event.type.startsWith("model.")) {
    return firstText(data.model, data.message, data.reason) || "模型请求状态已记录";
  }
  if (event.type.startsWith("task.")) {
    return firstText(data.stage, data.message, data.outcome) || "资源任务状态已更新";
  }
  return firstText(data.message, data.outcome, data.reason, data.content) || "事件已记录";
}

function firstText(...values: unknown[]) {
  return (
    values
      .find((value): value is string => typeof value === "string" && value.trim() !== "")
      ?.trim() || ""
  );
}

function eventStatus(event: AgentEvent) {
  if (event.type.includes("failed") || event.type.includes("cancelled")) return "failed";
  if (event.type === "approval.required") return "pending";
  if (event.type.endsWith(".completed")) return "success";
  if (event.type.endsWith(".started")) return "running";
  return "recorded";
}

function eventStatusLabel(event: AgentEvent) {
  return {
    recorded: "已记录",
    running: "进行中",
    pending: "待审批",
    success: "成功",
    failed: "失败/取消",
  }[eventStatus(event)];
}

function formatClock(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        timeZone: timezone,
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        fractionalSecondDigits: 3,
        hourCycle: "h23",
      }).format(new Date(value))
    : "—";
}

function formatDate(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        timeZone: timezone,
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        fractionalSecondDigits: 3,
        hourCycle: "h23",
      }).format(new Date(value))
    : "—";
}

function formatAxisTime(timestamp: number) {
  return timestamp
    ? new Intl.DateTimeFormat("zh-CN", {
        timeZone: timezone,
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hourCycle: "h23",
      }).format(new Date(timestamp))
    : "—";
}

function dateKey(value?: string) {
  const timestamp = Date.parse(value || "");
  if (!Number.isFinite(timestamp)) return "unknown";
  const date = new Date(timestamp + shanghaiOffset);
  return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, "0")}-${String(date.getUTCDate()).padStart(2, "0")}`;
}

function formatDayLabel(value: string) {
  if (value === "unknown") return "日期未知";
  return new Intl.DateTimeFormat("zh-CN", {
    timeZone: "UTC",
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(new Date(`${value}T00:00:00Z`));
}

function makeTimelineTicks(start: number, end: number) {
  if (!start || end <= start) return [];
  const firstDay = Math.floor((start + shanghaiOffset) / 86_400_000);
  const lastDay = Math.floor((end + shanghaiOffset) / 86_400_000);
  const step = lastDay - firstDay > 14 ? 7 : 1;
  const ticks: { timestamp: number; label: string }[] = [];
  for (let day = firstDay + 1; day <= lastDay; day += step) {
    const timestamp = day * 86_400_000 - shanghaiOffset;
    ticks.push({
      timestamp,
      label: new Intl.DateTimeFormat("zh-CN", {
        timeZone: timezone,
        month: "numeric",
        day: "numeric",
      }).format(new Date(timestamp)),
    });
  }
  return ticks;
}

function formatDuration(milliseconds: number) {
  const totalSeconds = Math.floor(milliseconds / 1000);
  if (totalSeconds < 60) return `${totalSeconds} 秒`;
  const seconds = totalSeconds % 60;
  const totalMinutes = Math.floor(totalSeconds / 60);
  if (totalMinutes < 60) return `${totalMinutes} 分 ${seconds} 秒`;
  const minutes = totalMinutes % 60;
  const totalHours = Math.floor(totalMinutes / 60);
  if (totalHours < 24) return `${totalHours} 小时 ${minutes} 分`;
  return `${Math.floor(totalHours / 24)} 天 ${totalHours % 24} 小时`;
}
