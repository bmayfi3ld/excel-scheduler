package mcpserver_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bmayfi3ld/quilt/pkg/engine"
	"github.com/bmayfi3ld/quilt/pkg/mcpserver"
	"github.com/bmayfi3ld/quilt/pkg/store"
)

// expectedTools lists every tool name the server must expose.
var expectedTools = []string{
	"init", "info", "copy", "import",
	"add_class", "add_timeslot", "add_cohort",
	"remove_class", "remove_timeslot", "remove_cohort",
	"enable_rule", "set_travel", "add_blackout",
	"assign", "unassign",
	"validate", "list_unassigned", "report",
	"grid", "board", "list_classes", "list_timeslots", "list_cohorts", "show_rules",
}

// connect starts an in-memory server/client pair and returns the client
// session. The server and session are cleaned up when the test ends.
func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	srv := mcpserver.New("test")
	srvT, cliT := mcp.NewInMemoryTransports()

	ss, err := srv.Connect(ctx, srvT, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(ctx, cliT, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// call invokes a tool and fails the test on protocol or tool errors.
func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	if res.IsError {
		t.Fatalf("CallTool(%s) returned tool error: %v", name, res.Content)
	}
	return res
}

// unmarshalOut re-marshals the StructuredContent from the result into dst.
func unmarshalOut(t *testing.T, res *mcp.CallToolResult, dst any) {
	t.Helper()
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
}

// TestRegistration asserts every expected tool name is present.
func TestRegistration(t *testing.T) {
	cs := connect(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	have := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		have[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !have[name] {
			t.Errorf("missing tool %q", name)
		}
	}
	if t.Failed() {
		t.Logf("registered tools: %v", res.Tools)
	}
}

// TestSelfDocumenting asserts the server advertises instructions and exposes
// the guide:// / view:// resources and the workflow prompts.
func TestSelfDocumenting(t *testing.T) {
	cs := connect(t)
	ctx := context.Background()

	if got := cs.InitializeResult().Instructions; got == "" {
		t.Error("expected non-empty server instructions")
	}

	resReq, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}
	wantRes := []string{"guide://db-model", "guide://migrate-xlsx", "guide://live-views", "guide://tools", "view://master", "view://class"}
	haveRes := map[string]bool{}
	for _, r := range resReq.Resources {
		haveRes[r.URI] = true
	}
	for _, uri := range wantRes {
		if !haveRes[uri] {
			t.Errorf("missing resource %q", uri)
		}
	}

	// A resource must actually read back non-empty content.
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "guide://live-views"})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(rr.Contents) == 0 || rr.Contents[0].Text == "" {
		t.Error("guide://live-views read back empty")
	}

	promptReq, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	wantPrompts := map[string]bool{"live_view": false, "migrate_xlsx": false, "new_schedule": false}
	for _, p := range promptReq.Prompts {
		if _, ok := wantPrompts[p.Name]; ok {
			wantPrompts[p.Name] = true
		}
	}
	for name, found := range wantPrompts {
		if !found {
			t.Errorf("missing prompt %q", name)
		}
	}

	// live_view must return the embedded skeleton.
	gp, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "live_view",
		Arguments: map[string]string{"kind": "master", "db": "x.db"},
	})
	if err != nil {
		t.Fatalf("GetPrompt(live_view): %v", err)
	}
	if len(gp.Messages) == 0 {
		t.Fatal("live_view returned no messages")
	}
}

