# Quilt MCP tools

Every tool takes a `db` argument (path to a `.db` schedule). Mutation tools
return a short text confirmation; query tools return structured JSON.

## Lifecycle
- `init` — create an empty schedule.
- `import` — build a schedule from an `.xlsx` workbook.
- `copy` — branch a schedule into a new independent file.
- `info` — name, timestamps, entity counts, rule status.

## Structure
- `add_class` / `remove_class` — grid rows.
- `add_timeslot` / `remove_timeslot` — grid columns (label + optional day/period).
- `add_cohort` / `remove_cohort` — the `AllCohorts` master list.

## Rules
- `enable_rule` — toggle `one-class-at-a-time`.
- `set_travel` — per-class building travel groups.
- `add_blackout` — forbid a cohort during a timeslot.

## Grid edits
- `assign` — place a cohort in a `(class, timeslot)` cell (overwrites).
- `unassign` — clear a cell.

## Queries
- `board` — **the render payload**: grid + violations + timeslots in one call.
  Prefer this for views.
- `grid` — full master grid.
- `validate` — every violation (empty list = clean).
- `list_unassigned` — empty cells.
- `report` — per-cohort calendar.
- `list_classes` / `list_timeslots` / `list_cohorts` / `show_rules`.

After any edit, re-call `board` or `validate` — the engine is the source of
truth for validity.
