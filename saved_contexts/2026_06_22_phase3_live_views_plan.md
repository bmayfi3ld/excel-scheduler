# Phase 3 Plan — Cowork Live Views, Self-Documenting MCP, Packaging

> **SUPERSEDED / AS-BUILT 2026-06-23.** This is the original plan. Phase 3 has
> shipped — see `2026_06_23_phase3_implementation_report.md` for what was
> actually built, the deviations (notably the `.mcp.json` block being left as a
> manual step and the embedded-docs/views being mirrored copies rather than
> byte-identical), and the remaining follow-ups (`.dxt` signing, Canva).

*Date: 2026-06-22*

Folds together: (a) live views for the master + per-class schedules, (b) a
self-documenting MCP server so Claude Cowork always knows how the `.db` model
and `xlsx→db` migration work, (c) docs pages that double as MCP resources, and
(d) the DXT/packaging bundle that ships all of it. End goal (NOT built here):
feed live-view data into a Canva template — the data shape is designed for it.

---

## 0. What already exists (verified)

- 23 stdio MCP tools (`pkg/mcpserver`), server name **`excel-scheduler`**.
- A working Cowork master-grid prototype: `index.html` (166 lines). It uses
  `window.cowork.callMcpTool('mcp__excel-scheduler__grid'|'__validate', {db})`,
  a `<script id="cowork-artifact-meta">` header declaring `mcpTools` +
  `mcpServerNames`, reads `structuredContent` (fallback `content[0].text`),
  and paints `filled`/`conflict`/`invalid` cells. Refreshes every 30s.
- Tool shapes confirmed from the prototype:
  - `grid` → `{ Classes:[], Timeslots:[], Cells:[[]] }` (row-major, Cells[ci][si]).
  - `validate` → `{ violations:[{ Cell:{Class,Timeslot}, Rule } ] }`.
  - Timeslot labels are `"Monday, 8:40-9:20"`.
- Docs are Hugo pages in `docs/content/docs/` (install/setup/rules/functions/
  mcp-server). `mcp-server.md` lists all tools.

### Two real gaps the prototype exposes
1. **Fragile day bucketing.** `index.html` hardcodes `activeDay*8` (8 slots/day,
   5 days). Real schedules vary. `list_timeslots` already returns day/period
   metadata — views must bucket by that, not by index arithmetic.
2. **Server is "tools-only."** No MCP **resources**, **prompts**, or top-level
   **instructions**. So Cowork has no in-band way to learn the db model, the
   migration path, or how to build a view — it only sees 23 tool signatures.

---

## 1. Self-documenting MCP server (the core of "Cowork always knows")

Add three capabilities to `pkg/mcpserver` beyond tools. All content is authored
once as markdown, `go:embed`-ed into the binary, and served three ways.

### 1a. Server `instructions` (always-on operating guidance)
Set the go-sdk server `Instructions` field on construction. Short, always
injected on connect. Contents:
- The mental model: **one `.db` file = one schedule, like an `.xlsx`. No
  registry — every tool takes a `db` path.**
- The discipline: **views render, the engine computes. After any
  `assign`/`unassign`, re-call `board`/`validate` — never infer conflicts in JS.**
- Pointers: "For the db model, read resource `guide://db-model`. To build a
  live view, use prompt `live_view`. To migrate an xlsx, use prompt `migrate_xlsx`."

### 1b. MCP resources (reference docs, pulled on demand)
Register the docs pages as resources (URI scheme `guide://`). Reuse the SAME
markdown that Hugo renders (single source — see §3), embedded via `go:embed`:
| Resource URI | Content |
|---|---|
| `guide://db-model` | What a `.db` holds (classes/timeslots/cohorts/assignments/rules), how to use it, lifecycle (init/copy/import). |
| `guide://migrate-xlsx` | Step-by-step `import` from a legacy `.xlsx`, the blank-delimiter rules sheet, expected warnings (e.g. comma-mismatch timeslots). |
| `guide://tools` | The tool reference table (generated/kept in sync with `mcp-server.md`). |
| `guide://live-views` | How to build a Cowork live artifact off these tools (the recipe cookbook, §4). |

### 1c. MCP prompts (parameterized, named recipes)
Register prompts that return ready-to-run instruction text. These are how a
non-expert triggers a view or a migration without writing the prompt by hand:
| Prompt | Args | Returns |
|---|---|---|
| `live_view` | `kind` (master\|class\|cohort), `db`, optional `target` | Full instruction + the matching HTML template skeleton (§4) wired to the right tools. |
| `migrate_xlsx` | `xlsx_path`, `db_path`, `name` | Stepwise `import` → `validate` → review-warnings flow. |
| `new_schedule` | `db_path`, `name` | `init` then how to populate. |

