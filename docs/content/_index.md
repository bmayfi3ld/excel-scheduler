+++
title = 'Introduction'
+++
# Quilt

<img src="/schedule-manager-logo.svg" alt="Quilt logo" width="120" height="120">

**Quilt** helps you build basic school (or any group-based) schedules and checks
them against a configurable list of rules so you always know whether a schedule
is valid. Each schedule is **one self-contained `.db` file** — no servers, no
shared state — and you work with it through Claude using the
[Quilt Desktop Plugin]({{< ref "desktop-plugin" >}}) and its MCP tools.

## Getting Started

1. Install the [Quilt Desktop Plugin]({{< ref "desktop-plugin" >}}) — the
   one-click `.dxt` install for Claude Desktop.
2. Learn [the `.db` model]({{< ref "db-model" >}}): how a schedule is stored and
   edited.
3. Already have an Excel workbook? [Migrate it]({{< ref "migrate-xlsx" >}}) with
   `quilt import`.
4. Render it: build [Live Views]({{< ref "live-views" >}}) on the `board` tool.

For the rule reference, see [Scheduler Rules]({{< ref "rules" >}}). The full tool
surface is in the [MCP Server]({{< ref "mcp-server" >}}) reference.

{{< hint warning >}}
**Legacy frontend.** Quilt began as an Excel add-in. That add-in still works and
its [install instructions]({{< ref "install" >}}) remain, but it is deprecated in
favor of the Desktop plugin and the `.db`/MCP workflow.
{{< /hint >}}

## Questions?

Issues, comments, or requests can be sent to the GitHub issue tracker:

https://github.com/bmayfi3ld/quilt/issues
