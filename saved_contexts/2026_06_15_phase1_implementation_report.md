# Implementation Report: Phase 1 — SQLite Store + CLI

**Date:** 2026-06-15
**Author:** Claude
**Plan:** `~/.claude/plans/consider-everything-in-saved-contexts-staged-cook.md`
**Predecessors:** `saved_contexts/2026_06_13_feasibility_report.md` (the decision to
re-platform onto a Go core), `saved_contexts/2026_06_13_phase0_implementation_report.md`
(the rules engine this builds on)

---

## Summary

Phase 1 is **complete**. A SQLite-backed store (`pkg/store`) and a human-drivable
CLI (`cmd/scheduler`) now sit on top of the Phase 0 engine. The store owns a
schedule's `.db` file end-to-end — create/open, an embedded normalized schema,
CRUD, the `Load()` transform back into `engine.Grid`/`engine.Rules`, and the
shared operation methods that the Phase 2 MCP server will reuse verbatim. The
CLI is a thin verb-noun wrapper: parse flags → call one store method → render.

The guiding requirement is honored literally: **each `.db` file is one
self-contained schedule, used exactly like an `.xlsx` file.** There is no
registry and no global state — every command names the file it acts on, so the
filesystem *is* the document model. `copy` is a first-class "branch this
schedule" operation, and a plain `cp` works too.

**Status:** `just validate` (= `go test ./...` + `golangci-lint run`) clean;
`gofmt -l` clean. `bin/scheduler` builds via `just cli-build`.

---

## What was built

```
pkg/store/schema.sql      embedded normalized schema (7 tables)
pkg/store/store.go        lifecycle (Create/Open/Close), ctx-aware DB helpers, touch
pkg/store/load.go         relational → engine.Grid + engine.Rules (the Phase 0-deferred loader)
pkg/store/ops.go          shared operation surface (the CLI⇄MCP "one function" methods) + DTOs
pkg/store/copy.go         Copy(): verify + byte-copy + metadata refresh
pkg/store/store_test.go   unit tests: each rule → violation, Load round-trip, copy independence
cmd/scheduler/main.go     dispatch table, shared helpers, --json, usage
cmd/scheduler/lifecycle.go  init / info / copy
cmd/scheduler/edit.go     structure + rules + grid edit commands
cmd/scheduler/read.go     grid / list-* / show-rules / validate / list-unassigned / report
cmd/scheduler/main_test.go  CLI end-to-end via run(): build → validate exit, copy, errors
```

`go.mod` gained one runtime dependency, `modernc.org/sqlite` (pure Go, no cgo);
the `justfile` gained a `cli-build` recipe; `changelog.md` has an Unreleased
entry. `bin/` was already git-ignored.

---

## Schema (`pkg/store/schema.sql`)

Embedded via `//go:embed` and applied on `Create`. Normalized tables (not
`config_json` blobs) so they map cleanly to the four rules and the per-cohort
report is one SQL query:

| Table | Role |
| --- | --- |
| `meta` (singleton, `id=1`) | name, `schema_version`, the `one_class_at_a_time` toggle, `created_at`/`updated_at` |
| `class(id, name UNIQUE, sort_order)` | grid rows |
| `timeslot(id, label UNIQUE, day, period, sort_order)` | grid columns; `sort_order` drives both display and ClassRequiresTravel adjacency |
| `cohort(id, name UNIQUE, sort_order)` | the **AllCohorts** master list |
| `assignment(class_id, timeslot_id, cohort_value, PK(class_id,timeslot_id))` | a filled cell |
| `travel_group(class_id, building_index, cohort_name)` | ClassRequiresTravel groupings |
| `blackout(cohort_name, timeslot_id, PK(cohort_name,timeslot_id))` | CohortBlacklist |

Foreign keys are declared on `assignment`/`travel_group`/`blackout` with
`ON DELETE CASCADE` and enforced via `PRAGMA foreign_keys(1)` set in the
connection DSN (it is per-connection, so it lives in `dsn()`, not the schema).

---

## Loader (`pkg/store/load.go`)

