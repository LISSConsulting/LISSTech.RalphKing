# Feature Specification: Webhook Notifications

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-28
**Status**: Proposed

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Gets Notified on Iteration Complete (Priority: P1)

A developer leaves `ralph run` running unattended and wants to know when each
iteration finishes without watching the terminal. They configure a ntfy.sh topic
URL in `ralph.toml`. When each iteration completes, they receive a push
notification on their phone with the iteration result and cost.

**Why this priority**: Iteration-complete notifications are the highest-value
event — the developer can check progress without returning to the terminal.

**Independent Test**: Set `notifications.url` to an HTTP test server; run the
loop for one iteration; verify the server received a POST with the iteration
message as the body.

**Acceptance Scenarios**:

1. **Given** `notifications.url` is set and `notifications.on_complete = true`,
   **When** an iteration completes, **Then** a POST is sent to the URL with the
   iteration message as the body.
2. **Given** `notifications.on_complete = false`,
   **When** an iteration completes, **Then** no POST is sent.
3. **Given** `notifications.url` is empty,
   **When** any event occurs, **Then** no HTTP request is made.

---

### User Story 2 - Developer Gets Notified on Error (Priority: P2)

A developer's loop encounters an error while they are away. They receive a push
notification alerting them to the failure so they can take action.

**Why this priority**: Error notifications prevent long unattended failures from
going unnoticed until the developer manually checks.

**Independent Test**: Trigger a `LogError` event; verify a POST is sent to the
URL when `on_error = true` and no POST when `on_error = false`.

**Acceptance Scenarios**:

1. **Given** `notifications.url` is set and `notifications.on_error = true`,
   **When** the loop emits a `LogError` event, **Then** a POST is sent.
2. **Given** `notifications.on_error = false`,
   **When** a `LogError` event occurs, **Then** no POST is sent.

---

### User Story 3 - Developer Gets Notified When Loop Stops (Priority: P3)

A developer wants to know when the loop finishes naturally or is stopped
gracefully, so they know the session is complete.

**Why this priority**: Stop notifications close the feedback loop — the
developer knows the session is over without actively watching.

**Independent Test**: Trigger `LogDone` and `LogStopped` events; verify POSTs
are sent when `on_stop = true`.

**Acceptance Scenarios**:

1. **Given** `notifications.url` is set and `notifications.on_stop = true`,
   **When** the loop emits `LogDone`, **Then** a POST is sent.
2. **Given** `notifications.url` is set and `notifications.on_stop = true`,
   **When** the loop emits `LogStopped`, **Then** a POST is sent.
3. **Given** `notifications.on_stop = false`,
   **When** `LogDone` or `LogStopped` occurs, **Then** no POST is sent.

---

### User Story 4 - ntfy.sh-Compatible Format (Priority: P4)

A developer uses ntfy.sh for notifications. The POST body is the message text
and the `X-Title` header is set to the project name for clear attribution.

**Why this priority**: ntfy.sh is the primary use case named in the issue. The
plain-text body + `X-Title` header format is natively supported by ntfy.sh.

**Independent Test**: Inspect the HTTP request headers; verify `Content-Type:
text/plain` and `X-Title: <project.name>` are set.

**Acceptance Scenarios**:

1. **Given** `project.name = "myapp"`, **When** a notification is sent, **Then**
   the `X-Title` header is `"myapp"`.
2. **Given** `project.name` is empty, **When** a notification is sent, **Then**
   the `X-Title` header is `"RalphKing"` (fallback).
3. **Given** any event triggers a notification, **When** the POST is sent,
   **Then** `Content-Type: text/plain` is set.

---

### Edge Cases

- What if the webhook server is unreachable? Notifications are fire-and-forget;
  failures are silently discarded to never interrupt the loop.
- What if the URL is malformed? `config.Validate()` rejects malformed URLs.
- What if the URL is empty? No HTTP requests are made (feature disabled).
- What if `on_complete`, `on_error`, and `on_stop` are all false? No
  notifications are sent regardless of URL.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `ralph.toml` MUST support a `[notifications]` section with
  `url` (string), `on_complete` (bool), `on_error` (bool), `on_stop` (bool).
- **FR-002**: When `notifications.url` is empty, no HTTP requests are made.
- **FR-003**: When `notifications.on_complete = true` and a URL is set, a POST
  MUST be sent on each `LogIterComplete` event.
- **FR-004**: When `notifications.on_error = true` and a URL is set, a POST
  MUST be sent on each `LogError` event.
- **FR-005**: When `notifications.on_stop = true` and a URL is set, a POST MUST
  be sent on each `LogDone` or `LogStopped` event.
- **FR-006**: Default values for `on_complete`, `on_error`, and `on_stop` MUST
  be `true`.
- **FR-007**: Notifications MUST be fire-and-forget (non-blocking); the loop
  MUST NOT wait for the HTTP response.
- **FR-008**: The POST body MUST be the event message as plain text
  (`Content-Type: text/plain`).
- **FR-009**: The `X-Title` header MUST be set to `project.name` when non-empty,
  or `"RalphKing"` otherwise.
- **FR-010**: Notification failures (connection errors, non-2xx responses) MUST
  be silently discarded and MUST NOT affect loop execution.
- **FR-011**: `config.Validate()` MUST return an error when `notifications.url`
  is non-empty but not a valid HTTP/HTTPS URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Loading `ralph.toml` with `[notifications]` sets all four fields.
- **SC-002**: `config.Defaults()` sets `on_complete`, `on_error`, `on_stop` to
  `true` and `url` to `""`.
- **SC-003**: `config.Validate()` rejects `notifications.url = "not-a-url"`.
- **SC-004**: `config.Validate()` accepts `notifications.url = ""`.
- **SC-005**: When `url` is set and `on_complete = true`, a `LogIterComplete`
  event causes a POST to the URL.
- **SC-006**: When `on_complete = false`, a `LogIterComplete` event causes no POST.
- **SC-007**: When `url` is set and `on_error = true`, a `LogError` event causes
  a POST.
- **SC-008**: When `url` is set and `on_stop = true`, `LogDone` and `LogStopped`
  events each cause a POST.
- **SC-009**: POST requests carry `Content-Type: text/plain` and `X-Title`
  headers.
- **SC-010**: A POST failure does not panic, return an error, or stop the loop.
- **SC-011**: All existing tests pass unchanged.
