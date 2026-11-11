# Spec Delta

## Purpose

Allows users to initiate delivery-resource imports with Agent assistance while ensuring redeemable card values are entered and processed outside model-visible and traceable Agent content.

## ADDED Requirements

### Requirement: Delivery-card values use a protected input channel
The system SHALL collect delivery-card values through a protected user input flow and SHALL NOT include the values in model prompts, Agent tool arguments, approval events, run traces, SSE events, or ordinary responses.

#### Scenario: Agent initiates a card import
- **WHEN** the user asks the Agent to import delivery cards
- **THEN** the Agent directs the user to a protected input flow without requesting or forwarding the card values in conversation

#### Scenario: User submits card values through the protected flow
- **WHEN** the user confirms card values in the protected flow
- **THEN** the system imports them using the authenticated user's ownership scope without exposing the values to Agent-visible channels

#### Scenario: Import is cancelled or fails
- **WHEN** the user cancels the protected flow or the import fails
- **THEN** no card values are exposed in the Agent conversation or ordinary operation result and the system reports only a safe status

### Requirement: Delivery-resource imports enforce ownership and storage protections
The system SHALL verify that the target delivery rule belongs to the authenticated user and SHALL preserve the existing protected storage behavior for imported card values.

#### Scenario: Import targets another user's rule
- **WHEN** a user submits an import for a delivery rule owned by another user
- **THEN** the import is rejected without storing any submitted values

#### Scenario: Import succeeds
- **WHEN** valid card values are submitted for an owned delivery rule
- **THEN** the values are stored using the existing encrypted resource-storage behavior and are not returned in plaintext
