package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bmayfi3ld/quilt/pkg/engine"
	"github.com/bmayfi3ld/quilt/pkg/ingest"
	"github.com/bmayfi3ld/quilt/pkg/store"
)

// withStore opens the schedule at db, calls fn, and closes the store.
func withStore[Out any](db string, fn func(*store.Store) (Out, error)) (Out, error) {
	s, err := store.Open(db)
	if err != nil {
		var zero Out
		return zero, err
	}
	defer func() { _ = s.Close() }()
	return fn(s)
}

// mutate opens the store at db, runs fn, and closes it. Used by tools that
// have no meaningful structured return value.
func mutate(db string, fn func(*store.Store) error) error {
	s, err := store.Open(db)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	return fn(s)
}

func textResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}

// runImport is the shared body of the import tool handler, extracted to keep
// registerLifecycleTools below the cognitive complexity limit.
func runImport(in ImportInput) (ImportOutput, error) {
	schedSheet := in.ScheduleSheet
	if schedSheet == "" {
		schedSheet = "Schedule"
	}
	rulesSheet := in.RulesSheet
	if rulesSheet == "" {
		rulesSheet = "Rules"
	}
	displayName := in.Name
	if displayName == "" {
		base := filepath.Base(in.DB)
		displayName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	schedule, rules, err := ingest.ReadWorkbook(in.From, schedSheet, rulesSheet)
	if err != nil {
		return ImportOutput{}, fmt.Errorf("reading workbook: %w", err)
	}
	parsed, err := ingest.Parse(schedule, rules)
	if err != nil {
		return ImportOutput{}, fmt.Errorf("parsing workbook: %w", err)
	}
	st, err := store.Create(in.DB, displayName)
	if err != nil {
		return ImportOutput{}, err
	}
	sum, err := ingest.Apply(st, parsed)
	if err != nil {
		_ = st.Close()
		_ = os.Remove(in.DB) //nolint:gosec // path is the caller-provided db file
		return ImportOutput{}, fmt.Errorf("importing: %w", err)
	}
	violations, err := st.Validate()
	_ = st.Close()
	if err != nil {
		return ImportOutput{}, fmt.Errorf("validating: %w", err)
	}

	warnings := make([]string, len(sum.Warnings))
	for i, w := range sum.Warnings {
		warnings[i] = w.Text
	}
	return ImportOutput{
		Name:            displayName,
		Classes:         sum.Classes,
		Timeslots:       sum.Timeslots,
		Cohorts:         sum.Cohorts,
		Assignments:     sum.Assignments,
		TravelRules:     sum.TravelRules,
		Blackouts:       sum.Blackouts,
		OneClassAtATime: sum.OneClassAtATime,
		Warnings:        warnings,
		Violations:      len(violations),
	}, nil
}

// --- Input / Output types ---

type InitInput struct {
	DB   string `json:"db"             jsonschema:"path for the new schedule .db file"`
	Name string `json:"name,omitempty" jsonschema:"display name for the schedule; defaults to the file path"`
}

type InfoInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

type CopyInput struct {
	DB   string `json:"db"             jsonschema:"source schedule .db file"`
	Out  string `json:"out"            jsonschema:"destination .db file to create"`
	Name string `json:"name,omitempty" jsonschema:"name for the new schedule; keeps source name if empty"`
}

type ImportInput struct {
	DB            string `json:"db"                      jsonschema:"destination .db file to create"`
	From          string `json:"from"                    jsonschema:"source .xlsx workbook file"`
	Name          string `json:"name,omitempty"          jsonschema:"schedule name; defaults to the db basename without extension"`
	ScheduleSheet string `json:"scheduleSheet,omitempty" jsonschema:"sheet name for the schedule grid; defaults to Schedule"`
	RulesSheet    string `json:"rulesSheet,omitempty"    jsonschema:"sheet name for rules; defaults to Rules"`
}

type ImportOutput struct {
	Name            string   `json:"name"`
	Classes         int      `json:"classes"`
	Timeslots       int      `json:"timeslots"`
	Cohorts         int      `json:"cohorts"`
	Assignments     int      `json:"assignments"`
	TravelRules     int      `json:"travelRules"`
	Blackouts       int      `json:"blackouts"`
	OneClassAtATime bool     `json:"oneClassAtATime"`
	Warnings        []string `json:"warnings,omitempty"`
	Violations      int      `json:"violations"`
}

type AddClassInput struct {
	DB   string `json:"db"   jsonschema:"path to the schedule .db file"`
	Name string `json:"name" jsonschema:"class name to add"`
}

type AddTimeslotInput struct {
	DB     string `json:"db"              jsonschema:"path to the schedule .db file"`
	Label  string `json:"label"           jsonschema:"timeslot label (e.g. Mon-1)"`
	Day    string `json:"day,omitempty"    jsonschema:"day label (optional)"`
	Period string `json:"period,omitempty" jsonschema:"period label (optional)"`
}

type AddCohortInput struct {
	DB   string `json:"db"   jsonschema:"path to the schedule .db file"`
	Name string `json:"name" jsonschema:"cohort name to add"`
}

type RemoveClassInput struct {
	DB   string `json:"db"   jsonschema:"path to the schedule .db file"`
	Name string `json:"name" jsonschema:"class name to remove"`
}

type RemoveTimeslotInput struct {
	DB    string `json:"db"    jsonschema:"path to the schedule .db file"`
	Label string `json:"label" jsonschema:"timeslot label to remove"`
}

type RemoveCohortInput struct {
	DB   string `json:"db"   jsonschema:"path to the schedule .db file"`
	Name string `json:"name" jsonschema:"cohort name to remove"`
}

type EnableRuleInput struct {
	DB   string `json:"db"   jsonschema:"path to the schedule .db file"`
	Rule string `json:"rule" jsonschema:"rule name; currently supports one-class-at-a-time"`
	On   bool   `json:"on"   jsonschema:"true to enable the rule, false to disable"`
}

type SetTravelInput struct {
	DB        string     `json:"db"        jsonschema:"path to the schedule .db file"`
	Class     string     `json:"class"     jsonschema:"class the travel rule applies to"`
	Buildings [][]string `json:"buildings" jsonschema:"building groups; each inner list is cohort names colocated in that building"`
}

type AddBlackoutInput struct {
	DB       string `json:"db"       jsonschema:"path to the schedule .db file"`
	Cohort   string `json:"cohort"   jsonschema:"cohort name"`
	Timeslot string `json:"timeslot" jsonschema:"timeslot label"`
}

type AssignInput struct {
	DB       string `json:"db"       jsonschema:"path to the schedule .db file"`
	Class    string `json:"class"    jsonschema:"class name"`
	Timeslot string `json:"timeslot" jsonschema:"timeslot label"`
	Cohort   string `json:"cohort"   jsonschema:"cohort to assign (any text, including exempt placeholders like #### closed)"`
}

type UnassignInput struct {
	DB       string `json:"db"       jsonschema:"path to the schedule .db file"`
	Class    string `json:"class"    jsonschema:"class name"`
	Timeslot string `json:"timeslot" jsonschema:"timeslot label"`
}

type ValidateInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

// ValidateOutput wraps violations so the output schema has type object.
type ValidateOutput struct {
	Violations []engine.Violation `json:"violations"`
}

type ListUnassignedInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

// UnassignedOutput wraps empty cells so the output schema has type object.
type UnassignedOutput struct {
	Cells []engine.Cell `json:"cells"`
}

type ReportInput struct {
	DB     string `json:"db"              jsonschema:"path to the schedule .db file"`
	Cohort string `json:"cohort,omitempty" jsonschema:"filter to one cohort; omit or empty string returns all cohorts"`
}

// ReportOutput wraps per-cohort calendar entries.
type ReportOutput struct {
	Entries []store.ReportEntry `json:"entries"`
}

type GridInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

type BoardInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

type ListClassesInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

// StringListOutput wraps a list of names so the output schema has type object.
type StringListOutput struct {
	Items []string `json:"items"`
}

type ListTimeslotsInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

// TimeslotsOutput wraps timeslot metadata.
type TimeslotsOutput struct {
	Timeslots []store.TimeslotInfo `json:"timeslots"`
}

type ListCohortsInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

type ShowRulesInput struct {
	DB string `json:"db" jsonschema:"path to the schedule .db file"`
}

// registerTools installs every scheduler tool on s.
func registerTools(s *mcp.Server) {
	registerLifecycleTools(s)
	registerStructureTools(s)
	registerRuleTools(s)
	registerGridTools(s)
	registerQueryTools(s)
	registerListTools(s)
}

func registerLifecycleTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "init",
		Description: "Create an empty schedule database at the given path.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in InitInput) (*mcp.CallToolResult, any, error) {
		name := in.Name
		if name == "" {
			name = in.DB
		}
		st, err := store.Create(in.DB, name)
		if err != nil {
			return nil, nil, err
		}
		_ = st.Close()
		return textResult(fmt.Sprintf("Created schedule %q at %s", name, in.DB)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "copy",
		Description: "Branch a schedule into a new independent file.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CopyInput) (*mcp.CallToolResult, any, error) {
		if err := store.Copy(in.DB, in.Out, in.Name); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Copied %s → %s", in.DB, in.Out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "import",
		Description: "Import an .xlsx workbook into a new schedule database.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ImportInput) (*mcp.CallToolResult, ImportOutput, error) {
		out, err := runImport(in)
		return nil, out, err
	})

	// --- info ---

	mcp.AddTool(s, &mcp.Tool{
		Name:        "info",
		Description: "Return the schedule name, timestamps, entity counts, and rule status.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in InfoInput) (*mcp.CallToolResult, store.Info, error) {
		info, err := withStore(in.DB, func(s *store.Store) (store.Info, error) {
			return s.Info()
		})
		return nil, info, err
	})
}

func registerStructureTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_class",
		Description: "Add a class (grid row) to the schedule.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in AddClassInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error { return s.AddClass(in.Name) }); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Added class %q", in.Name)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_timeslot",
		Description: "Add a timeslot (grid column) to the schedule.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in AddTimeslotInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error {
			return s.AddTimeslot(in.Label, in.Day, in.Period)
		}); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Added timeslot %q", in.Label)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_cohort",
		Description: "Add a cohort to the AllCohorts master list.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in AddCohortInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error { return s.AddCohort(in.Name) }); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Added cohort %q", in.Name)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_class",
		Description: "Remove a class and all its assignments from the schedule.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in RemoveClassInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error { return s.RemoveClass(in.Name) }); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Removed class %q", in.Name)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_timeslot",
		Description: "Remove a timeslot and all its assignments from the schedule.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in RemoveTimeslotInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error { return s.RemoveTimeslot(in.Label) }); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Removed timeslot %q", in.Label)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_cohort",
		Description: "Remove a cohort from the AllCohorts master list (does not clear assignments).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in RemoveCohortInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error { return s.RemoveCohort(in.Name) }); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Removed cohort %q", in.Name)), nil, nil
	})
}

func registerRuleTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "enable_rule",
		Description: "Enable or disable a validation rule. Supported rule: one-class-at-a-time.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in EnableRuleInput) (*mcp.CallToolResult, any, error) {
		if in.Rule != "one-class-at-a-time" {
			return nil, nil, fmt.Errorf("unknown rule %q; supported: one-class-at-a-time", in.Rule)
		}
		if err := mutate(in.DB, func(s *store.Store) error { return s.SetOneClassAtATime(in.On) }); err != nil {
			return nil, nil, err
		}
		state := "enabled"
		if !in.On {
			state = "disabled"
		}
		return textResult(fmt.Sprintf("Rule %q %s", in.Rule, state)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_travel",
		Description: "Set building travel groups for a class. Buildings is a list of cohort groups; a cohort can only be in one building.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in SetTravelInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error {
			return s.SetTravelGroups(in.Class, in.Buildings)
		}); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Set travel rule for %q across %d building(s)", in.Class, len(in.Buildings))), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_blackout",
		Description: "Forbid a cohort from being scheduled during a timeslot (e.g. lunch, assembly).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in AddBlackoutInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error {
			return s.AddBlackout(in.Cohort, in.Timeslot)
		}); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Blacked out %q during %q", in.Cohort, in.Timeslot)), nil, nil
	})
}

func registerGridTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "assign",
		Description: "Assign a cohort to a (class, timeslot) cell. Re-assigning overwrites the existing value.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in AssignInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error {
			return s.Assign(in.Class, in.Timeslot, in.Cohort)
		}); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Assigned %q to %q during %q", in.Cohort, in.Class, in.Timeslot)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "unassign",
		Description: "Clear a (class, timeslot) cell. Clearing an already-empty cell is not an error.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in UnassignInput) (*mcp.CallToolResult, any, error) {
		if err := mutate(in.DB, func(s *store.Store) error {
			return s.Unassign(in.Class, in.Timeslot)
		}); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Cleared %q during %q", in.Class, in.Timeslot)), nil, nil
	})
}

func registerQueryTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate",
		Description: "Get broken rules: run all validation rules and return every violation. An empty violations list means the schedule is clean.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ValidateInput) (*mcp.CallToolResult, ValidateOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (ValidateOutput, error) {
			violations, err := s.Validate()
			if violations == nil {
				violations = []engine.Violation{}
			}
			return ValidateOutput{Violations: violations}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_unassigned",
		Description: "List unassigned slots: return every (class, timeslot) cell with no cohort assignment.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ListUnassignedInput) (*mcp.CallToolResult, UnassignedOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (UnassignedOutput, error) {
			cells, err := s.ListUnassigned()
			if cells == nil {
				cells = []engine.Cell{}
			}
			return UnassignedOutput{Cells: cells}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "report",
		Description: "Return a per-cohort calendar of all assignments. An empty cohort field returns all cohorts.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ReportInput) (*mcp.CallToolResult, ReportOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (ReportOutput, error) {
			entries, err := s.Report(in.Cohort)
			if entries == nil {
				entries = []store.ReportEntry{}
			}
			return ReportOutput{Entries: entries}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "grid",
		Description: "Return the full master schedule grid (classes × timeslots with cohort values).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in GridInput) (*mcp.CallToolResult, engine.Grid, error) {
		grid, err := withStore(in.DB, func(s *store.Store) (engine.Grid, error) {
			return s.Grid()
		})
		return nil, grid, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "board",
		Description: "Return the authoritative render payload for live views: the full grid, every violation, and timeslot metadata (label/day/period) in one call. Views render this and never recompute validity; re-call board after any assign/unassign.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in BoardInput) (*mcp.CallToolResult, store.BoardData, error) {
		board, err := withStore(in.DB, func(s *store.Store) (store.BoardData, error) {
			return s.Board()
		})
		return nil, board, err
	})
}

func registerListTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_classes",
		Description: "List all class names in display order.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ListClassesInput) (*mcp.CallToolResult, StringListOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (StringListOutput, error) {
			items, err := s.Classes()
			if items == nil {
				items = []string{}
			}
			return StringListOutput{Items: items}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_timeslots",
		Description: "List all timeslots with their label, day, and period in display order.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ListTimeslotsInput) (*mcp.CallToolResult, TimeslotsOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (TimeslotsOutput, error) {
			ts, err := s.Timeslots()
			if ts == nil {
				ts = []store.TimeslotInfo{}
			}
			return TimeslotsOutput{Timeslots: ts}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cohorts",
		Description: "List all cohort names from the AllCohorts master list in display order.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ListCohortsInput) (*mcp.CallToolResult, StringListOutput, error) {
		out, err := withStore(in.DB, func(s *store.Store) (StringListOutput, error) {
			items, err := s.Cohorts()
			if items == nil {
				items = []string{}
			}
			return StringListOutput{Items: items}, err
		})
		return nil, out, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "show_rules",
		Description: "Dump all configured rules: the AllCohorts list, travel groups, blackouts, and OneClassAtATime flag.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ShowRulesInput) (*mcp.CallToolResult, engine.Rules, error) {
		rules, err := withStore(in.DB, func(s *store.Store) (engine.Rules, error) {
			return s.RulesConfig()
		})
		return nil, rules, err
	})
}
