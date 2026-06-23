# The Quilt `.db` model

A Quilt schedule is **one self-contained SQLite `.db` file**. There is no
registry, server-side state, or shared database — a `.db` *is* the schedule.
Copy the file and you have a branched schedule; delete it and the schedule is
gone. Every MCP tool takes a `db` path argument, so one session can work on
many schedules at once by passing different paths.

## What a schedule contains

- **Classes** — grid rows (e.g. "Latin Cart", "Music"). Display order is fixed.
- **Timeslots** — grid columns, in chronological order. Each has a `label`
  and optional `day` / `period` metadata. Column adjacency drives the
  `ClassRequiresTravel` rule.
- **Cohorts** — the master list of valid cohort names (the `AllCohorts` rule).
- **Assignments** — a cohort placed in a `(class, timeslot)` cell.
- **Rules** — `OneClassAtATime` flag, per-class travel groups, and per-cohort
  blackouts.

## Lifecycle

- **Create**: `init` makes an empty schedule at a path.
- **Import**: `import` builds a schedule from an `.xlsx` workbook (see the
  migrate-xlsx guide).
- **Branch**: `copy` clones a `.db` into a new independent file.
- **Edit**: `add_class` / `add_timeslot` / `add_cohort`, `assign` / `unassign`,
  `enable_rule` / `set_travel` / `add_blackout`.
- **Inspect**: `info`, `grid`, `board`, `validate`, `report`, the `list_*`
  tools, and `show_rules`.

## The golden rule

The **engine** decides validity; tools and views only render. After any
`assign` / `unassign` (or any structural edit), re-call `board` (or `validate`)
to get fresh violations — never infer validity client-side.
