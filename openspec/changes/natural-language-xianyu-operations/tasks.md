# Tasks

## 1. Establish Safe Service Contracts

- [x] 1.1 Add a presence-aware product automation patch request and owner-scoped Service update that applies only supplied fields, validates the resulting full setting, and persists atomically; verify omitted bool/string fields remain unchanged and invalid patches roll back in focused Service tests.
- [x] 1.2 Add owner-scoped relist-task listing and failed-only retry behavior; verify pagination, cross-owner rejection, failed retry, and `unknown` rejection in focused Service tests.
- [x] 1.3 Review existing Service methods used by Agent tools for owner checks and safe response fields; add missing ownership tests for product settings, rules, delivery tasks, inventory, business events, and messages, and run the affected Service test packages.

## 2. Expose Xianyu Read Tools

- [x] 2.1 Add Agent tools for business-event list, order detail, delivery-task list/inventory, reply-rule list, delivery-rule list, and relist-task list using existing Service methods; verify schemas, owner scope, pagination/filter mapping, safe outputs, and service-error mapping in `internal/xianyu/agenttool` tests.
- [x] 2.2 Register the read tools with their required Service dependencies and label built-in Xianyu tools “闲鱼” in the user-facing catalog while retaining internal `local` source; verify catalog labels and unchanged persisted source semantics in Agent tests.
- [x] 2.3 Update Agent instructions with the supported Xianyu operating scope and truthful result-reporting rules; verify instruction/tool catalog tests and run affected Agent and Xianyu tool package tests.

## 3. Expose Approval-Gated Business Writes

- [x] 3.1 Add approval-gated tools for product sync, pending-order sync, product automation patch, reply/delivery rule save/delete, failed delivery retry, failed relist retry, message takeover update, and retain approval on message reply; verify every write queues approval without invoking its Service before approval.
- [x] 3.2 Register write-tool dependencies and verify approval, rejection, approved execution, budget accounting, owner isolation, and audit outcomes with focused Agent integration tests.
- [x] 3.3 Add or update docs for the Agent's Xianyu operation boundary and approval behavior; verify documentation matches registered tools and the account/credential exclusion.

## 4. Reuse the Protected Delivery-Card Import Flow

- [x] 4.1 Add a non-secret Agent handoff to the existing delivery-rule import UI, including an optional owned rule identifier; verify the link carries no card values and opens the existing import modal for the selected rule.
- [x] 4.2 Keep card submission on the authenticated direct UI-to-API path and prevent card values from entering chat state, Agent arguments, approval events, SSE, logs, or analytics; verify by code-path tests and a browser inspection of the network payload/Agent trace.
- [x] 4.3 Verify cross-owner rule IDs are rejected, successful imports return count/status only, and existing encrypted storage behavior remains intact; run the focused API/Service tests and `npm run build`.

## 5. Verify Natural-Language End-to-End Workflows

- [x] 5.1 Add controlled-model and stubbed-Xianyu end-to-end tests for read-only status inspection, partial product automation update, rule creation, approved delivery retry, event/task result verification, and rejected approval; assert persisted state and audit events rather than assistant text alone.
- [x] 5.2 Add a relist workflow test proving explicit failure may be retried but unknown outcome is never retried or reported as success; verify no duplicate publish call occurs for `unknown`.
- [x] 5.3 Run all affected Go tests and frontend build, then perform browser checks for catalog source labels, approval/rejection, post-approval result lookup, and protected card-import handoff; record that stubbed tests do not constitute real-account or real-transaction acceptance.
