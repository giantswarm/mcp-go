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
| `giantswarm` (default) | **The line**: the upstream release tag muster runs ("the pin") + the carried patches + this fork's own files (`FORK.md`, `renovate.json5`, and upstream's `release.yml` and `pages.yml` workflows removed: the line's auto-release creates its releases, and the line publishes no site). Every change is a pull request against it. |
| `main` | upstream `main` at the time of the fork; not consumed, not synced. |
| `fork/<topic>` | pull-request branches against `giantswarm` |

## Pin

Upstream **v1.2.0** (`4cb0d201`, 2026-10-08). The line moves to the next upstream release muster takes, with every
carried patch kept in effect on it or dropped once upstream contains it.

## Moving the pin

A pull request from `fork/v<version>` against `giantswarm` merges the upstream tag into the line
(`git merge --no-ff v<version>`), resolving conflicts so both carried patches stay in effect on the new upstream code
and the fork's own files stay as the line has them; upstream's server tests and `go vet ./...` pass before it opens.
It is squash-merged like every other change, so the line's history keeps every earlier commit and the protected
branch is never force-pushed or rebased. Since the squash commit has one parent, the new upstream tag is not an
ancestor of the line: `git describe` names the previous upstream tag on the line plus a distance, and so does the
merged head's pseudo-version (`v1.1.2-0.…` while `v1.1.1` is the newest ancestor tag), which consumers pin in the
`replace`; `go list -m github.com/giantswarm/mcp-go@<commit>` prints it.

## Consumers

muster takes the line through a `replace` directive, keeping upstream's module path and release in its `require`,
pinned to a commit of `giantswarm` by its pseudo-version:

```
replace github.com/mark3labs/mcp-go v1.2.0 => github.com/giantswarm/mcp-go v1.1.2-0.<date>-<commit>
```

The module path in `go.mod` stays `github.com/mark3labs/mcp-go`, so the line is only usable through `replace`.
The `replace` goes away, and this line retires, at the first upstream release that contains every carried patch.

## Releases

None. A Go module needs no build or publish step, and a pseudo-version names exactly the commit of the protected
`giantswarm` branch a consumer runs, so the line cuts no tags of its own (an own tag would sit in upstream's `v1`
version space, since a module's major is part of its import path). Upstream tags on the line are upstream's, kept for
`git describe`.

## Carried patches

| Patch | Why muster needs it | Upstream | State |
|---|---|---|---|
| `fix(server): stop dropping notifications when no event store is configured` | The streamable-http server's forwarder takes a notification off the session channel, then waits for the response lock; when the response path takes the lock first, drains the (empty) channel and closes `done`, the forwarder returns and the notification is lost. A tool that reports progress and returns at once loses its last `notifications/progress`, at a backend and again at muster's relay hop. | [mark3labs/mcp-go#969](https://github.com/mark3labs/mcp-go/pull/969) closed unmerged on 2026-10-07; to be re-sent, queued in giantswarm/giantswarm#37742 | carried |
| `fix(server): send a request's notifications on its POST stream` | A notification a handler sends during a POST goes to the session's shared channel, which the standalone GET stream's forwarder also reads; with a GET stream open (muster's downstream clients keep one), a call's progress can arrive on the GET stream after the call's response, and muster's per-call progress router drops it. Routed like request-scoped server requests (upstream #933). | none yet; merged into the line as giantswarm/mcp-go#3, queued in giantswarm/giantswarm#37742 | carried |
