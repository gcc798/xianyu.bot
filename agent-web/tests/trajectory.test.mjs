import assert from "node:assert/strict";
import test from "node:test";
import { buildTrajectoryRows } from "../src/pages/bot/trajectory.ts";

test("merges model and tool lifecycles without losing their start time", () => {
  const rows = buildTrajectoryRows([
    event(4, "tool.completed", "2026-11-13T02:00:04Z", {
      runId: 7,
      callId: "call-1",
      durationMillis: 3000,
    }),
    event(1, "request.snapshot", "2026-11-13T02:00:00Z", {
      runId: 7,
      stepNo: 1,
      messages: ["hello"],
    }),
    event(3, "tool.started", "2026-11-13T02:00:01Z", {
      runId: 7,
      callId: "call-1",
      arguments: { query: "book" },
    }),
    event(2, "model.attempt.completed", "2026-11-13T02:00:01Z", {
      runId: 7,
      stepNo: 1,
      response: "call tool",
    }),
  ]);

  assert.equal(rows.length, 2);
  assert.deepEqual(
    rows.map(({ lane, startedAt, completedAt, durationMillis, status }) => ({
      lane,
      startedAt,
      completedAt,
      durationMillis,
      status,
    })),
    [
      {
        lane: "model",
        startedAt: "2026-11-13T02:00:00Z",
        completedAt: "2026-11-13T02:00:01Z",
        durationMillis: 1000,
        status: "success",
      },
      {
        lane: "tool",
        startedAt: "2026-11-13T02:00:01Z",
        completedAt: "2026-11-13T02:00:04Z",
        durationMillis: 3000,
        status: "success",
      },
    ],
  );
  assert.deepEqual(rows[0].data.messages, ["hello"]);
  assert.equal(rows[0].data.response, "call tool");
});

test("does not invent a zero duration for a single terminal event", () => {
  const [row] = buildTrajectoryRows([
    event(1, "run.completed", "2026-11-13T02:00:04Z", { runId: 7 }),
  ]);

  assert.equal(row.durationMillis, undefined);
});

function event(id, type, createdAt, data) {
  return { id, type, createdAt, data };
}
