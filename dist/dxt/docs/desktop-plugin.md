+++
title = 'Quilt Desktop Plugin'
weight = 5
+++
# Quilt Desktop Plugin

The recommended way to use Quilt is the **Claude Desktop Extension** (`.dxt`) — a
one-click install that bundles the `quilt` binary, its documentation, and the
live-view templates. Once installed, Claude can read, validate, edit, and render
your schedules through natural language.

## Install

1. Download `quilt.dxt` from the
   [releases page](https://github.com/bmayfi3ld/quilt/releases).
2. Open Claude Desktop → Settings → Extensions.
3. Drag `quilt.dxt` in (or use **Install from file**).

{{< hint warning >}}
**Unsigned build.** The current `.dxt` is not yet code-signed. On macOS,
Gatekeeper may block it on first launch — right-click the app and choose
**Open**, or allow it under System Settings → Privacy & Security. Windows
SmartScreen behaves similarly. Linux is unaffected. Notarization is a planned
follow-up.
{{< /hint >}}

## What it exposes

The extension registers the **`quilt`** MCP server, which is *self-documenting*:

- **Instructions** — an always-on operating manual (the `.db` model and the
  render-not-compute rule) sent to Claude on connect.
- **Tools** — every schedule operation, including `board` (the authoritative
  render payload), `import`, `assign`, and `validate`. See
  [MCP Server]({{< ref "mcp-server" >}}).
- **Resources** — `guide://db-model`, `guide://migrate-xlsx`,
  `guide://live-views`, `guide://tools`, plus `view://master` and `view://class`
  HTML skeletons.
- **Prompts** — `live_view`, `migrate_xlsx`, and `new_schedule` workflows.

## Under the hood

The extension simply runs `quilt mcp` (stdio transport). You can run the same
server yourself from a built binary:

```bash
quilt mcp            # stdio (default)
quilt mcp --http :8080   # opt-in Streamable HTTP
```

Each schedule is one self-contained `.db` file — see
[The .db model]({{< ref "db-model" >}}).
