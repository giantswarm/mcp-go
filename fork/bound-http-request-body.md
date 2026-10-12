# Bound the POST body the HTTP transports read

- **Status:** carried
- **Upstream:** none yet; to be sent. mcp-go through v1.2.1 is affected by CVE-2026-108859 and no upstream release fixes it
- **Our commit:** the squash commit of this change on `giantswarm`; the code is the upstream-ready commit `fix(server): bound the POST body the HTTP transports read`

## Change

`StreamableHTTPServer` and `SSEServer` read at most `DefaultMaxRequestBodyBytes` (4 MiB, the TypeScript SDK's message size limit) of a POST body and refuse a larger one with 413 Request Entity Too Large before reading it in full. `WithMaxRequestBodyBytes` and `WithSSEMaxRequestBodyBytes` change the bound; zero or a negative value removes it.

## Reason

CVE-2026-108859: the Streamable HTTP server read a POST body whole with `io.ReadAll` before validating it, and the SSE message endpoint decoded a body without a bound, so one large or many concurrent large bodies could exhaust the server's memory. muster serves both transports to every client of its aggregator.
