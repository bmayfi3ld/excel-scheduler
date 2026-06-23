# Phase 3 Implementation Report — Self-Documenting MCP, Live Views, Packaging
*Date: 2026-06-23*

Supersedes the as-planned `2026_06_22_phase3_live_views_plan.md`. Part of the
series: `2026_06_13_feasibility_report.md` → phase0 → phase1 →
`2026_06_15_phase2_implementation_report.md` → this.

## Summary

Phase 3 is complete. The `quilt` MCP server now carries its own operating
instructions, exposes the schedule as a single authoritative `board` payload,
ships reusable master + per-class live views, can serve over HTTP, and has a
full release/packaging path (goreleaser, Claude Desktop Extension, Claude Code
plugin). The docs were re-anchored on the Quilt Desktop plugin + `.db`/MCP
workflow, with the Excel add-in marked legacy.

`just validate` is clean: `go test ./...` (50 passing), `go vet`, golangci-lint.
govulncheck reports only transitive advisories with no reachable call paths
(unchanged from before this phase).

---

## Deliverables

### `board` tool (build step 1)
- `store.BoardData{Grid, Violations, Timeslots}` + `Store.Board()` in
  `pkg/store/ops.go` — pure composition of `Load`/`engine.Validate`/`Timeslots`.
- `board` MCP tool in `pkg/mcpserver/tools.go`.
- `TestBoardParity` asserts board == separate grid/validate/list_timeslots.

### Self-documenting MCP (build step 2)
- `pkg/mcpserver/docserver.go`: `go:embed guides/*.md views/*.html`, the
  always-on `Instructions` string, `guide://` resources (`db-model`,
  `migrate-xlsx`, `live-views`, `tools`) and `view://` resources (`master`,
  `class`).
- `pkg/mcpserver/prompts.go`: `live_view`, `migrate_xlsx`, `new_schedule`
  prompts. `live_view` returns the matching embedded HTML skeleton.
- `server.go` now passes `&mcp.ServerOptions{Instructions: …}` (was `nil`) and
  calls `registerDocs`.
- `TestSelfDocumenting` covers instructions, resource list + read-back, prompt
  list + get.

### Live views (build steps 4–5)
- `packaging/live-views/master.html` — generalized from the old `index.html`:
  day tabs + period columns built from `board.timeslots` metadata (no `*8`
  index math), paints all four rule types, 30s + manual refresh.
- `packaging/live-views/class.html` — per-class Day × Period calendar (one grid
  row), class picker, violations overlaid.
- Embedded copies live at `pkg/mcpserver/views/` (see Deviations).

### HTTP transport (build step 6)
- `mcpserver.RunStdio` / `mcpserver.RunHTTP(ctx, version, addr)` using the
  SDK's `NewStreamableHTTPHandler`. `cmd/quilt/mcp.go` gained `--http <addr>`;
  stdio stays the default.

### Canva contract (build step 7) — designed, not built
- Documented in `docs/content/docs/live-views.md` and `guide://live-views`: the
  pure `board → autofill rows` reshape (one row per class: `class_name`,
  `mon_p1`…`fri_pN`). No OAuth / API / template IDs.

### Packaging (build step 8)
- `cmd/quilt/main.go`: `version` var (`-X main.version`) + kong `--version`.
- `.goreleaser.yaml` (CGO_ENABLED=0; linux/darwin/windows × amd64/arm64; bundles
  readme, changelog, docs pages, live-view HTML).
- `.github/workflows/release.yml` (tag `v*` → `just validate` → goreleaser);
  `just release-snapshot` and `just dxt` recipes.
- `packaging/dxt/manifest.json` + `icon.png`/`icon.svg` (copied from
  `addin/assets/schedule-manager-logo.*`).
- `.claude-plugin/plugin.json` declaring the `quilt` MCP server.

### Docs pass (build step 3)
- New: `db-model.md`, `migrate-xlsx.md`, `live-views.md`, `desktop-plugin.md`.
- Re-anchored: `_index.md` (Quilt-first + logo), `install.md` (deprecation
  hint), `mcp-server.md` (board + resources/prompts/instructions + `--http`,
  `quilt` naming), `setup.md` / `functions.md` (legacy hints), `rules.md`
  (weight only — substance still applies).
- `changelog.md`: Phase 3 Unreleased entry; `scheduler …` → `quilt …` string
  fixes (kept the one historical "old scripts called `scheduler …`" reference).
- Weights reordered: desktop-plugin 5, db-model 15, migrate-xlsx 20, rules 25,
  install 30, setup 35, mcp-server 50, live-views 55, functions 65.

---

## Deviations from the plan

1. **`.mcp.json` quilt block not added.** The edit was blocked by the harness as
   a startup-config self-modification. The `quilt` server is declared in
   `.claude-plugin/plugin.json` instead; adding it to `.mcp.json` is a one-line
   manual step for the maintainer. (Follow-up.)
2. **Embedded docs are copies, not byte-identical to Hugo pages.** `go:embed`
   cannot reach `docs/content/` from `pkg/mcpserver/`, so the MCP resources are
   authored as standalone files in `pkg/mcpserver/guides/` and the Hugo pages
   mirror their substance (with Hugo frontmatter/hints). The two are kept in
   sync by hand; the "identical bytes" goal is approximated, not enforced.
3. **Live-view HTML is duplicated** in `packaging/live-views/` (canonical, for
   the release archive/DXT) and `pkg/mcpserver/views/` (embedded copy for the
   resources/prompts), same reason as #2. Keep them in sync when editing.
4. **DXT `.dxt` not actually zipped/validated in Claude Desktop** here — the
   `just dxt` recipe builds the bundle, but I could not install it in Desktop
   from this environment. goreleaser is not installed locally, so
   `just release-snapshot` / `goreleaser check` were not run.
5. **`cohort.html`** (optional report-based view) was not built — it was marked
   optional and not required by the user's choice.

## Follow-ups

- Add the `quilt` block to `.mcp.json` (manual; harness-blocked here).
- `.dxt` signing/notarization (Apple Developer ID + notarytool in CI; Windows
  SmartScreen) — ship unsigned right-click-open first.
- Canva integration (Connect OAuth + autofill API) — designed, not built.
- Run `goreleaser check` / `just release-snapshot` and a real `.dxt` install to
  validate the packaging end to end.
- Decide whether to collapse the duplicated guide/view files via a build-time
  copy step to guarantee Hugo↔MCP parity.
