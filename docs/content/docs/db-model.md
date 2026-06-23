+++
title = 'The .db Model'
weight = 15
+++
# The Quilt `.db` model

A Quilt schedule is **one self-contained SQLite `.db` file**. There is no
registry, server-side state, or shared database — a `.db` *is* the schedule.
Copy the file and you have a branched schedule; delete it and the schedule is
gone. Every MCP tool (and every `quilt` CLI command) takes a `db` path, so one
session can work on many schedules at once by passing different paths.

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

| Action | CLI | MCP tool |
|---|---|---|
| Create empty | `quilt init --db s.db` | `init` |
| Import workbook | `quilt import --db s.db --from w.xlsx` | `import` |
| Branch | `quilt copy --db s.db --out new.db` | `copy` |
| Edit | `quilt assign …` / `quilt add-class …` | `assign`, `add_class`, … |
| Inspect | `quilt grid --db s.db` | `board`, `grid`, `validate`, … |

## The golden rule

{{< hint info >}}
The **engine** decides validity; tools and views only render. After any
`assign` / `unassign` (or any structural edit), re-call `board` (or `validate`)
to get fresh violations — never infer validity yourself.
{{< /hint >}}
