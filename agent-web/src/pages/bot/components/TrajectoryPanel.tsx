import {
  CloseCircleOutlined,
  CloseOutlined,
  ExclamationCircleOutlined,
  FileTextOutlined,
  RobotOutlined,
  SearchOutlined,
  ToolOutlined,
  UnorderedListOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getToolCall, getTrajectory, streamEvents } from "../../../api/agent";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { AgentEvent, ToolCallDetail } from "../../../models/agent";
import { buildTrajectoryRows, type TrajectoryLane, type TrajectoryRow } from "../trajectory";

type Props = { sessionId: number };
type InspectorTab =
  "overview" | "input" | "output" | "prompt" | "schema" | "timing" | "token" | "raw";

const eventNames: Record<string, string> = {
  "run.started": "开始运行",
  "run.completed": "运行完成",
  "run.failed": "运行失败",
  "run.cancelled": "运行取消",
  "system.prompt": "系统提示词",
  "request.snapshot": "模型请求",
  "model.attempt.started": "模型请求",
  "model.attempt.completed": "模型返回",
  "model.attempt.failed": "模型失败",
  "model.fallback": "切换备用模型",
  "model.unavailable": "模型不可用",
  "assistant.message": "助手消息",
  "tool.started": "工具调用",
  "tool.completed": "工具返回",
  "tool.failed": "工具失败",
  "approval.required": "等待审批",
  "approval.rejected": "审批已拒绝",
  "skill.loaded": "加载 Skill",
  "task.created": "创建任务",
  "task.updated": "任务进度",
  "task.completed": "任务完成",
};
const laneLabels: Record<TrajectoryLane, string> = {
  input: "输入",
  model: "模型",
  tool: "工具",
};
const lanes = Object.keys(laneLabels) as TrajectoryLane[];
const timezone = "Asia/Shanghai";

