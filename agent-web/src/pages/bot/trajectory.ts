import type { AgentEvent } from "../../models/agent";

export type TrajectoryLane = "input" | "model" | "tool";
export type TrajectoryStatus = "recorded" | "running" | "pending" | "success" | "failed";

export type TrajectoryRow = AgentEvent & {
  rowId: string;
  lane: TrajectoryLane;
  status: TrajectoryStatus;
  startedAt?: string;
  completedAt?: string;
  durationMillis?: number;
};

const hiddenBoundaries = new Set([
  "turn.started",
  "turn.completed",
  "turn.failed",
  "step.started",
  "step.completed",
  "step.failed",
]);

export function buildTrajectoryRows(events: AgentEvent[]): TrajectoryRow[] {
  const rows: TrajectoryRow[] = [];
  const calls = new Map<string, TrajectoryRow>();

  for (const event of [...events].sort(compareEvents)) {
    if (hiddenBoundaries.has(event.type)) continue;
    const key = trajectoryCallKey(event);
    if (!key) {
      rows.push(toRow(event, `event:${event.id}`));
      continue;
    }
    const current = calls.get(key);
    if (!current) {
      const row = toRow(event, key);
      calls.set(key, row);
      rows.push(row);
      continue;
    }
    current.id = event.id;
    current.type = event.type;
    current.data = { ...current.data, ...event.data };
    current.status = trajectoryStatus(event.type);
    current.callId ||= event.callId || stringValue(event.data.callId);
    current.runId ||= event.runId || numberValue(event.data.runId);
    current.turnNo ||= event.turnNo || numberValue(event.data.turnNo);
    current.stepNo ||= event.stepNo || numberValue(event.data.stepNo);
    if (isTerminal(event.type)) current.completedAt = event.createdAt;
    current.durationMillis = eventDuration(current);
  }

  return rows.sort(
    (left, right) => Date.parse(left.startedAt || "") - Date.parse(right.startedAt || ""),
  );
}

export function trajectoryLane(type: string): TrajectoryLane {
  if (type.startsWith("tool.") || type.startsWith("approval.") || type.startsWith("task."))
    return "tool";
  if (
    type.startsWith("model.") ||
    type === "request.snapshot" ||
    type === "system.prompt" ||
    type === "assistant.message"
  )
    return "model";
  return "input";
}

export function trajectoryStatus(type: string): TrajectoryStatus {
  if (type.includes("failed") || type.includes("cancelled") || type === "approval.rejected")
    return "failed";
  if (type === "approval.required") return "pending";
  if (type.endsWith(".completed")) return "success";
  if (type.endsWith(".started")) return "running";
  return "recorded";
}

function toRow(event: AgentEvent, rowId: string): TrajectoryRow {
  const completedAt = isTerminal(event.type) ? event.createdAt : undefined;
  const row: TrajectoryRow = {
    ...event,
    callId: event.callId || stringValue(event.data.callId) || undefined,
    runId: event.runId || numberValue(event.data.runId) || undefined,
    turnNo: event.turnNo || numberValue(event.data.turnNo) || undefined,
    stepNo: event.stepNo || numberValue(event.data.stepNo) || undefined,
    data: { ...event.data },
    rowId,
    lane: trajectoryLane(event.type),
    status: trajectoryStatus(event.type),
    startedAt: event.createdAt,
    completedAt,
  };
  row.durationMillis = eventDuration(row);
  return row;
}

function trajectoryCallKey(event: AgentEvent) {
  const runId = event.runId || numberValue(event.data.runId);
  const callId = event.callId || stringValue(event.data.callId);
  if (runId && callId && event.type.startsWith("tool.")) return `tool:${runId}:${callId}`;
  const stepNo = event.stepNo || numberValue(event.data.stepNo);
  if (
    runId &&
    stepNo &&
    (event.type === "request.snapshot" || event.type.startsWith("model.attempt."))
  ) {
    return `model:${runId}:${stepNo}:${numberValue(event.data.attemptNo) || 1}`;
  }
  return "";
}

function eventDuration(row: TrajectoryRow) {
  const recorded = numberValue(row.data.durationMillis);
  if (recorded > 0) return recorded;
  const started = Date.parse(row.startedAt || "");
  const completed = Date.parse(row.completedAt || "");
  return Number.isFinite(started) && Number.isFinite(completed) && completed > started
    ? Math.max(0, completed - started)
    : undefined;
}

function compareEvents(left: AgentEvent, right: AgentEvent) {
  const byTime = Date.parse(left.createdAt || "") - Date.parse(right.createdAt || "");
  return byTime || left.id - right.id;
}

function isTerminal(type: string) {
  return type.endsWith(".completed") || type.endsWith(".failed") || type.endsWith(".cancelled");
}

function numberValue(value: unknown) {
  const number = Number(value || 0);
  return Number.isFinite(number) ? number : 0;
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value : "";
}
