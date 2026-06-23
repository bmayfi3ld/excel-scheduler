+++
title = 'Live Views'
weight = 55
+++
# Live views cookbook

A **live view** is a self-contained HTML artifact that renders a schedule and
refreshes itself by calling the `quilt` MCP server. Canonical templates ship in
the release archive under `packaging/live-views/` (`master.html`, `class.html`)
and are also served by the MCP server as the `view://master` and `view://class`
resources.

## The meta header

Every view begins with a `cowork-artifact-meta` JSON script declaring the tools
and server it uses:

```html
<script type="application/json" id="cowork-artifact-meta">
{
  "name": "Master Schedule",
  "schemaVersion": 1,
  "description": "Live master schedule grid.",
  "mcpTools": ["mcp__quilt__board"],
  "mcpServerNames": ["quilt"]
}
</script>
```

## Calling tools

Views call `window.cowork.callMcpTool(name, args)` and read
`res.structuredContent` (falling back to `JSON.parse(res.content[0].text)`).

## The one tool to call: `board`

`board` returns `{ grid, violations, timeslots }` in a single call. **Use it
instead of calling `grid` + `validate` separately** — it guarantees the grid and
its violations are consistent, and `timeslots[]` carries real `day` / `period`
metadata so you can build day tabs and period columns without index math.

```js
const res = await window.cowork.callMcpTool('mcp__quilt__board', { db: DB });
const { grid, violations, timeslots } = res.structuredContent;
```

Map violations by `cell.Class + '|' + cell.Timeslot + '|' + rule` to paint
cells. Paint all four rule types: `OneClassAtATime`, `AllCohorts`,
`ClassRequiresTravel`, `CohortBlacklist`.

## Recipes

- **Master view** (`master.html`) — classes × timeslots, day tabs built from
  distinct `timeslot.day`, period columns from distinct `timeslot.period`.
- **Per-class view** (`class.html`) — one class rendered as a Day × Period
  calendar (one grid ROW across the week), with that row's violations overlaid.

The fastest way to scaffold one is the **`live_view` prompt** (args:
`kind=master|class`, `db`, `target`), which returns the matching skeleton
ready to fill in.

## The render-not-compute rule

{{< hint info >}}
Views **render** `board`; they never recompute validity. After any
`assign`/`unassign`, re-call `board`. The engine is the single source of truth.
{{< /hint >}}

## Canva handoff (designed, not integrated)

`board` is also the export payload for a Canva autofill template. The reshape is
pure: for a per-class template, produce **one row per class** with fields
`class_name`, `mon_p1` … `fri_pN`, filling each from the cohort at the matching
`(class, day, period)` cell of `board.grid`. Because live views and any future
Canva design render the *same* `board` data, they stay in parity by
construction.

Out of scope for now: Canva Connect OAuth, the autofill API call, and template
IDs.