// TestRoundTrip drives the headline loop:
//  1. init → add_class × 2 / add_timeslot × 2 / add_cohort → assign → validate (0 violations)
//  2. enable_rule (one-class-at-a-time) → assign same cohort to a second class in the same slot
//     → validate (1 violation of type OneClassAtATime)
func TestRoundTrip(t *testing.T) {
	cs := connect(t)
	db := filepath.Join(t.TempDir(), "test.db")

	// 1. Build a minimal schedule.
	call(t, cs, "init", map[string]any{"db": db, "name": "Round-trip test"})
	call(t, cs, "add_class", map[string]any{"db": db, "name": "Latin Cart"})
	call(t, cs, "add_class", map[string]any{"db": db, "name": "Math"})
	call(t, cs, "add_timeslot", map[string]any{"db": db, "label": "Mon 9am"})
	call(t, cs, "add_timeslot", map[string]any{"db": db, "label": "Mon 10am"})
	call(t, cs, "add_cohort", map[string]any{"db": db, "name": "1st"})
	call(t, cs, "assign", map[string]any{"db": db, "class": "Latin Cart", "timeslot": "Mon 9am", "cohort": "1st"})

	// validate should return 0 violations (1st is in AllCohorts, no rules violated).
	var v0 mcpserver.ValidateOutput
	unmarshalOut(t, call(t, cs, "validate", map[string]any{"db": db}), &v0)
	if len(v0.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(v0.Violations), v0.Violations)
	}

	// 2. Enable one-class-at-a-time, then put the same cohort in two classes at
	// the same timeslot — that must produce exactly one violation.
	call(t, cs, "enable_rule", map[string]any{"db": db, "rule": "one-class-at-a-time", "on": true})
	call(t, cs, "assign", map[string]any{"db": db, "class": "Math", "timeslot": "Mon 9am", "cohort": "1st"})

	var v1 mcpserver.ValidateOutput
	unmarshalOut(t, call(t, cs, "validate", map[string]any{"db": db}), &v1)
	if len(v1.Violations) == 0 {
		t.Fatal("expected violations after same cohort assigned to two classes, got 0")
	}
	if v1.Violations[0].Rule != engine.RuleOneClassAtATime {
		t.Errorf("expected RuleOneClassAtATime violation, got %q", v1.Violations[0].Rule)
	}
}

// TestBoardParity asserts the board tool's payload equals the separate
// grid/validate/list_timeslots tools — board is pure composition, not new logic.
func TestBoardParity(t *testing.T) {
	cs := connect(t)
	db := filepath.Join(t.TempDir(), "board.db")

	call(t, cs, "init", map[string]any{"db": db, "name": "Board test"})
	call(t, cs, "add_class", map[string]any{"db": db, "name": "Latin Cart"})
	call(t, cs, "add_class", map[string]any{"db": db, "name": "Math"})
	call(t, cs, "add_timeslot", map[string]any{"db": db, "label": "Mon, P1", "day": "Monday", "period": "P1"})
	call(t, cs, "add_cohort", map[string]any{"db": db, "name": "1st"})
	call(t, cs, "enable_rule", map[string]any{"db": db, "rule": "one-class-at-a-time", "on": true})
	call(t, cs, "assign", map[string]any{"db": db, "class": "Latin Cart", "timeslot": "Mon, P1", "cohort": "1st"})
	call(t, cs, "assign", map[string]any{"db": db, "class": "Math", "timeslot": "Mon, P1", "cohort": "1st"})

	var board store.BoardData
	unmarshalOut(t, call(t, cs, "board", map[string]any{"db": db}), &board)

	var grid engine.Grid
	unmarshalOut(t, call(t, cs, "grid", map[string]any{"db": db}), &grid)
	var val mcpserver.ValidateOutput
	unmarshalOut(t, call(t, cs, "validate", map[string]any{"db": db}), &val)
	var ts mcpserver.TimeslotsOutput
	unmarshalOut(t, call(t, cs, "list_timeslots", map[string]any{"db": db}), &ts)

	if !reflect.DeepEqual(board.Grid, grid) {
		t.Errorf("board.Grid != grid tool:\n%+v\n%+v", board.Grid, grid)
	}
	if !reflect.DeepEqual(board.Violations, val.Violations) {
		t.Errorf("board.Violations != validate tool:\n%+v\n%+v", board.Violations, val.Violations)
	}
	if !reflect.DeepEqual(board.Timeslots, ts.Timeslots) {
		t.Errorf("board.Timeslots != list_timeslots tool:\n%+v\n%+v", board.Timeslots, ts.Timeslots)
	}
	if len(board.Violations) == 0 {
		t.Error("expected a OneClassAtATime violation in board payload")
	}
}

// TestToolErrors verifies that tool errors are surfaced as IsError results,
// not as protocol errors, so the client can inspect them.
func TestToolErrors(t *testing.T) {
	cs := connect(t)
	db := filepath.Join(t.TempDir(), "err.db")

	// init creates the file; a second init on the same path must tool-error.
	call(t, cs, "init", map[string]any{"db": db})

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "init",
		Arguments: map[string]any{"db": db},
	})
	if err != nil {
		t.Fatalf("CallTool(init duplicate): unexpected protocol error: %v", err)
	}
	if !res.IsError {
		t.Error("second init on existing file should return IsError=true")
	}

	// info on a non-existent file must tool-error.
	res2, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "info",
		Arguments: map[string]any{"db": "/tmp/does-not-exist-mcptest.db"},
	})
	if err != nil {
		t.Fatalf("CallTool(info missing): unexpected protocol error: %v", err)
	}
	if !res2.IsError {
		t.Error("info on missing file should return IsError=true")
	}
}