export default function TrajectoryPanel({ sessionId }: Props) {
  const { options: eventTypes } = useDictionaryOptions("agent_event_type");
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [selectedRowId, setSelectedRowId] = useState("");
  const [toolCall, setToolCall] = useState<ToolCallDetail | null>(null);
  const [tab, setTab] = useState<InspectorTab>("overview");
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [nextCursorId, setNextCursorId] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [live, setLive] = useState(false);
  const [viewport, setViewport] = useState({ left: 0, width: 100 });
  const listRef = useRef<HTMLDivElement>(null);
  const rowRefs = useRef(new Map<string, HTMLButtonElement>());

  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    setLoading(true);
    setError("");
    setEvents([]);
    setSelectedRowId("");
    setLive(false);
    void getTrajectory(sessionId, { order: "desc" })
      .then((page) => {
        if (!active) return;
        setEvents(page.events || []);
        setNextCursorId(page.nextCursorId || 0);
        setHasMore(page.hasMore);
        setLoading(false);
        setLive(true);
        return streamEvents(sessionId, page.latestEventId || 0, controller.signal, (event) => {
          if (event.id) setEvents((current) => mergeEvents(current, [event]));
        });
      })
      .catch((cause: unknown) => {
        if (!active || controller.signal.aborted) return;
        setLoading(false);
        setLive(false);
        setError(cause instanceof Error ? cause.message : "轨迹加载失败");
      });
    return () => {
      active = false;
      controller.abort();
    };
  }, [sessionId]);

  const rows = useMemo(() => buildTrajectoryRows(events), [events]);
  const selected = rows.find((row) => row.rowId === selectedRowId) || null;
  const selectedToolSchema = useMemo(() => findToolSchema(rows, selected), [rows, selected]);
  const visibleRows = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) return rows;
    return rows.filter((row) =>
      `${eventName(row, eventTypes)} ${rowInputPreview(row)} ${rowOutputPreview(row)} ${row.type} ${JSON.stringify(row.data)}`
        .toLowerCase()
        .includes(query),
    );
  }, [eventTypes, rows, search]);
  const runOrdinals = useMemo(() => {
    const ids = [...new Set(rows.map(runID).filter(Boolean))];
    return new Map(ids.map((id, index) => [id, index + 1]));
  }, [rows]);
  const groups = useMemo(() => groupByRun(visibleRows), [visibleRows]);
  const rowIndex = useMemo(() => new Map(rows.map((row, index) => [row.rowId, index])), [rows]);

  useEffect(() => {
    setToolCall(null);
    if (!selected?.callId) return;
    let active = true;
    getToolCall(sessionId, selected.callId)
      .then(({ toolCall: detail }) => {
        if (active) setToolCall(detail);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [selected?.callId, sessionId]);

  const updateViewport = useCallback(() => {
    const list = listRef.current;
    if (!list || !rows.length) return;
    const bounds = list.getBoundingClientRect();
    const visibleIndexes: number[] = [];
    for (const element of list.querySelectorAll<HTMLElement>("[data-row-id]")) {
      const rowBounds = element.getBoundingClientRect();
      if (rowBounds.bottom <= bounds.top || rowBounds.top >= bounds.bottom) continue;
      const index = rowIndex.get(element.dataset.rowId || "");
      if (index !== undefined) visibleIndexes.push(index);
    }
    if (!visibleIndexes.length) return;
    const left = timelinePosition(Math.min(...visibleIndexes), rows.length);
    const right = timelinePosition(Math.max(...visibleIndexes), rows.length);
    setViewport({ left, width: Math.max(1.5, right - left) });
  }, [rowIndex, rows.length]);

  useEffect(() => {
    updateViewport();
  }, [updateViewport, visibleRows]);

  const selectRow = (row: TrajectoryRow, clearSearch = false) => {
    if (clearSearch) setSearch("");
    setSelectedRowId(row.rowId);
    setTab("overview");
    requestAnimationFrame(() =>
      rowRefs.current.get(row.rowId)?.scrollIntoView({ block: "center", behavior: "smooth" }),
    );
  };

  const loadEarlier = async () => {
    if (!hasMore || loadingMore) return;
    setLoadingMore(true);
    setError("");
    try {
      const page = await getTrajectory(sessionId, { order: "desc", cursorId: nextCursorId });
      setEvents((current) => mergeEvents(current, page.events || []));
      setNextCursorId(page.nextCursorId || nextCursorId);
      setHasMore(page.hasMore);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "更早的轨迹加载失败");
    } finally {
      setLoadingMore(false);
    }
  };

  const summary = trajectorySummary(rows);

  return (
    <section className="trajectory-workspace" aria-label="运行轨迹">
      <header className="trajectory-overview">
        <div className="trajectory-toolbar">
          <div className="trajectory-metrics" aria-label="轨迹统计">
            <span>时长 {formatDuration(summary.duration)}</span>
            <span>轮次 {runOrdinals.size}</span>
            <span>模型 {summary.modelCalls}</span>
            <span>工具 {summary.toolCalls}</span>
          </div>
          <label className="trajectory-search">
            <SearchOutlined />
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="搜索轨迹"
              aria-label="搜索轨迹"
            />
          </label>
        </div>
        <div className="trajectory-map" aria-label="输入、模型和工具调用时间轴">
          <div className="trajectory-lane-labels" aria-hidden="true">
            {lanes.map((lane) => (
              <span key={lane}>{laneLabels[lane]}</span>
            ))}
          </div>
          <div className="trajectory-map-track">
            {lanes.map((lane) => (
              <div className={`trajectory-lane lane-${lane}`} key={lane}>
                {rows.map((row, index) =>
                  row.lane === lane ? (
                    <button
                      key={row.rowId}
                      type="button"
                      className={`trajectory-node status-${row.status}${selectedRowId === row.rowId ? " selected" : ""}`}
                      style={{ left: `${timelinePosition(index, rows.length)}%` }}
                      aria-label={`${eventName(row, eventTypes)}，${formatDate(row.startedAt)}`}
                      title={`${eventName(row, eventTypes)} · ${rowOutputPreview(row) || rowInputPreview(row)}`}
                      onClick={() => selectRow(row, true)}
                    />
                  ) : null,
                )}
              </div>
            ))}
            {!!rows.length && (
              <span
                className="trajectory-viewport"
                style={{ left: `${viewport.left}%`, width: `${viewport.width}%` }}
              />
            )}
          </div>
        </div>
      </header>

      <div className={`trajectory-main${selected ? " inspector-open" : ""}`}>
        <div className="trajectory-event-list" ref={listRef} onScroll={updateViewport}>
          {error && (
            <p className="trajectory-notice error" role="alert">
              {error}
            </p>
          )}
          {loading && <TrajectorySkeleton />}
          {!loading && !rows.length && (
            <p className="trajectory-empty">当前会话还没有运行轨迹，发送消息后会在这里实时显示。</p>
          )}
          {!loading && !!rows.length && !visibleRows.length && (
            <p className="trajectory-empty">没有匹配的轨迹事件</p>
          )}
          {hasMore && (
            <button
              className="trajectory-load-more"
              type="button"
              disabled={loadingMore}
              onClick={() => void loadEarlier()}
            >
              {loadingMore ? "正在加载…" : "加载更早轨迹"}
            </button>
          )}
          {groups.map(([id, runRows]) => (
            <section className="trajectory-run" key={id}>
              <header>
                <span>{id === 0 ? "会话事件" : `第 ${runOrdinals.get(id) || 1} 轮`}</span>
                <small>
                  {id === 0 ? `${runRows.length} 条记录` : `Run #${id} · ${runRows.length} 步`}
                </small>
              </header>
              {runRows.map((row) => (
                <button
                  key={row.rowId}
                  ref={(element) => {
                    if (element) rowRefs.current.set(row.rowId, element);
                    else rowRefs.current.delete(row.rowId);
                  }}
                  data-row-id={row.rowId}
                  type="button"
                  className={`trajectory-event status-${row.status}${selectedRowId === row.rowId ? " active" : ""}`}
                  aria-pressed={selectedRowId === row.rowId}
                  onClick={() => selectRow(row)}
                >
                  <span className={`trajectory-kind kind-${row.lane}`}>{eventIcon(row)}</span>
                  <span className="trajectory-event-copy">
                    <strong>{eventName(row, eventTypes)}</strong>
                    <span>{rowInputPreview(row) || "—"}</span>
                    {(rowOutputPreview(row) || row.durationMillis !== undefined) && <i>→</i>}
                    <span>{rowOutputPreview(row)}</span>
                  </span>
                  <small>
                    {row.durationMillis !== undefined
                      ? formatDuration(row.durationMillis)
                      : formatClock(row.startedAt)}
                  </small>
                </button>
              ))}
            </section>
          ))}
          <span className={`trajectory-live${live ? " connected" : ""}`} role="status">
            {live ? "实时轨迹已连接" : loading ? "正在连接实时轨迹" : "实时轨迹未连接"}
          </span>
        </div>

        {selected && (
          <aside className="trajectory-inspector" aria-label="事件详情">
            <header className="inspector-title">
              <div>
                <span className={`trajectory-kind kind-${selected.lane}`}>
                  {eventIcon(selected)}
                </span>
                <div>
                  <strong>{eventName(selected, eventTypes)}</strong>
                  <small>
                    第 {runOrdinals.get(runID(selected)) || 1} 轮 ·{" "}
                    {selected.stepNo ? `第 ${selected.stepNo} 步` : `事件 #${selected.id}`}
                  </small>
                </div>
              </div>
              <button type="button" aria-label="关闭事件详情" onClick={() => setSelectedRowId("")}>
                <CloseOutlined />
              </button>
            </header>
            <nav className="inspector-tabs" aria-label="事件详情分类">
              {inspectorTabs(selected).map((item) => (
                <button
                  key={item.id}
                  type="button"
                  className={tab === item.id ? "active" : ""}
                  onClick={() => setTab(item.id)}
                >
                  {item.label}
                </button>
              ))}
            </nav>
            <div className="inspector-content">
              {renderTab(tab, selected, toolCall, selectedToolSchema)}
            </div>
          </aside>
        )}
      </div>
    </section>
  );
}

function inspectorTabs(row: TrajectoryRow): { id: InspectorTab; label: string }[] {
  const tabs: { id: InspectorTab; label: string }[] = [{ id: "overview", label: "概览" }];
  if (row.lane === "tool")
    tabs.push(
      { id: "input", label: "参数" },
      { id: "output", label: "结果" },
      { id: "schema", label: "Schema" },
    );
  else if (row.lane === "model")
    tabs.push(
      { id: "input", label: "输入" },
      { id: "output", label: "输出" },
      { id: "prompt", label: "提示词" },
      { id: "token", label: "Token" },
    );
  else tabs.push({ id: "output", label: "内容" });
  tabs.push({ id: "timing", label: "计时" }, { id: "raw", label: "原始数据" });
  return tabs;
}

function renderTab(
  tab: InspectorTab,
  row: TrajectoryRow,
  toolCall: ToolCallDetail | null,
  toolSchema: unknown,
) {
  const data = row.data;
  if (tab === "overview")
    return (
      <dl className="trace-definition">
        <Info
          label="层级"
          value={row.lane === "tool" ? "工具调用" : row.lane === "model" ? "模型消息" : "运行事件"}
        />
        <Info label="状态" value={statusLabel(row)} />
        <Info label="Run" value={runID(row) ? `#${runID(row)}` : "—"} />
        <Info label="Step" value={row.stepNo ? `#${row.stepNo}` : "—"} />
        <Info label="Call ID" value={row.callId || "—"} />
        <Info label="摘要" value={rowOutputPreview(row) || rowInputPreview(row) || "已记录"} />
      </dl>
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
          data.message ??
          null
        }
        empty="该事件没有输出数据"
      />
    );
  if (tab === "prompt")
    return (
      <Code value={data.content ?? extractSystemMessage(data.messages)} empty="该事件没有 Prompt" />
    );
  if (tab === "schema")
    return <Code value={toolSchema ?? data.tools ?? null} empty="该事件没有 Tool Schema" />;
  if (tab === "token")
    return (
      <Code
        value={(data.responseMeta as { usage?: unknown } | undefined)?.usage ?? null}
        empty="模型供应商未返回 Token 数据"
      />
    );
  if (tab === "timing")
    return (
      <dl className="trace-definition">
        <Info label="开始时间" value={formatDate(row.startedAt)} />
        <Info label="结束时间" value={formatDate(toolCall?.completedAt || row.completedAt)} />
        <Info
          label="总时长"
          value={row.durationMillis !== undefined ? formatDuration(row.durationMillis) : "未提供"}
        />
      </dl>
    );
  return <Code value={{ event: row, toolCall }} />;
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <>
      <dt>{label}</dt>
      <dd>{value}</dd>
    </>
  );
}