`Load() (engine.Grid, engine.Rules, error)` is the relational→engine transform
Phase 0 deferred. It reads the class/timeslot axes ordered by `sort_order` (each
via the shared `loadAxis`, which also returns an id→index map), fills the
row-major `Cells` matrix from `assignment` (empty cells stay `""`), pulls
`AllCohorts` from `cohort`, groups `travel_group` rows into one `TravelRule` per
class (buildings ordered by `building_index`), joins `blackout` to `timeslot.label`
into per-cohort `BlacklistRule`s, and reads `OneClassAtATime` from `meta`. Every
validation/report path funnels through `Load`, so the engine sees exactly one
shape. `TestLoadRoundTrip` asserts the reconstructed `Grid`/`Rules` byte-for-byte.

---

## Operation surface (`pkg/store/ops.go`)

The CLI⇄MCP "one function" layer. Every method returns structured data (engine
types or the `TimeslotInfo`/`ReportEntry`/`Info` DTOs), never formatted text:

- **Structure:** `AddClass`, `AddTimeslot`, `AddCohort` (auto-append
  `sort_order`); `RemoveClass`/`RemoveTimeslot`/`RemoveCohort` (error if absent).
- **Rules:** `SetOneClassAtATime`, `SetTravelGroups(class, [][]cohort)` (in a
  transaction — delete then re-insert), `AddBlackout`.
- **Grid:** `Assign` (upsert; class+timeslot must exist), `Unassign`.
- **Queries:** `Validate` (→ `Load` → `engine.Validate`), `ListUnassigned`,
  `Report(cohort)` (one SQL join), `Info` (counts + rule status).
- **Inspect:** `Grid`, `Classes`, `Timeslots`, `Cohorts`, `RulesConfig` (several
  reuse `Load`).

Every mutating method bumps `meta.updated_at` via `touch()`.

---

## CLI (`cmd/scheduler`)

Form: `scheduler <command> <file.db> [flags]` — the file is the first positional
after the command, reinforcing file=document. Built on the stdlib `flag` package
with a small **map-based dispatch table** (`commands`), which doubles as a
preview of the Phase 2 MCP tool registry. Helpers (`needFile`, `openFile`,
`withFile`, `readCmd`) keep each command to flag-parse + one store call + render.
Read commands take `--json` (emits the raw structs — the Phase 2 MCP shape);
`validate` prints violations and **exits non-zero** when any exist (via an
`errSilent` sentinel so no spurious `error:` line is printed). Full surface:
`init`/`info`/`copy`; `add-*`/`remove-*`; `enable-rule`/`set-travel`
(repeated `--building`)/`add-blackout`; `assign`/`unassign`; `grid`/`list-*`/
`show-rules`/`validate`/`list-unassigned`/`report`.

---

## Notable decisions & deviations

- **Free-text `assignment.cohort_value`, NOT a FK to `cohort`** (deliberate
  divergence from feasibility report §3.2). A FK would make the AllCohorts
  *validation* unenforceable — flagging values not in the master list is the
  engine's job — and would forbid exempt placeholders like `#### closed`, which
  are valid cell values but not cohorts. `travel_group.cohort_name` and
  `blackout.cohort_name` are likewise free text, matched by name like the engine.
  `RemoveCohort` therefore intentionally leaves assignments alone: orphaned cells
  simply become AllCohorts violations, which is the intended signal.
- **One schedule per `.db` file, no registry.** `Copy` verifies the source is a
  real scheduler db, byte-copies, then refreshes `created_at` (and `name` if
  given) so the branch reads as a fresh schedule. `TestCopyIndependence` proves
  editing a copy does not touch the original.
- **`.xlsx` importer deferred** (per the plan). Schedules are populated via CLI
  commands now; the importer is a clean standalone subcommand in a later step,
  reusing `splitArrayByEmptyStrings`.
- **`go 1.23` → `go 1.25.0`.** `modernc.org/sqlite v1.52.0` requires a newer Go,
  so `go mod tidy` raised the module's `go` directive. `go.mod`'s only direct
  dependency is the sqlite driver; the rest are its runtime indirects.
