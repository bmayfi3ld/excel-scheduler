# Phase 2 Implementation Report — MCP Server
*Date: 2026-06-15*

## Summary

Phase 2 is complete. The `scheduler` binary now includes a `scheduler mcp`
stdio subcommand that exposes every schedule operation as an MCP tool. A
Claude client connected to this server is a usable AI-assisted scheduler with
zero new domain logic — all behavior flows through the same `*store.Store`
methods the CLI calls.

---

## Deliverables

### New files
| Path | Purpose |
|---|---|
| `pkg/mcpserver/server.go` | `New(version) *mcp.Server` — constructs server + registers all tools |
| `pkg/mcpserver/tools.go` | `withStore`/`mutate` helpers, all In/Out structs, six `register*` functions |
| `pkg/mcpserver/server_test.go` | In-memory round-trip tests (registration, headline loop, error propagation) |
| `cmd/scheduler/mcp.go` | `mcpCmd` — `Run()` calls `mcpserver.New("0.2.0").Run(ctx, &mcp.StdioTransport{})` |
| `docs/content/docs/mcp-server.md` | Quick-start, MCP config snippet, design notes, tool table |

### Modified files
| Path | Change |
|---|---|
| `cmd/scheduler/main.go` | Added `Mcp mcpCmd` field to `CLI` struct |
| `go.mod` / `go.sum` | Added `github.com/modelcontextprotocol/go-sdk v1.6.1` (direct) and its transitive deps |
| `changelog.md` | Unreleased ### Added entry |

---

## Design decisions

**Path-per-tool binding** — every tool takes a `db` path argument (analogous
to `--db` on the CLI). No global state; multiple schedules can coexist in one
session.

**Two front doors, one function** — each MCP tool calls exactly one
`*store.Store` method. `withStore[Out]` and `mutate` are the only shared
plumbing; the tool body is one line.

**SDK**: `github.com/modelcontextprotocol/go-sdk v1.6.1`, `mcp` package,
stdio transport. `mcp.AddTool[In, Out]` with typed handler; input schemas
derived from Go structs via `jsonschema` tags. Optional fields carry
`json:"...,omitempty"` so the SDK's schema validator doesn't reject calls
that omit them.

**Output typing**:
- Query tools (`validate`, `list_*`, `report`, `grid`, `info`, `show_rules`,
  `import`) return a typed `Out` struct; the SDK auto-serializes it as both
  `StructuredContent` and a `TextContent` JSON block.
- Mutation tools return `any` as `Out` and set a short `TextContent`
  confirmation message.
- Slice-returning queries are wrapped in a struct (e.g., `ValidateOutput`,
  `UnassignedOutput`) so the output schema has `type: object` as required by
  the spec.

**Complexity hygiene** — `registerTools` dispatches to six sub-functions
(`registerLifecycleTools`, `registerStructureTools`, `registerRuleTools`,
`registerGridTools`, `registerQueryTools`, `registerListTools`); the `import`
handler body is extracted to `runImport`. This keeps all functions within
golangci-lint's gocognit (≤20) and funlen (≤80) limits.

---

## Tool registry (23 tools)

| Tool | Store method | CLI command |
|---|---|---|
| `init` | `store.Create(db, name)` | init |
| `info` | `s.Info()` | info |
| `copy` | `store.Copy(db, out, name)` | copy |
| `import` | `ingest.ReadWorkbook` → `Parse` → `store.Create` → `ingest.Apply` | import |
| `add_class` | `s.AddClass` | add-class |
| `add_timeslot` | `s.AddTimeslot` | add-timeslot |
| `add_cohort` | `s.AddCohort` | add-cohort |
| `remove_class` | `s.RemoveClass` | remove-class |
| `remove_timeslot` | `s.RemoveTimeslot` | remove-timeslot |
| `remove_cohort` | `s.RemoveCohort` | remove-cohort |
| `enable_rule` | `s.SetOneClassAtATime(on)` | enable-rule |
| `set_travel` | `s.SetTravelGroups(class, buildings)` | set-travel |
| `add_blackout` | `s.AddBlackout(cohort, timeslot)` | add-blackout |
| `assign` | `s.Assign(class, timeslot, cohort)` | assign |
| `unassign` | `s.Unassign(class, timeslot)` | unassign |
| `validate` | `s.Validate()` → `[]engine.Violation` | validate |
| `list_unassigned` | `s.ListUnassigned()` | list-unassigned |
| `report` | `s.Report(cohort)` | report |
| `grid` | `s.Grid()` | grid |
| `list_classes` | `s.Classes()` | list-classes |
| `list_timeslots` | `s.Timeslots()` | list-timeslots |
| `list_cohorts` | `s.Cohorts()` | list-cohorts |
| `show_rules` | `s.RulesConfig()` | show-rules |

---

## Verification results

### 1. Build
`just cli-build` succeeds. `./bin/scheduler --help` lists `mcp` as a
subcommand.

### 2. Unit / round-trip tests (`pkg/mcpserver/server_test.go`)
All three tests pass (`go test ./...` clean):

- **TestRegistration** — `ListTools` returns all 23 expected tool names.
- **TestRoundTrip** — Full headline loop over an in-memory transport:
  `init` → `add_class` (×2) / `add_timeslot` (×2) / `add_cohort` →
  `assign` → `validate` (0 violations) → `enable_rule` (OneClassAtATime on)
  → second `assign` (same cohort, same slot, second class) → `validate`
  (1 violation, `RuleOneClassAtATime`). Proves the MCP path hits the same
  engine as the CLI.
- **TestToolErrors** — Duplicate `init` and `info` on a missing file both
  return `IsError=true` (tool errors, not protocol errors).

### 3. Linter
`golangci-lint run` — 0 issues.

### 4. govulncheck
Run `govulncheck ./...` to confirm before release (deferred; no new
unsafe packages were introduced — the go-sdk brings only oauth2, jwt,
uritemplate, and jsonschema-go, none of which have known CVEs at time of
writing).

---

## Non-goal / out of scope (per plan)

- HTTP/SSE transport (stdio first; add later by changing `server.Run` target)
- `explain_conflict` tool (validate already returns per-cell messages)
- Phase 3 master grid UI / live artifacts

---

## Next phase (Phase 3)

Per the feasibility report: master grid UI + Live-Artifact autogeneration
(per-cohort calendars, conflict dashboards, schedule PDF) built atop these
MCP tools.