function Code({ value, empty = "暂无数据" }: { value: unknown; empty?: string }) {
  if (value === null || value === undefined || value === "")
    return <p className="trajectory-empty">{empty}</p>;
  return (
    <pre className="json-view">
      <code>{typeof value === "string" ? value : JSON.stringify(value, null, 2)}</code>
    </pre>
  );
}

function TrajectorySkeleton() {
  return (
    <div className="trajectory-skeleton" aria-label="正在读取轨迹">
      {Array.from({ length: 8 }, (_, index) => (
        <span key={index} />
      ))}
    </div>
  );
}

function groupByRun(rows: TrajectoryRow[]) {
  const groups = new Map<number, TrajectoryRow[]>();
  for (const row of rows) {
    const id = runID(row);
    groups.set(id, [...(groups.get(id) || []), row]);
  }
  return [...groups.entries()];
}

function mergeEvents(current: AgentEvent[], incoming: AgentEvent[]) {
  const events = new Map(current.map((event) => [event.id, event]));
  for (const event of incoming) events.set(event.id, event);
  return [...events.values()];
}

function trajectorySummary(rows: TrajectoryRow[]) {
  const timestamps = rows
    .flatMap((row) => [Date.parse(row.startedAt || ""), Date.parse(row.completedAt || "")])
    .filter(Number.isFinite);
  return {
    duration: timestamps.length ? Math.max(...timestamps) - Math.min(...timestamps) : 0,
    modelCalls: rows.filter((row) => row.rowId.startsWith("model:")).length,
    toolCalls: rows.filter((row) => row.rowId.startsWith("tool:")).length,
  };
}

