+++
title = 'MCP Server'
weight = 50
+++
# MCP Server

The `quilt` binary includes a built-in MCP (Model Context Protocol) server that
exposes every schedule operation as a tool. A Claude client connected to this
server can read, validate, and mutate schedules through natural language — the
same operations available via the CLI, with the same `.db` files. The easiest
way to install it is the [Quilt Desktop Plugin]({{< ref "desktop-plugin" >}}).

## Quick start

```bash
# build the binary
just cli-build

# start the server (blocks; clients connect via stdin/stdout)
./bin/quilt mcp
```

## Configuring in a Claude / MCP client

Add an entry to your MCP configuration file:

```json
{
  "mcpServers": {
    "quilt": {
      "command": "/path/to/bin/quilt",
      "args": ["mcp"]
    }
  }
}
```

For Claude Desktop on macOS the config file is at
`~/Library/Application Support/Claude/claude_desktop_config.json`. After
restarting the client, the **`quilt`** server should appear in the tool list.

## Self-documenting

The server carries its own operating instructions so a client always knows the
model:

- **Instructions** — an always-on manual (the `.db` model and the
  render-not-compute rule) advertised on connect.
- **Resources** — `guide://db-model`, `guide://migrate-xlsx`,
  `guide://live-views`, `guide://tools`, plus the `view://master` and
  `view://class` HTML skeletons.
- **Prompts** — `live_view` (args `kind`, `db`, `target`), `migrate_xlsx`
  (`xlsx_path`, `db_path`, `name`), and `new_schedule` (`db_path`, `name`).

## Transport

stdio is the default. Opt into Streamable HTTP with `--http`:

```bash
quilt mcp --http :8080
```

The tools are transport-agnostic; the same set is served either way.

## Design

- **Path-per-tool**: every tool takes a `db` argument (the path to a `.db`
  schedule file). There is no global state — work with multiple schedules in one
  session by passing different paths.
- **Two front doors, one function**: each MCP tool calls exactly the same
  `*store.Store` method as the corresponding CLI command. Verifying behavior via
  the CLI verifies it for the MCP path.

## Available tools

| Tool | Description |
|---|---|
| `init` | Create an empty schedule database |
| `info` | Return name, timestamps, counts, and rule status |
| `copy` | Branch a schedule into a new independent file |
| `import` | Import an `.xlsx` workbook into a new schedule |
| `add_class` | Add a class (grid row) |
| `add_timeslot` | Add a timeslot (grid column) |
| `add_cohort` | Add a cohort to the master list |
| `remove_class` | Remove a class and its assignments |
| `remove_timeslot` | Remove a timeslot and its assignments |
| `remove_cohort` | Remove a cohort from the master list |
| `enable_rule` | Enable or disable a validation rule |
| `set_travel` | Set building travel groups for a class |
| `add_blackout` | Forbid a cohort during a timeslot |
| `assign` | Assign a cohort to a (class, timeslot) cell |
| `unassign` | Clear a (class, timeslot) cell |
| `validate` | Get broken rules — returns all violations |
| `list_unassigned` | List every unassigned cell |
| `report` | Per-cohort calendar of assignments |
| `grid` | Full master schedule grid |
| `board` | Authoritative render payload — grid + violations + timeslots |
| `list_classes` | List all class names |
| `list_timeslots` | List all timeslots with metadata |
| `list_cohorts` | List all cohort names |
| `show_rules` | Dump all configured rules |