- **Lint adaptations** to satisfy the repo's golangci-lint config: DB calls go
  through context-aware helpers (`s.exec`/`s.query`/`s.queryRow`/`s.begin` with
  `context.Background()`) to satisfy `noctx` while keeping the plan's
  context-free public API; the generic table-name builder was dropped in favor of
  explicit per-table SQL to avoid the `gosec` G202 concatenation warning; and the
  command dispatcher is a map rather than a giant `switch` to stay under
  `gocyclo`.

---

## Verification

```
$ just validate
go test ./...
ok  github.com/bmayfi3ld/excel-scheduler/cmd/scheduler
ok  github.com/bmayfi3ld/excel-scheduler/pkg/engine
ok  github.com/bmayfi3ld/excel-scheduler/pkg/store
golangci-lint run
0 issues.

$ gofmt -l pkg/store cmd/scheduler    # clean (no files listed)
```

Multi-file smoke (the headline requirement): `init 2026.db` → build it → `copy`
to `2027-draft.db --name "2027-2028 draft"` → add a class to the draft →
`info` on both shows they diverged (original 1 class, draft 2; distinct names),
original untouched. `validate` returns a cross-building `ClassRequiresTravel`
violation and exit code 1, matching the engine's `rules.md` fixtures.

Store unit tests cover each rule driven to a violation through the store (a
cross-building Latin Cart case, a blackout hit, a OneClassAtATime double-book, an
unknown cohort, an exempt placeholder, and a clean grid → 0), a `Load`
round-trip, assign-overwrite + unassign, missing-reference errors, and FK
cascade on `RemoveClass`. CLI tests drive `run()` end-to-end (init → add* →
assign → validate exit behavior), the error paths, and `copy` independence.

---

## `.xlsx` importer (`scheduler import`) — addendum 2026-06-15

**Plan:** `~/.claude/plans/consider-the-current-state-serialized-wilkinson.md`
**Feasibility reference:** `saved_contexts/2026_06_13_feasibility_report.md` §Appendix line 248

The migration tool called out in the feasibility report is now implemented.

### What was built

```
pkg/ingest/xlsx.go      ReadWorkbook(): excelize OpenFile → Sheet ([][]string, trimmed, row-major)
pkg/ingest/parse.go     Pure parsing — splitByBlanks, column, splitTimeslot, Parse, parseTravel, parseBlacklist
pkg/ingest/apply.go     Apply(): orchestrates store writes in dependency order, warnings vs. fatal errors
pkg/ingest/parse_test.go  Table tests: splitByBlanks, splitTimeslot, end-to-end Parse fixture
cmd/scheduler/import.go  cmdImport: read → parse → Create → Apply → print summary + Validate
```

`go.mod` gained `github.com/xuri/excelize/v2 v2.10.1` (direct import confined to `xlsx.go`).
`scheduler import` is registered in the `commands` map and in `usage()`.
`changelog.md` has an Unreleased `### Added` entry.

### Source-sheet layout parsed

**`Schedule` sheet** (`A1:AO14` in the example workbook):
- Row 0: timeslot labels in cols 1+ (`"Monday, 8:40-9:20"`, …); col 0 empty.
- Rows 1+: col 0 = class name; interior cells = cohort value or blank.

**`Rules` sheet** (`A1:D149` in the example workbook), blank-delimiter convention:
- Col A `AllCohorts`: flat vertical list of valid cohort names (no blanks between entries).
- Col B `ClassRequiresTravel`: class name on its own row, then a blank, then building-1 cohorts (one per row), then blank, then building-2 cohorts, then two blanks before the next class.
- Col C `CohortBlacklist`: cohort name on its own row, then a blank, then the forbidden timeslot labels (one per row), then two blanks before the next cohort.
- Col D `OneClassAtATime`: header presence alone enables the rule.