function timelinePosition(index: number, total: number) {
  return total <= 1 ? 1.5 : 1.5 + (index / (total - 1)) * 97;
}

function runID(row: TrajectoryRow) {
  return row.runId || Number(row.data.runId || 0);
}

function findToolSchema(rows: TrajectoryRow[], selected: TrajectoryRow | null) {
  if (!selected || selected.lane !== "tool") return null;
  const code = firstText(selected.data.toolCode, selected.data.tool, selected.data.toolName);
  if (!code) return null;
  for (const row of rows) {
    if (!Array.isArray(row.data.tools)) continue;
    const schema = row.data.tools.find((tool) => {
      if (!tool || typeof tool !== "object") return false;
      const definition = tool as { name?: unknown; code?: unknown };
      return definition.name === code || definition.code === code;
    });
    if (schema) return schema;
  }
  return null;
}

function eventName(
  row: TrajectoryRow,
  eventTypes: ReturnType<typeof useDictionaryOptions>["options"],
) {
  if (row.rowId.startsWith("tool:"))
    return firstText(row.data.toolName, row.data.tool, row.data.toolCode) || "工具调用";
  if (row.rowId.startsWith("model:")) return firstText(row.data.model) || "模型请求";
  const localized = dictionaryName(eventTypes, row.type);
  return localized && localized !== row.type ? localized : eventNames[row.type] || row.type;
}

