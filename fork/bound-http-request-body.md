# Bound the POST body the HTTP transports read

- **Status:** carried
- **Upstream:** [mark3labs/mcp-go#1119](https://github.com/mark3labs/mcp-go/pull/1119), merged 2026-10-27 as `971d157`, not yet released, fixing [#1118](https://github.com/mark3labs/mcp-go/issues/1118). It was stacked on [mark3labs/mcp-go#1092](https://github.com/mark3labs/mcp-go/pull/1092) (third-party, tracked by [#1107](https://github.com/mark3labs/mcp-go/issues/1107), still open), so its squash carries both bounds: `WithMaxRequestBodyBytes` for the Streamable HTTP server and `WithSSEMaxRequestBodyBytes` for the SSE message endpoint, with the same `DefaultMaxRequestBodyBytes`. Upstream also refuses a body whose `Content-Length` is over the bound unread, which this line does not; the line takes upstream's version when the patch drops. mcp-go through v1.4.0 is affected by CVE-2026-108859 ([GHSA-c673-ppxx-9wvj](https://github.com/advisories/GHSA-c673-ppxx-9wvj))
- **Our commit:** the squash commit of this change on `giantswarm`; the code is the upstream-ready commit `fix(server): bound the POST body the HTTP transports read`

## Change

`StreamableHTTPServer` and `SSEServer` read at most `DefaultMaxRequestBodyBytes` (4 MiB, the TypeScript SDK's message size limit) of a POST body and refuse a larger one with 413 Request Entity Too Large before reading it in full. `WithMaxRequestBodyBytes` and `WithSSEMaxRequestBodyBytes` change the bound; zero or a negative value removes it.

## Reason

CVE-2026-108859: the Streamable HTTP server read a POST body whole with `io.ReadAll` before validating it, and the SSE message endpoint decoded a body without a bound, so one large or many concurrent large bodies could exhaust the server's memory. muster serves both transports to every client of its aggregator.
