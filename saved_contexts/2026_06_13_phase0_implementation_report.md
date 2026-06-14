# Implementation Report: Phase 0 — Go Rules Engine

**Date:** 2026-06-13
**Author:** Claude
**Plan:** `~/.claude/plans/lazy-petting-creek.md`
**Predecessor:** `saved_contexts/2026_06_13_feasibility_report.md` (the decision to re-platform onto a Go core)

---

## Summary

Phase 0 is **complete**. The four schedule-validation rules from the Office.js
add-in (`addin/src/taskpane/taskpane.js:218-330`) have been re-implemented against
a plain Go domain model, unit-tested with the worked examples already written in
`docs/content/docs/rules.md`. This is the durable, load-bearing IP per the
feasibility report — every later surface (CLI, MCP tools, generated views) will
render the output of this engine and never recompute it.

No SQLite store, no CLI, no `.xlsx` importer, and no MCP server were built; those
are explicitly Phase 1+ and out of scope. `go test` is the driver and the doc
examples are the fixtures, exactly as the plan specified.

**Status:** `go vet ./...` clean, `gofmt -l` clean, all tests pass. No binary is
produced — the test suite is the deliverable and the proof.

---

## What was built

New Go module rooted at the repo root, library code under `pkg/`:

```
go.mod                          module github.com/bmayfi3ld/excel-scheduler  (go 1.23)
pkg/engine/model.go             domain types (85 lines)
pkg/engine/validate.go          Validate() + the four rule checks + helpers (166 lines)
pkg/engine/validate_test.go     table-driven tests from the rules.md fixtures (224 lines)
pkg/engine/exempt.go            isExempt() — the 4-repeating-char quirk (18 lines)
pkg/engine/exempt_test.go       edge cases for the quirk (30 lines)
```

`.gitignore` was extended with a `# Go` section (`bin/`, `*.exe`) — no artifacts
are expected in Phase 0, but it keeps later phases clean.

---

## Domain model (`pkg/engine/model.go`)

The engine's input is a **typed 2D grid with ordered label axes**, identity by
name — the contract every later phase feeds. It mirrors both the Excel source
(`scheduleRange.values`) and the future master-grid UI, and makes the two
index-sensitive rules (`ClassRequiresTravel` needs the prior column,
`OneClassAtATime` needs the column) direct index math rather than rebuilt indices.

| Type | Role |
| --- | --- |
| `Grid{Classes, Timeslots, Cells}` | The schedule. Classes are rows, timeslots are columns (both ordered). `Cells` is row-major: `Cells[r][c]` is the cohort at `(Classes[r], Timeslots[c])`; `""` is unassigned. Must be rectangular. |
| `Cell{Class, Timeslot}` | A position identified by *labels*, so a `Violation` is meaningful without a `Grid` reference. |
| `Rules{AllCohorts, ClassRequiresTravel, CohortBlacklist, OneClassAtATime}` | The full constraint set. |
| `TravelRule{Class, Buildings}` | `Buildings[i]` = cohorts colocated in building `i` for a mobile class. |
| `BlacklistRule{Cohort, Timeslots}` | Timeslots in which `Cohort` may not be scheduled. |
| `RuleType` + 4 constants | Names the rule that produced a `Violation`, matching the doc/add-in wording. |
| `Violation{Cell, Cohort, Rule, Message}` | One broken rule at one cell. Multiple violations per cell are allowed. |

Design decisions carried out from the plan:
- **Fixed struct + per-rule functions** (`checkAllCohorts`, `checkTravel`,
  `checkBlacklist`, `checkOneAtATime`), not a `Rule` interface — matches the
  exactly-four-rules reality; promoting to an interface later is mechanical.
- The engine **only computes** `[]Violation`. It does not paint cells or attach
  comments (that's a renderer's job), unlike the add-in which mutated the sheet
  inline.
- Message strings are kept close to the add-in's wording (`taskpane.js:223-325`)
  so a future importer/diff against the add-in stays comparable.
- **Determinism guardrail:** the grid is walked row-major (classes outer,
  timeslots inner), so `[]Violation` ordering is stable for tests and callers.

---

## Validation logic (`pkg/engine/validate.go`)

`Validate(g Grid, r Rules) []Violation` walks every assigned cell, skips empty
(`""`) and rule-exempt cells, then runs the four checks and accumulates
violations. Per-cell semantics mirror the JS exactly:

1. **AllCohorts** — cohort not in `r.AllCohorts` → violation.
   *Caveat documented in code:* an **empty** `AllCohorts` flags every assigned
   cohort rather than disabling the check, so it must be populated whenever the
   grid has assignments. (The other slice-based rules are genuinely off when empty.)
