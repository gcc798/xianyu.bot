# Proposal

## Why

Users should be able to inspect and operate existing Xianyu order-automation workflows through natural language instead of translating each request into a console operation. The current Agent exposes only a small set of Xianyu tools, so important business capabilities remain unavailable even where the service and HTTP API already support them.

## What Changes

- Expose owner-scoped Xianyu product, order, message, automation-rule, delivery-task, inventory, and business-event operations to the Agent through domain-oriented tools.
- Require approval before state-changing operations, and make product automation updates preserve settings the user did not ask to change.
- Add missing relist-task visibility and only permit retry of explicitly failed tasks; unknown outcomes must not be retried automatically.
- Keep delivery-card secrets out of model prompts, tool arguments, approval events, traces, and ordinary responses by handing secret entry to a protected UI flow.
- Present Xianyu tools as “闲鱼” in the user-facing tool catalog while retaining internal source metadata for governance and audit.
- Keep account and credential management outside the Agent's writable capabilities. Do not add unsupported Xianyu platform operations as part of this change.

## Capabilities

### New Capabilities
- `xianyu-agent-operations`: Natural-language access to owner-scoped Xianyu operations, with approval for writes and safe handling of partial configuration changes.
- `delivery-resource-import`: Protected delivery-card import handoff that keeps secret values outside Agent-visible data and trace channels.

### Modified Capabilities

None. The repository currently has no published OpenSpec capabilities; these behaviors will be specified as new capabilities.

## Impact

- Agent tools and registration: `server/internal/xianyu/agenttool`, `server/cmd/server/main.go`, and Agent tool catalog/governance code.
- Existing Xianyu services and request models for product, order, message, reply rule, delivery rule, delivery task, business event, and relist behavior.
- New owner-scoped relist task query/retry behavior where required.
- Frontend Agent tool catalog, approval/trace interactions, and a protected delivery-resource import input flow.
- Tests for tool schemas, owner isolation, approval, partial updates, task state restrictions, secret redaction, UI build, and browser workflows.
