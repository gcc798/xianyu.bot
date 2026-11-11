# Spec Delta

## Purpose

Allows users to manage supported Xianyu order-automation operations through natural language while preserving owner isolation, approval for changes, and truthful reporting of execution results.

## ADDED Requirements

### Requirement: Agent exposes supported Xianyu business operations
The system SHALL let the Agent query and operate supported Xianyu products, orders, messages, automation rules, delivery tasks, inventory, and business events without exposing account credential management as an Agent write capability.

#### Scenario: User asks to inspect automation status
- **WHEN** the user asks a natural-language question about orders, automation settings, tasks, inventory, or business events
- **THEN** the Agent uses the relevant supported Xianyu capability and answers from its result

#### Scenario: User asks for unsupported platform operation
- **WHEN** the requested operation is not supported by an available business capability
- **THEN** the Agent explains that it cannot perform the operation and does not claim that it did

### Requirement: Xianyu business data is isolated by owner
The system SHALL derive the owner identity from the authenticated Agent run and SHALL reject access to business records not owned by that user.

#### Scenario: User requests another owner's record
- **WHEN** an Agent tool is invoked with an identifier belonging to another user
- **THEN** the operation returns an authorization-safe failure and exposes no record data

#### Scenario: Tool arguments include an owner override
- **WHEN** a tool request attempts to provide or override the owner identity
- **THEN** the system ignores or rejects that override and uses the authenticated Agent scope

### Requirement: State-changing Xianyu operations require approval
The system SHALL obtain explicit user approval before executing an Agent-requested operation that changes Xianyu business configuration, sends a message, synchronizes remote data, retries a task, or changes conversation takeover state.

#### Scenario: User has not approved a write
- **WHEN** the Agent proposes a state-changing Xianyu operation
- **THEN** the operation remains pending and no business side effect occurs before approval

#### Scenario: User rejects a proposed write
- **WHEN** the user rejects the pending operation
- **THEN** the operation is not executed and the Agent reports that it was rejected

#### Scenario: User approves a proposed write
- **WHEN** the user approves the pending operation
- **THEN** the system executes it under the authenticated user's ownership scope and records its outcome

### Requirement: Partial automation updates preserve unspecified settings
The system SHALL change only the automation fields explicitly requested by the user and SHALL preserve all other stored settings.

#### Scenario: User changes one automation option
- **WHEN** the user requests a change to one product automation option without mentioning other options
- **THEN** the requested option is updated while all unspecified options and their values remain unchanged

#### Scenario: Partial update violates a product invariant
- **WHEN** the requested partial update would leave the product automation configuration invalid
- **THEN** the update is rejected with an actionable validation result and the stored configuration remains unchanged

### Requirement: Agent reports operation outcomes from verified results
The system SHALL distinguish pending approval, rejection, successful execution, and failure, and SHALL NOT report a side effect as complete until the business operation succeeds.

#### Scenario: Operation is waiting for approval
- **WHEN** the tool call requires approval
- **THEN** the Agent asks the user to approve and does not claim the operation is complete

#### Scenario: Operation completes or fails
- **WHEN** an approved operation returns a success or failure result
- **THEN** the Agent reports that result and may query the corresponding task or business event to verify the resulting state

### Requirement: Unknown relist outcomes are not retried automatically
The system SHALL allow relist retry only for tasks known to have failed and SHALL prevent retry of tasks whose platform outcome is unknown.

#### Scenario: Failed relist task is retried
- **WHEN** the user approves retry of a relist task in an explicit failed state owned by that user
- **THEN** the system retries that task and records the resulting state

#### Scenario: Relist outcome is unknown
- **WHEN** the user asks to retry a relist task with an unknown outcome
- **THEN** the system does not retry it and explains that the platform outcome must be verified first

### Requirement: User-facing tool catalog identifies Xianyu tools
The system SHALL display the source of built-in Xianyu tools as “闲鱼” while retaining internal source metadata needed for execution governance and audit.

#### Scenario: User views available tools
- **WHEN** the tool catalog includes a built-in Xianyu tool
- **THEN** the catalog labels its source “闲鱼” and MCP tools continue to show their configured service name
