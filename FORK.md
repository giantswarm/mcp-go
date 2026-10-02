# The Giant Swarm line of mcp-go

This repository is Team Bumblebee's fork of [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go), the MCP Go SDK
[`giantswarm/muster`](https://github.com/giantswarm/muster) builds on. It carries upstream fixes muster needs before
upstream releases them, and answers one question from one place: **which mcp-go does muster run, and why does it
differ from upstream?**

Tracking (this fork has issues disabled): [giantswarm/muster#1405](https://github.com/giantswarm/muster/issues/1405)
and the upstream review queue [giantswarm/giantswarm#37742](https://github.com/giantswarm/giantswarm/issues/37742).

## Branches

| Branch | What it is |
|---|---|
| `giantswarm` (default) | **The line**: the upstream release tag muster runs ("the pin") + the carried patches + this fork's own files (`FORK.md`, `renovate.json5`). Every change is a pull request against it. |
| `main` | upstream `main` at the time of the fork; not consumed, not synced. |
| `fork/<topic>` | pull-request branches against `giantswarm` |

## Pin

Upstream **v1.1.1** (`d74db509`, 2026-09-23). The line moves to the next upstream release muster takes, with every
carried patch rebased onto it or dropped once upstream contains it.

## Consumers

muster takes the line through a `replace` directive, keeping upstream's module path in its `require`:

```
replace github.com/mark3labs/mcp-go => github.com/giantswarm/mcp-go v1.1.2
```

The module path in `go.mod` stays `github.com/mark3labs/mcp-go`, so the line is only usable through `replace`.
The `replace` goes away, and this line retires, at the first upstream release that contains every carried patch.

## Releases

The auto-release workflow tags every conventional merge to `giantswarm`, counted from the highest stable tag (the
pin's upstream tag, then the line's own). The tag is the release: a Go module needs no build or publish step.
Versions stay on upstream's major (`v1`), since a Go module's major is part of its import path.

## Carried patches

| Patch | Why muster needs it | Upstream | State |
|---|---|---|---|
| `fix(server): stop dropping notifications when no event store is configured` | The streamable-http server's forwarder takes a notification off the session channel, then waits for the response lock; when the response path takes the lock first, drains the (empty) channel and closes `done`, the forwarder returns and the notification is lost. A tool that reports progress and returns at once loses its last `notifications/progress`, at a backend and again at muster's relay hop. | [mark3labs/mcp-go#969](https://github.com/mark3labs/mcp-go/pull/969) (open), cherry-picked unchanged | carried |