**Scope/lint note:** registering resources/prompts is net-new wiring but small;
keep each `register*` function under gocognit ≤20 / funlen ≤80 as Phase 2 did.

---

## 2. New data tool: `board` (one authoritative render payload)

Today a view calls `grid` + `validate` (+ should call `list_timeslots`)
separately and re-derives day buckets in JS. Add one composition tool.

- **`store.Board(db)`** → `{ Grid:engine.Grid, Violations:[]engine.Violation,
  Timeslots:[]TimeslotInfo }`. Pure composition of existing `Grid()` +
  `Validate()` + `Timeslots()`. No new domain logic; engine stays sole truth.
- **`board` MCP tool** wraps it. `TimeslotInfo` carries `Day`/`Period`/`Label`
  so views bucket by real metadata (fixes gap #1) — no `*8` arithmetic.
- This is also the **Canva export shape** (see §5): one flat, authoritative
  object per schedule.
- Test: round-trip asserting `board` == separate `grid`/`validate`/`list_timeslots`.

No new mutation tools — `assign`/`unassign` already cover write-back.

---

## 3. Docs pages (Hugo + reused as MCP resources — single source)

Author markdown ONCE under `docs/content/docs/`, embed the same files into the
binary for §1b. New/updated pages:
| Page | Purpose |
|---|---|
| `db-model.md` (new) | The `.db` document model, how to use it, lifecycle. → `guide://db-model` |
| `migrate-xlsx.md` (new) | xlsx→db migration walkthrough + warning catalog. → `guide://migrate-xlsx` |
| `live-views.md` (new) | Cowork live-view cookbook: the meta header, `callMcpTool` API, the master/class/cohort recipes, the "render-not-compute" rule. → `guide://live-views` |
| `mcp-server.md` (update) | Add `board`; document resources + prompts + instructions; note HTTP transport (§6). |

Mechanism: `//go:embed` a `docs/content/docs/*.md` subset into `pkg/mcpserver`
(or a small `pkg/guide` embed package the server reads), so Hugo and MCP serve
byte-identical text. Updating the doc updates what Cowork sees. CI check: embed
list matches the files on disk.

---

## 4. Live views — concrete build plan

Base everything on the working `index.html`, generalized and de-fragilized.
Three view kinds, each a self-contained HTML artifact with the
`cowork-artifact-meta` header and a `load()` that calls `board`.

### 4a. Master total schedule (generalize `index.html`)
- Switch data source from `grid`+`validate` to the single **`board`** tool.
- **Bucket columns by `Timeslot.Day`/`Period`** from `board` metadata instead of
  `activeDay*8`. Day tabs built from the distinct `Day` values present.
- Keep the cell styling (filled / conflict=OneClassAtATime / invalid=AllCohorts).
  Also surface ClassRequiresTravel + CohortBlacklist violation types (the
  prototype only handled two of four rules).
- Keep 30s auto-refresh + manual refresh.
- Deliverable: `packaging/live-views/master.html` (the canonical, parameterized
  version) and the `live_view kind=master` prompt that emits it.

### 4b. Per-class schedules (one class across the week)
- For a chosen class (grid row), render that row over all timeslots, laid out as
  a week calendar (Day × Period) showing the assigned cohort per cell.
- Data: filter `board.Grid.Cells[classIndex]` by day/period buckets; overlay any
  violations on that row.
- One view per class, or a class-picker dropdown in a single artifact.
- Deliverable: `packaging/live-views/class.html` + `live_view kind=class`
  (arg: class name). This is the layout intended to map onto a Canva
  per-class template (§5).

### 4c. Per-cohort calendars (student-group view) — bonus, near-free
- Use existing **`report`** tool (per-cohort assignments) → week calendar for one
  cohort. Deliverable: `cohort.html` + `live_view kind=cohort`. Lower priority;
  include because `report` already returns exactly this.

### Verification for all views
Numbers and cells in every view must equal CLI output for the same `.db`
(`grid`/`validate`/`report`). Drive the headline loop from Cowork: assign a
cohort → view reflects it on next refresh; create a OneClassAtATime double-book
→ correct cell turns conflict-colored.

---

## 5. Canva end goal (design now, integrate later — NOT built here)

Not integrated in this phase, but the data shape is chosen so the hop is small.

- **`board` is already the export payload.** Canva Connect API "Autofill"
  (bulk create from a data set) wants a flat list of `{field: value}` rows per
  generated design.
- Plan a thin **`export` shape** (a documented transform of `board`, no new
  engine logic): for per-class templates, one row per class with fields like
  `class_name`, `mon_p1`…`fri_pN`; for per-cohort, one row per cohort. This is a
  pure reshape of `board` → Canva autofill rows.
- Document the mapping in `live-views.md` ("Canva handoff") so when Canva
  integration is built, the field contract already exists. The live views in §4
  and the future Canva designs render the **same `board` data** — parity by
  construction.
- Explicitly out of scope now: Canva Connect OAuth, the autofill API call, the
  template IDs.

---

## 6. Transport: add HTTP/SSE (needed for live refresh outside stdio)

The Cowork prototype works because Cowork brokers `callMcpTool` over its own
connection to the stdio server. Keep stdio as default. Add an opt-in
**`--http <addr>`** flag to `scheduler mcp` (go-sdk Streamable HTTP handler) for
hosts that need a network endpoint (and for the future Canva/automation path).
- `pkg/mcpserver`: a `RunHTTP(ctx, addr)` alongside the stdio path; tools stay
  transport-agnostic.
- Default unchanged so existing desktop config keeps working.
- Test: HTTP smoke test; `ListTools` returns the full set incl. `board`.

---

## 7. Packaging — bundle binary + docs + views (folds prior Phase 4 plan)

One pure-Go binary; GoReleaser cross-compile matrix feeds every wrapper.

- **`version` var + `--version`** in `cmd/scheduler/main.go` (only source change
  for packaging; ldflags inject it).
- **`.goreleaser.yaml`** — `CGO_ENABLED=0`; `goos`=[linux,darwin,windows],
  `goarch`=[amd64,arm64]; archives include `LICENSE`, the docs pages, and the
  `packaging/live-views/*.html` templates.
- **`.github/workflows/release.yml`** — on tag `v*`: `just validate` gate →
  goreleaser. `just release-snapshot` for local dry-run.
- **Claude Desktop Extension (`.dxt`)** — `packaging/dxt/manifest.json`:
  `server.command = ${__dirname}/scheduler`, `args:["mcp"]`. **The DXT zip also
  carries the markdown docs + the live-view HTML templates** so they ship
  offline alongside the binary. (Loose files in the zip are inert to Claude on
  their own — they reach the model via the MCP resources/prompts of §1, which
  the bundled binary serves. The bundle guarantees they're present offline; MCP
  is how they're surfaced.)
- **Claude Code plugin (secondary)** — `.claude-plugin/plugin.json` +
  add an `excel-scheduler` server block to `.mcp.json` (today it only has
  code-review-graph). In the Claude Code world, the markdown can ALSO ride as
  native `.claude/skills/*.md` (skills are read directly there, unlike Desktop).
- **Signing** — macOS Gatekeeper blocks unsigned `.dxt` (hard blocker for
  non-technical school admins); needs Apple Developer ID + notarytool in CI.
  Decision gate: ship unsigned w/ right-click-open workaround vs. pay+notarize.
  Windows SmartScreen similar; Linux fine.

---

## 8. Build order (with stop/verify gates)

1. **`board` tool + `store.Board()`** (+ test). Smallest, unblocks de-fragilized
   views. Verify: round-trip parity with separate tools.
2. **MCP resources + prompts + instructions** (§1) reading `go:embed`-ed docs
   (§3). Verify: connect from Cowork; `ListResources`/`ListPrompts` populated;
   server instructions visible.
3. **Docs pages** db-model / migrate-xlsx / live-views (§3). Verify: Hugo builds;
   embedded bytes == files (CI check).
4. **Master view on `board`** — generalize `index.html`, bucket by real day/
   period, cover all 4 rule types (§4a). Verify: parity with CLI; conflict paint.
5. **Per-class view** (§4b). Verify: one class's week matches `grid` row.
6. **Per-cohort view** (§4c, bonus). Verify against `report`.
7. **HTTP transport** behind `--http` (§6). Verify: smoke test + manual connect.
8. **Canva export shape doc** (§5) — no code, just the documented `board`→rows
   mapping in live-views.md.
9. **Packaging** (§7): version flag → goreleaser → DXT bundling docs+views →
   plugin → (signing gated).

---

## Out of scope (explicit)
- Canva Connect integration (OAuth, autofill API, template IDs) — §5 designs the
  data contract only.
- A pre-authored Interactive Connector React component — chat/Cowork artifact is
  the surface; harden later only if it earns it.
- New engine rules — Phase 3 is surface + plumbing only.

## Critical files
- `pkg/mcpserver/server.go`, `tools.go` (instructions, resources, prompts, `board`)
- `pkg/store/ops.go` (`Board()`)
- `pkg/mcpserver/server_test.go` (board + resource/prompt tests)
- `cmd/scheduler/mcp.go`, `main.go` (`--http`, `--version`)
- `docs/content/docs/{db-model,migrate-xlsx,live-views,mcp-server}.md`
- `packaging/live-views/{master,class,cohort}.html` (from `index.html`)
- `packaging/dxt/manifest.json`, `.goreleaser.yaml`, `.github/workflows/release.yml`,
  `.claude-plugin/plugin.json`, `.mcp.json`
