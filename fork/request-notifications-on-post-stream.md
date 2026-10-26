# Send a request's notifications on its POST stream

- **Status:** carried
- **Upstream:** [mark3labs/mcp-go#1115](https://github.com/mark3labs/mcp-go/pull/1115), open (draft), from `upstream/request-notifications-on-post-stream`
- **Our commit:** `7391a80` (giantswarm/mcp-go#3)

## Change

`fix(server): send a request's notifications on its POST stream`

## Reason

A notification a handler sends during a POST goes to the session's shared channel, which the standalone GET stream's forwarder also reads; with a GET stream open (muster's downstream clients keep one), a call's progress can arrive on the GET stream after the call's response, and muster's per-call progress router drops it. Routed like request-scoped server requests (upstream #933).
