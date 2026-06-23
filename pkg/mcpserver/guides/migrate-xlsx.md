# Migrating an `.xlsx` workbook into a `.db`

The legacy Excel add-in stored a schedule in a workbook with a `Schedule` sheet
and a `Rules` sheet. The `import` tool (and `quilt import` CLI) reads that
workbook directly and writes a fresh `.db`.

## Usage

MCP tool `import`:

```json
{ "db": "schedule.db", "from": "25-26 Schedule.xlsx", "name": "25-26" }
```

CLI:

```bash
quilt import --db schedule.db --from "25-26 Schedule.xlsx" --name "25-26"
```

Optional `scheduleSheet` / `rulesSheet` override the default sheet names
(`Schedule` and `Rules`).

## The blank-delimiter convention

Both sheets use **blank rows/columns as section delimiters** — the same
convention the add-in used. The importer follows it to find the grid, the
`AllCohorts` list, travel groups, and blackouts.

## Expected warnings

`import` is tolerant: referential mismatches become **warnings**, not errors,
so a partial workbook still imports. The most common is a **comma-mismatch on
timeslot labels** — a blackout that names a timeslot whose label doesn't
exactly match a schedule-grid header. The cell still imports; the warning tells
you to reconcile the label. After import, run `validate` to see the resulting
violations.
