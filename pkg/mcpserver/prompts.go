package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPrompts installs the workflow prompts: ready-to-run instruction text
// that teaches the model how to drive a common task end to end.
func registerPrompts(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "live_view",
		Description: "Build a self-refreshing live HTML view of a schedule (master grid or per-class calendar).",
		Arguments: []*mcp.PromptArgument{
			{Name: "kind", Description: "master or class", Required: true},
			{Name: "db", Description: "path to the .db schedule file", Required: true},
			{Name: "target", Description: "where to write the view (optional)"},
		},
	}, liveViewPrompt)

	s.AddPrompt(&mcp.Prompt{
		Name:        "migrate_xlsx",
		Description: "Import a legacy Excel workbook into a new .db schedule.",
		Arguments: []*mcp.PromptArgument{
			{Name: "xlsx_path", Description: "source .xlsx workbook", Required: true},
			{Name: "db_path", Description: "destination .db file to create", Required: true},
			{Name: "name", Description: "display name for the schedule (optional)"},
		},
	}, migrateXlsxPrompt)

	s.AddPrompt(&mcp.Prompt{
		Name:        "new_schedule",
		Description: "Create and populate a fresh schedule from scratch.",
		Arguments: []*mcp.PromptArgument{
			{Name: "db_path", Description: "path for the new .db file", Required: true},
			{Name: "name", Description: "display name for the schedule (optional)"},
		},
	}, newSchedulePrompt)
}

func arg(req *mcp.GetPromptRequest, key string) string {
	if req.Params == nil || req.Params.Arguments == nil {
		return ""
	}
	return req.Params.Arguments[key]
}

func textPrompt(desc, text string) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: desc,
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: text},
		}},
	}, nil
}

func liveViewPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	kind := arg(req, "kind")
	if kind == "" {
		kind = "master"
	}
	db := arg(req, "db")
	target := arg(req, "target")
	if target == "" {
		target = kind + ".html"
	}
	viewFile := "views/master.html"
	if kind == "class" {
		viewFile = "views/class.html"
	}
	skeleton := mustRead(viewFile)

	text := fmt.Sprintf(`Build a %s live view for the schedule at %q and save it to %q.

Start from the skeleton below. It already reads the "board" tool `+
		"(`mcp__quilt__board`)"+`, builds day/period buckets from the timeslot metadata, `+
		`paints all four rule types, and refreshes every 30s. Set the DB constant (or `+
		"`window.QUILT_DB`"+`) to %q. Do not recompute validity in JS — render the violations `+
		`the board tool returns. See guide://live-views for the full cookbook.

--- skeleton (%s) ---
%s`, kind, db, target, db, viewFile, skeleton)

	return textPrompt(fmt.Sprintf("Live %s view scaffold", kind), text)
}

func migrateXlsxPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	xlsx := arg(req, "xlsx_path")
	db := arg(req, "db_path")
	name := arg(req, "name")
	text := fmt.Sprintf(`Import the Excel workbook %q into a new schedule at %q%s.

1. Call the "import" tool: {"from": %q, "db": %q%s}.
2. Report the summary (classes/timeslots/cohorts/assignments) and any warnings — `+
		`comma-mismatch timeslot warnings are expected and non-fatal.
3. Call "validate" and summarize the resulting violations.

See guide://migrate-xlsx for the blank-delimiter convention and expected warnings.`,
		xlsx, db, nameClause(name), xlsx, db, nameArg(name))
	return textPrompt("Migrate an .xlsx workbook", text)
}

func newSchedulePrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	db := arg(req, "db_path")
	name := arg(req, "name")
	text := fmt.Sprintf(`Create a new schedule at %q%s.

1. Call "init": {"db": %q%s}.
2. Add classes (add_class), timeslots (add_timeslot, with day/period metadata so views `+
		`can build calendars), and cohorts (add_cohort).
3. Configure rules: enable_rule (one-class-at-a-time), set_travel, add_blackout as needed.
4. Fill cells with assign, then call "board" to render and check violations.

See guide://db-model for the model and lifecycle.`, db, nameClause(name), db, nameArg(name))
	return textPrompt("Create a new schedule", text)
}

func nameClause(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf(" named %q", name)
}

func nameArg(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf(", %q: %q", "name", name)
}
