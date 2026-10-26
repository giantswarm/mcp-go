# Send a request's notifications on its POST stream

- **Status:** carried
- **Upstream:** [mark3labs/mcp-go#1117](https://github.com/mark3labs/mcp-go/pull/1117), merged 2026-10-26 as `4944283`, not yet released, fixing [mark3labs/mcp-go#1116](https://github.com/mark3labs/mcp-go/issues/1116) (supersedes [mark3labs/mcp-go#1115](https://github.com/mark3labs/mcp-go/pull/1115), closed for lacking a linked issue)
- **Our commit:** `7391a80` (giantswarm/mcp-go#3), brought to the upstream state by `a19544b` (giantswarm/mcp-go#17)

## Change

`fix(server): send a request's notifications on its POST stream`

## Reason

A notification a handler sends during a POST goes to the session's shared channel, which the standalone GET stream's forwarder also reads; with a GET stream open (muster's downstream clients keep one), a call's progress can arrive on the GET stream after the call's response, and muster's per-call progress router drops it. Routed like request-scoped server requests (upstream #933). Only notifications addressed to the session the POST serves take this route, sends are ordered against the stream closing so none is queued after the last drain, and a POST carrying a client notification drains what its handler queued before answering.
