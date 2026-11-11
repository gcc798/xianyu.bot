# Design

## Context

The Agent currently registers five Xianyu tools in `server/internal/xianyu/agenttool`; those tools call services with the owner ID supplied by the Agent run scope. Many related services already support product/order/message operations, rule CRUD, delivery task retry, inventory, and business-event queries. The Agent's governed tool adapter persists tool arguments and routes definitions marked `RequiresApproval` into the existing approval flow. This makes ordinary business writes straightforward to expose, but makes plaintext delivery cards unsafe as Agent tool arguments.

## Goals / Non-Goals

**Goals:**
- Expose supported existing Xianyu business behavior through intent-oriented Agent tools.
- Keep business authorization and validation in Service methods, scoped by the Agent run owner.
- Preserve unspecified product automation settings during natural-language partial updates.
- Keep write operations behind the existing approval mechanism and make outcomes queryable.
- Give delivery-card values a protected UI-to-Service path that bypasses model and Agent trace content.

**Non-Goals:**
- Agent management of Xianyu accounts, QR login, cookies, tokens, or credentials.
- New Xianyu platform protocol operations not already supported by current services.
- Direct user-facing MCP exposure for first-party Xianyu operations; local tools remain the in-process adapter.
- Agent receipt of delivery-card plaintext in prompts, normal chat messages, or tool arguments.

## Decisions

### Register domain-oriented local tools over existing services

Keep the current `xianyutool.Register` pattern and inject only the service dependencies each tool needs. Add query tools for business events, order detail, delivery tasks/inventory, and reply/delivery rules; add write tools for product/order sync, product automation configuration, rule save/delete, delivery retry, message takeover, and existing message reply. Tool codes stay stable and Xianyu-specific. Each invocation derives `ownerID` from `agenttool.Scope`; schemas do not accept owner identity. No tool accesses GORM or calls HTTP handlers.

For settings, rules, and task operations, service methods remain the authority for ownership and domain invariants. Add focused Service methods only where no suitable method exists, notably relist-task listing and failed-only retry. Do not add wrappers around methods that already express the required business action.

### Treat product automation configuration as a patch

The existing `UpdateProductSetting` request is a full replacement with non-pointer booleans. Do not pass partially populated natural-language input into it: omitted booleans would become false and omitted strings could clear saved values. Introduce a presence-aware patch contract for the Agent path, where each mutable field is optional. In the Product Service, load the current owner-scoped setting, apply only present fields, validate the resulting complete configuration using existing invariants, and persist atomically. The console's existing full-update contract can remain unchanged unless implementation discovers it can safely share the same service logic without API compatibility changes.

This patch behavior is deliberately in Service, not assembled by the Tool, so all callers share correct validation and updates do not race through a read-then-write split in the adapter.

### Reuse approval for all business writes

Mark every state-changing Xianyu tool `RequiresApproval`, including remote synchronization, message send/reply, takeover changes, rule/configuration writes, delivery retry, and relist retry. The existing governed adapter records the proposed arguments and queues approval before invoking the Service. Read-only tools remain non-approval tools. Model-facing instructions must say that pending approval is not completion and that the Agent must report the confirmed Service result.

Relist retry is limited to records in explicit `failed` state. `unknown` is not retryable because the remote publish may have succeeded even if local confirmation failed. Querying events and task state is the resolution path; no automatic retry or deduplication claim is added without platform evidence.

### Separate catalog display source from internal execution source

Keep `Definition.Source = "local"` for persisted audit and execution semantics. Map built-in Xianyu tool catalog entries to display source “闲鱼”; continue using the configured MCP server name for remote tools. This avoids changing stored trace meanings or the source discriminator consumed by governance.

### Keep card values outside Agent execution records

Do not register `Delivery.ImportResources` as a normal Agent tool. The governed adapter persists arguments and publishes them in approval lifecycle events, so passing card values through it cannot satisfy the secrecy requirement. The existing console already has a card import modal whose values stay in page-local state and are posted directly to the authenticated import API. Reuse that flow: Agent can return a console link to the delivery rules page, optionally identifying an owned delivery rule so the page can open the existing modal. The card values continue directly from the protected form to the existing owner-checking, encrypting Service; only a count/status may be returned. Do not copy the form payload into chat state, Agent events, logs, or analytics.

Product synchronization is included in the initial natural-language operations scope because the existing product Service supports it and it is needed for requests such as “sync my latest listings.” It is a write tool and requires approval.
### Add relist task read and retry service behavior before its tools

`Relist` currently plans and processes tasks in the background but has no owner-scoped list or user-triggered retry API. Add a paginated owner-scoped query that returns task state and safe business identifiers, and a retry operation guarded by owner and `failed` status. Do not expose internal publish snapshots or credentials. Any Agent retry is approval-gated and uses the same processing path as the scheduler after claiming the task atomically.

## Risks / Trade-offs

- [Partial updates can overwrite concurrent console edits if implemented as separate read and write steps] -> Apply patch and validate/persist under one Service transaction or equivalent conditional update.
- [Approvals can expose sensitive arguments because they are persisted and published] -> Do not put secrets in Agent tool schemas; specifically keep card values on the protected direct UI-to-Service path.
- [Retrying an externally ambiguous relist could create duplicate listings] -> Permit retry only from explicit `failed` state; reject `unknown` and require verification.
- [Many tools can overwhelm model selection and increase accidental writes] -> Use narrow descriptions, clear read/write separation, existing tool budgets, and approval for every write; avoid generic CRUD tools.
- [Tool outputs may contain private message or buyer data] -> Keep outputs owner-scoped, bounded by existing pagination, and limited to fields needed for the user's requested operation.

## Migration Plan

1. Add and test service contracts and Agent tool schemas without changing existing HTTP routes or stored trace source values.
2. Register tools and update user-facing catalog mapping; keep existing tool codes compatible.
3. Add protected import handoff using the existing import API/service where possible; never route card values through Agent execution.
4. Deploy backend before frontend handoff controls. If the handoff UI is unavailable, Agent card-import initiation remains unavailable while existing manual import continues to work.
5. Roll back by disabling/removing newly registered tools and handoff UI; existing console/API operations and persisted records remain valid.