2. **ClassRequiresTravel** — find the `TravelRule` for this class; if the current
   cohort's building differs from the prior column's cohort's building → violation.
   Guards: only for `col >= 1` (no prior cell at the first timeslot); requires ≥2
   buildings; if either cohort isn't placed in any building, no judgment is made.
3. **CohortBlacklist** — if a `BlacklistRule` for this cohort lists the current
   timeslot → violation.
4. **OneClassAtATime** (boolean toggle) — if enabled, scan other rows at the same
   column; if the same cohort appears in another class → violation. Each side of a
   conflict is flagged independently (`Validate` calls the check for both cells),
   so a conflicting pair yields **two** violations, matching the JS.

Unexported helpers: `findTravelRule`, `buildingOf` (returns building index or -1),
`contains`.

All exported types/functions and the rule checks carry doc comments; the rule
functions document *what they flag and why*, including the guard conditions.

---

## The exemption quirk (`pkg/engine/exempt.go`)

`isExempt(cohort)` replicates `taskpane.js:191-207` faithfully: a value whose
**first four characters are identical** (`####`, `@@@@ no class`, `**** closed`)
is skipped from *all* rule checks — the convention for blacked-out times holding a
placeholder rather than a real cohort. Values shorter than four characters are
never exempt. The comparison is byte-wise (matching the add-in's `substring(0,4)`),
so multi-byte runes are not special-cased.

---

## Tests / fixtures

Encoded directly from the `rules.md` worked examples, table-driven with `t.Run`
subtests asserting both the violation **count** and the `(Rule, Cell)` of each, so
failures point at the exact rule.

| Subtest | Fixture | Expected |
| --- | --- | --- |
| `Travel/cross-building_gives_violation` | `rules.md:122-134` — Latin Cart `1st`→`4th` across adjacent slots | 1 violation at the `4th` cell |
| `Travel/same-building_no_violation` | Lunch Cart `1st`→`2nd` (same building) | 0 |
| `Blacklist/blocked_timeslot_gives_violation` | `rules.md:155-181` — `1st` at `Monday, 11am` | 1 |
| `Blacklist/allowed_timeslot_no_violation` | `1st` at a non-blacklisted slot | 0 |
| `OneClassAtATime/conflict_gives_two_violations` | `rules.md:199-206` — `Math`=1st & `Art`=1st same slot | 2 (one per cell) |
| `OneClassAtATime/disabled_no_violation` | same grid, rule off | 0 |
| `AllCohorts/unknown_cohort_gives_violation` | `rules.md:42-44` — `53rd` not in list | 1 |
| `Exemption/repeating-char_cell_skipped` | `#### blocked` that would otherwise break AllCohorts | 0 |
| `Clean/valid_grid_no_violations` | a fully valid grid, all rules on | 0 |

`exempt_test.go` adds 13 edge cases for the quirk (`####`, `@@@@`, `aaaa` →
exempt; `###`, `###x`, `aaab`, `""` → not).

---

## Verification

```
$ go vet ./...          # clean
$ gofmt -l pkg/engine/  # clean (no files listed)
$ go test ./pkg/engine/...
ok  github.com/bmayfi3ld/excel-scheduler/pkg/engine
```

All 10 `Validate` subtests + the `isExempt` table pass, each tied to a documented
`rules.md` example.

---

## Notable decisions & deviations

- **AllCohorts empty-list behavior** was found to be a sharp edge during the
  doc-comment pass: unlike the other slice rules, an empty `AllCohorts` does *not*
  disable the rule — `contains` returns false for everything, flagging the whole
  grid. This is documented in both `model.go` (the `Rules` struct) and
  `checkAllCohorts`, rather than silently "fixed", to preserve parity with the
  add-in's assumption that the master list is always configured. A future Phase 1
  loader should treat a missing AllCohorts column as "rule absent" before calling
  `Validate`.
- **Blank-delimiter parsing is deliberately absent.** `splitArrayByEmptyStrings`
  (`taskpane.js:497-532`) is an *Excel-input* concern; the engine operates on the
  already-parsed typed model. It belongs to the Phase 1 `.xlsx` importer.

---

## Out of scope (later phases, unchanged from the plan)

- SQLite store, file open/save (Phase 1)
- CLI commands: `validate`, `assign`, `list-unassigned`, `report` (Phase 1)
- `.xlsx` importer reusing the blank-delimiter parsing (Phase 1 subcommand)
- MCP server wrapping the same functions (Phase 2)
- Master grid UI + Live-Artifact autogeneration (Phase 3)

The relational→`Grid` transform (Phase 1's loader) and the day/period
decomposition of timeslots (a store/view concern) are intentionally *not* the
engine's job — the four rules only need order + label, which `Grid` already
provides.
