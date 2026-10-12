# Stop dropping notifications when no event store is configured

- **Status:** carried
- **Upstream:** [mark3labs/mcp-go#969](https://github.com/mark3labs/mcp-go/pull/969), closed unmerged on 2026-10-07; to be re-sent
- **Our commit:** `62d6b0c` (giantswarm/mcp-go#2)

## Change

`fix(server): stop dropping notifications when no event store is configured`

## Reason

The streamable-http server's forwarder takes a notification off the session channel, then waits for the response lock; when the response path takes the lock first, drains the (empty) channel and closes `done`, the forwarder returns and the notification is lost. A tool that reports progress and returns at once loses its last `notifications/progress`, at a backend and again at muster's relay hop.
