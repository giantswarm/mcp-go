# The Giant Swarm line of mcp-go

This repository is Team Bumblebee's fork of [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go), the MCP Go SDK
[`giantswarm/muster`](https://github.com/giantswarm/muster) builds on. It carries upstream fixes muster needs before
upstream releases them, and answers one question from one place: **which mcp-go does muster run, and why does it
differ from upstream?**

Tracking (this fork has issues disabled): [giantswarm/muster#1405](https://github.com/giantswarm/muster/issues/1405).
A carried patch is one upstream-ready commit on this fork with its ledger file in `fork/`, pushed as an
`upstream/<slug>` branch on this repository and queued as a row on
[giantswarm/giantswarm#37742](https://github.com/giantswarm/giantswarm/issues/37742). A human decides what goes upstream to mark3labs/mcp-go and opens it; an agent
never opens, comments on, reviews or closes a pull request or issue on the upstream repository. Once a human has
opened one, its ledger file names that pull request.

## Branches

| Branch | What it is |
|---|---|
| `giantswarm` (default) | **The line**: the upstream release tag muster runs ("the pin") + the carried patches + this fork's own files (`FORK.md`, `renovate.json5`, and upstream's `release.yml` and `pages.yml` workflows removed: the line's auto-release creates its releases, and the line publishes no site). Every change is a pull request against it. |
| `main` | upstream `main` at the time of the fork; not consumed, not synced. |
| `fork/<topic>` | pull-request branches against `giantswarm` |

## Pin

Upstream **v1.3.0** (`e1bd0abd`, 2026-10-14). The line moves to the next upstream release muster takes, with every
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
replace github.com/mark3labs/mcp-go v1.3.0 => github.com/giantswarm/mcp-go v1.1.2-0.<date>-<commit>
```

The module path in `go.mod` stays `github.com/mark3labs/mcp-go`, so the line is only usable through `replace`.
The `replace` goes away, and this line retires, at the first upstream release that contains every carried patch.

## Releases

None. A Go module needs no build or publish step, and a pseudo-version names exactly the commit of the protected
`giantswarm` branch a consumer runs, so the line cuts no tags of its own (an own tag would sit in upstream's `v1`
version space, since a module's major is part of its import path). Upstream tags on the line are upstream's, kept for
`git describe`.

## Carried patches

Each change the line carries has its own ledger file in [`fork/`](fork/); the directory listing is the set of carried
changes, and a file's name is the change's stable key. A change adds or edits only its own file, never this one. A
ledger file has a title naming the change and:

- **Status:** `carried`, or `dropped` once a re-pin onto an upstream release that contains it makes it redundant
- **Upstream:** the upstream pull request or issue, and its state
- **Our commit:** the commit on `giantswarm` and the pull request that merged it
- `## Change`: what the change does
- `## Reason`: why the line carries it