function eventIcon(row: TrajectoryRow) {
  if (row.status === "failed") return <CloseCircleOutlined />;
  if (row.type.startsWith("approval.")) return <ExclamationCircleOutlined />;
  if (row.type.startsWith("task.")) return <UnorderedListOutlined />;
  if (row.lane === "tool") return <ToolOutlined />;
  if (row.lane === "model") return <RobotOutlined />;
  if (row.type.startsWith("run.")) return <UserOutlined />;
  return <FileTextOutlined />;
}

function rowInputPreview(row: TrajectoryRow) {
  const data = row.data;
  if (row.rowId.startsWith("tool:")) return compactValue(data.arguments);
  if (row.rowId.startsWith("model:")) return compactValue(lastMessage(data.messages));
  if (row.type === "system.prompt") return compactValue(data.content);
  if (row.type === "assistant.message" || row.type.startsWith("run.")) return "";
  return firstText(data.skill, data.message, data.reason, data.content);
}

function rowOutputPreview(row: TrajectoryRow) {
  const data = row.data;
  const error =
    data.error && typeof data.error === "object"
      ? (data.error as { message?: unknown }).message
      : data.error;
  if (row.status === "failed") return firstText(error, data.message, data.reason) || "未完成";
  if (row.rowId.startsWith("tool:")) return compactValue(data.trace ?? data.result);
  if (row.rowId.startsWith("model:")) return compactValue(responseContent(data.response));
  if (row.type === "assistant.message") return compactValue(data.content);
  if (row.type === "run.completed")
    return `${Number(data.modelCalls || 0)} 次模型，${Number(data.toolCalls || 0)} 次工具`;
  if (row.type === "approval.required") return "等待确认";
  return firstText(data.outcome, data.stage, data.status);
}

function responseContent(value: unknown) {
  if (!value || typeof value !== "object") return value;
  const response = value as { content?: unknown; toolCalls?: unknown };
  return response.content || response.toolCalls || value;
}

function lastMessage(value: unknown) {
  if (!Array.isArray(value) || !value.length) return value;
  const message = value[value.length - 1];
  return message && typeof message === "object" && "content" in message
    ? (message as { content?: unknown }).content
    : message;
}

function compactValue(value: unknown) {
  if (value === null || value === undefined || value === "") return "";
  const text = typeof value === "string" ? value : JSON.stringify(value);
  return text.replace(/\s+/g, " ").trim().slice(0, 160);
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

function statusLabel(row: TrajectoryRow) {
  return {
    recorded: "已记录",
    running: "进行中",
    pending: "待审批",
    success: "已完成",
    failed: "失败/取消",
  }[row.status];
}

function firstText(...values: unknown[]) {
  return (
    values
      .find((value): value is string => typeof value === "string" && value.trim() !== "")
      ?.trim() || ""
  );
}

function formatClock(value?: string) {
  return value
    ? new Intl.DateTimeFormat("zh-CN", {
        timeZone: timezone,
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
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

function formatDuration(milliseconds: number) {
  if (milliseconds < 1000) return `${Math.round(milliseconds)} ms`;
  if (milliseconds < 60_000)
    return `${(milliseconds / 1000).toFixed(milliseconds < 10_000 ? 2 : 1)} 秒`;
  const minutes = Math.floor(milliseconds / 60_000);
  return `${minutes} 分 ${Math.round((milliseconds % 60_000) / 1000)} 秒`;
}
