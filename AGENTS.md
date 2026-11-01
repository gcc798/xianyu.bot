# Repository guidance

## Frontend ownership

- `agent-web/` is the only actively maintained frontend. All new UI work, fixes, dependencies, tests, and design-system changes belong here.
- `agent-web/` uses React 18 and TypeScript. Do not add JavaScript or JSX source files; use `.ts` and `.tsx`.
- `web/` is the deprecated Gin-Vue-Admin/Vue frontend. Its source is retained temporarily only for behavior comparison, migration verification, and rollback.
- Do not add features to `web/`, fix its UI, refactor it, update its dependencies, regenerate its build output, or treat its architecture as a pattern for new code.
- Reading `web/` to confirm an existing API contract or legacy behavior is allowed. Implement the result in `agent-web/`.
- Removing `web/` is a separate cleanup step after React capability parity is accepted; do not delete it incidentally.

## React migration quality bar

- Preserve backend API behavior while migrating UI capability.
- Keep authentication, authorization boundaries, empty/loading/error states, keyboard access, responsive behavior, and destructive-action confirmation working.
- Validate material UI changes with a production build and a rendered browser check.