The blank-delimiter semantics match the legacy `splitArrayByEmptyStrings` add-in convention. **Key structural detail discovered from the real file:** both `ClassRequiresTravel` and `CohortBlacklist` place their group *header* (class name / cohort name) on a single-element sub-list separated by a blank from the data sub-list(s). `parseTravel` and `parseBlacklist` both use the same state machine: first non-nil sub-list after reset = header; subsequent sub-lists = data; nil (double-blank) = group separator and reset.

### Warning-tolerant design

`Apply` wraps every store call in `tryApply`, which converts UNIQUE constraint violations and "not found" errors into `Warning` entries rather than aborting. Genuine I/O/DB errors still abort and remove the partial `.db` file.

The `Friday 2:40-3:20` (no comma) / `Friday, 2:40-3:20` (with comma) mismatch documented in the plan is the primary real-world instance of this: `AddBlackout` returns `timeslot "Friday 2:40-3:20" not found`, which becomes a warning.

### Verification against `25-26 Schedule Example.xlsx`

```
$ rm -f /tmp/25-26.db
$ bin/scheduler import /tmp/25-26.db --from "25-26 Schedule Example.xlsx" --name "25-26"
Imported "25-26" → /tmp/25-26.db

  Classes:         11
  Timeslots:       40
  Cohorts:         30
  Assignments:     242   (217 unique cells; 25 overwrites via ON CONFLICT UPDATE)
  Travel rules:    2     (Latin A, Latin B — two buildings each)
  Blackouts:       104
  OneClassAtATime: true

Warnings (6):
  ! blackout P3B@Friday 2:40-3:20: timeslot "Friday 2:40-3:20" not found
  ! blackout P3M@Friday 2:40-3:20: ...
  (6 total — all comma-mismatch for the same timeslot, 6 different cohorts)

Validation: 6 violation(s)

$ bin/scheduler info /tmp/25-26.db --json     # counts confirmed
$ bin/scheduler show-rules /tmp/25-26.db      # travel + blackouts render correctly
$ bin/scheduler validate /tmp/25-26.db        # 6 pre-existing OneClassAtATime/AllCohorts violations
```

`just validate` passes (0 lint issues) with the new package included.

---

## Kong CLI refactor — addendum 2026-06-15

The stdlib `flag`-based dispatcher described in the "CLI" section above was replaced with `github.com/alecthomas/kong` in the same session.

### What changed

- **Breaking positional → flag**: the `.db` file argument moved from the first positional after the command to a named `--db <file.db>` flag (short `-d`). `copy` uses `--db <src>` for the source and `--out <dst>` for the destination. `init` and `import` use `--db` for the output path.
- **~200 lines removed**: per-command `FlagSet` setup, required-flag checks, `needFile`/`openFile`/`withFile`/`readCmd` helpers, and the hand-written `usage()` function are gone. Each command is now a struct with field-tag-declared flags and a single `Run(...)` method.
- **Auto-generated `--help`**: kong produces `--help` output for every command from the struct tags; no manual `usage()` maintenance.
- **`withDB` / `closeReg` pattern**: commands that operate on an existing schedule embed `withDB`, whose `AfterApply` hook opens the store and binds it for injection into `Run(*store.Store)`. `closeReg` collects all opened stores so `defer reg.closeAll()` is the only teardown site.
- **`just validate` description updated**: validate now runs `go test ./...`, `go vet ./...`, `govulncheck ./...`, and `golangci-lint run` (the previous description said only tests + lint).

### Why

The dispatch-table approach worked but required ~10 lines of flag wiring per command and a manually-kept `usage()` string. Kong kept the same one-function-per-command design while dropping the boilerplate and making `--help` authoritative.

---

## Out of scope / next

- **Phase 2 — MCP server** wrapping `pkg/store` ops (CLI command ⇄ MCP tool, one
  function underneath). The operation methods already return structured DTOs and
  the dispatch table is keyed by name, so the wiring is mechanical. `pkg/ingest`
  gives the MCP server a `ReadWorkbook`→`Parse`→`Apply` path with no duplication.
- **Phase 3** — master grid UI + Live Artifacts. **Phase 4** — packaging/signing
  (the pure-Go driver keeps cross-compile painless).
