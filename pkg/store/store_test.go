package store

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bmayfi3ld/excel-scheduler/pkg/engine"
)

// newSchedule creates a fresh schedule in a temp dir and registers cleanup.
func newSchedule(t *testing.T, name string) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	s, err := Create(path, "test")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// build adds classes, timeslots, and cohorts and asserts no error along the way.
func build(t *testing.T, s *Store, classes, timeslots, cohorts []string) {
	t.Helper()
	for _, c := range classes {
		if err := s.AddClass(c); err != nil {
			t.Fatalf("AddClass(%q): %v", c, err)
		}
	}
	for _, ts := range timeslots {
		if err := s.AddTimeslot(ts, "", ""); err != nil {
			t.Fatalf("AddTimeslot(%q): %v", ts, err)
		}
	}
	for _, co := range cohorts {
		if err := s.AddCohort(co); err != nil {
			t.Fatalf("AddCohort(%q): %v", co, err)
		}
	}
}

func TestCreateAndOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")

	s, err := Create(path, "my schedule")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_ = s.Close()

	// Re-create must fail (no silent clobber).
	if _, err := Create(path, "again"); err == nil {
		t.Fatal("Create on existing file should fail")
	}

	// Open round-trips the name.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s2.Close() }()
	info, err := s2.Info()
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Name != "my schedule" {
		t.Errorf("name = %q, want %q", info.Name, "my schedule")
	}
	if info.SchemaVersion != SchemaVersion {
		t.Errorf("schema version = %d, want %d", info.SchemaVersion, SchemaVersion)
	}

	// Open on a missing file fails.
	if _, err := Open(filepath.Join(dir, "nope.db")); err == nil {
		t.Fatal("Open on missing file should fail")
	}
}

// TestValidateRules drives each of the four rules to a violation (and a clean
// grid to zero), mirroring the pkg/engine rules.md fixtures through the store.
func TestValidateRules(t *testing.T) {
	classes := []string{"Latin Cart", "Math", "Art"}
	timeslots := []string{"Mon 9am", "Mon 10am", "Mon 11am"}
	cohorts := []string{"1st", "2nd", "3rd", "4th", "5th", "6th"}

	t.Run("cross-building travel", func(t *testing.T) {
		s := newSchedule(t, "travel.db")
		build(t, s, classes, timeslots, cohorts)
		if err := s.SetTravelGroups("Latin Cart", [][]string{{"1st", "2nd", "3rd"}, {"4th", "5th", "6th"}}); err != nil {
			t.Fatal(err)
		}
		mustAssign(t, s, "Latin Cart", "Mon 9am", "1st")
		mustAssign(t, s, "Latin Cart", "Mon 10am", "4th")

		assertOneViolation(t, s, engine.RuleClassRequiresTravel, "Latin Cart", "Mon 10am")
	})

	t.Run("blackout hit", func(t *testing.T) {
		s := newSchedule(t, "blackout.db")
		build(t, s, classes, timeslots, cohorts)
		if err := s.AddBlackout("1st", "Mon 11am"); err != nil {
			t.Fatal(err)
		}
		mustAssign(t, s, "Math", "Mon 11am", "1st")

		assertOneViolation(t, s, engine.RuleCohortBlacklist, "Math", "Mon 11am")
	})

	t.Run("one class at a time double-book", func(t *testing.T) {
		s := newSchedule(t, "oneclass.db")
		build(t, s, classes, timeslots, cohorts)
		if err := s.SetOneClassAtATime(true); err != nil {
			t.Fatal(err)
		}
		mustAssign(t, s, "Math", "Mon 9am", "1st")
		mustAssign(t, s, "Art", "Mon 9am", "1st")

		violations := mustValidate(t, s)
		if len(violations) != 2 {
			t.Fatalf("want 2 violations (one per cell), got %d: %+v", len(violations), violations)
		}
		for _, v := range violations {
			if v.Rule != engine.RuleOneClassAtATime {
				t.Errorf("rule = %q, want OneClassAtATime", v.Rule)
			}
		}
	})

	t.Run("unknown cohort", func(t *testing.T) {
		s := newSchedule(t, "unknown.db")
		build(t, s, classes, timeslots, cohorts)
		mustAssign(t, s, "Math", "Mon 9am", "53rd") // not in master list

		assertOneViolation(t, s, engine.RuleAllCohorts, "Math", "Mon 9am")
	})

	t.Run("exempt placeholder is not flagged", func(t *testing.T) {
		s := newSchedule(t, "exempt.db")
		build(t, s, classes, timeslots, cohorts)
		mustAssign(t, s, "Math", "Mon 9am", "#### closed") // exempt, not a cohort

		if got := mustValidate(t, s); len(got) != 0 {
			t.Fatalf("exempt cell should not be flagged, got %+v", got)
		}
	})

	t.Run("clean grid", func(t *testing.T) {
		s := newSchedule(t, "clean.db")
		build(t, s, classes, timeslots, cohorts)
		if err := s.SetOneClassAtATime(true); err != nil {
			t.Fatal(err)
		}
		mustAssign(t, s, "Math", "Mon 9am", "1st")
		mustAssign(t, s, "Art", "Mon 9am", "2nd")

		if got := mustValidate(t, s); len(got) != 0 {
			t.Fatalf("clean grid should have 0 violations, got %+v", got)
		}
	})
}

// TestLoadRoundTrip asserts the relational tables reconstruct the exact
// Grid/Rules the engine expects.
func TestLoadRoundTrip(t *testing.T) {
	s := newSchedule(t, "rt.db")
	build(t, s,
		[]string{"Latin Cart", "Math"},
		[]string{"Mon 9am", "Mon 10am"},
		[]string{"1st", "2nd", "4th"},
	)
	if err := s.SetOneClassAtATime(true); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTravelGroups("Latin Cart", [][]string{{"1st", "2nd"}, {"4th"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddBlackout("1st", "Mon 10am"); err != nil {
		t.Fatal(err)
	}
	mustAssign(t, s, "Latin Cart", "Mon 9am", "1st")
	mustAssign(t, s, "Math", "Mon 10am", "2nd")

	g, r, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantGrid := engine.Grid{
		Classes:   []string{"Latin Cart", "Math"},
		Timeslots: []string{"Mon 9am", "Mon 10am"},
		Cells: [][]string{
			{"1st", ""},
			{"", "2nd"},
		},
	}
	if !reflect.DeepEqual(g, wantGrid) {
		t.Errorf("grid mismatch:\n got %+v\nwant %+v", g, wantGrid)
	}

	wantRules := engine.Rules{
		AllCohorts:          []string{"1st", "2nd", "4th"},
		ClassRequiresTravel: []engine.TravelRule{{Class: "Latin Cart", Buildings: [][]string{{"1st", "2nd"}, {"4th"}}}},
		CohortBlacklist:     []engine.BlacklistRule{{Cohort: "1st", Timeslots: []string{"Mon 10am"}}},
		OneClassAtATime:     true,
	}
	if !reflect.DeepEqual(r, wantRules) {
		t.Errorf("rules mismatch:\n got %+v\nwant %+v", r, wantRules)
	}
}

// TestAssignOverwriteAndUnassign covers re-assigning and clearing a cell.
func TestAssignOverwriteAndUnassign(t *testing.T) {
	s := newSchedule(t, "edit.db")
	build(t, s, []string{"Math"}, []string{"Mon 9am"}, []string{"1st", "2nd"})

	mustAssign(t, s, "Math", "Mon 9am", "1st")
	mustAssign(t, s, "Math", "Mon 9am", "2nd") // overwrite

	g, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if g.Cells[0][0] != "2nd" {
		t.Errorf("overwrite failed, cell = %q, want 2nd", g.Cells[0][0])
	}

	if err := s.Unassign("Math", "Mon 9am"); err != nil {
		t.Fatal(err)
	}
	unassigned, err := s.ListUnassigned()
	if err != nil {
		t.Fatal(err)
	}
	if len(unassigned) != 1 || unassigned[0] != (engine.Cell{Class: "Math", Timeslot: "Mon 9am"}) {
		t.Errorf("ListUnassigned = %+v, want one Math/Mon 9am cell", unassigned)
	}
}

// TestMissingReferences asserts edits validate that the class/timeslot exist.
func TestMissingReferences(t *testing.T) {
	s := newSchedule(t, "missing.db")
	build(t, s, []string{"Math"}, []string{"Mon 9am"}, nil)

	if err := s.Assign("Nope", "Mon 9am", "1st"); err == nil {
		t.Error("Assign with unknown class should error")
	}
	if err := s.Assign("Math", "Nope", "1st"); err == nil {
		t.Error("Assign with unknown timeslot should error")
	}
	if err := s.RemoveClass("Nope"); err == nil {
		t.Error("RemoveClass on unknown class should error")
	}
}

// TestRemoveClassCascades confirms FK cascade clears the class's assignments.
func TestRemoveClassCascades(t *testing.T) {
	s := newSchedule(t, "cascade.db")
	build(t, s, []string{"Math"}, []string{"Mon 9am"}, []string{"1st"})
	mustAssign(t, s, "Math", "Mon 9am", "1st")

	if err := s.RemoveClass("Math"); err != nil {
		t.Fatal(err)
	}
	g, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Classes) != 0 {
		t.Errorf("classes after remove = %v, want empty", g.Classes)
	}
	// The assignment should have been cascaded away; a fresh Validate is clean.
	if got := mustValidate(t, s); len(got) != 0 {
		t.Errorf("orphaned assignment survived cascade: %+v", got)
	}
}

// TestCopyIndependence is the headline multi-file requirement: a copy is an
// independent file — editing it does not touch the original.
func TestCopyIndependence(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "2026.db")
	dst := filepath.Join(dir, "2027-draft.db")

	s, err := Create(src, "2026-2027")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddClass("Math"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddTimeslot("Mon 9am", "", ""); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()

	if err := Copy(src, dst, "2027 draft"); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	// Copy must refuse an existing destination.
	if err := Copy(src, dst, ""); err == nil {
		t.Error("Copy over existing dst should fail")
	}

	// Edit the copy.
	d, err := Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AddClass("Art"); err != nil {
		t.Fatal(err)
	}
	dInfo, _ := d.Info()
	_ = d.Close()

	// Original is untouched.
	o, err := Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = o.Close() }()
	oInfo, _ := o.Info()

	if oInfo.Classes != 1 {
		t.Errorf("original classes = %d, want 1 (copy edit leaked)", oInfo.Classes)
	}
	if dInfo.Classes != 2 {
		t.Errorf("copy classes = %d, want 2", dInfo.Classes)
	}
	if dInfo.Name != "2027 draft" {
		t.Errorf("copy name = %q, want %q", dInfo.Name, "2027 draft")
	}
	if oInfo.Name != "2026-2027" {
		t.Errorf("original name = %q, want %q", oInfo.Name, "2026-2027")
	}
}

func TestReport(t *testing.T) {
	s := newSchedule(t, "report.db")
	build(t, s, []string{"Math", "Art"}, []string{"Mon 9am", "Mon 10am"}, []string{"1st"})
	mustAssign(t, s, "Math", "Mon 9am", "1st")
	mustAssign(t, s, "Art", "Mon 10am", "1st")

	entries, err := s.Report("1st")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("report entries = %d, want 2: %+v", len(entries), entries)
	}
	// Ordered by timeslot sort_order.
	if entries[0].Timeslot != "Mon 9am" || entries[0].Class != "Math" {
		t.Errorf("entry[0] = %+v, want Math/Mon 9am", entries[0])
	}
}

// --- test helpers ---

func mustAssign(t *testing.T, s *Store, class, timeslot, cohort string) {
	t.Helper()
	if err := s.Assign(class, timeslot, cohort); err != nil {
		t.Fatalf("Assign(%q,%q,%q): %v", class, timeslot, cohort, err)
	}
}

func mustValidate(t *testing.T, s *Store) []engine.Violation {
	t.Helper()
	v, err := s.Validate()
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	return v
}

func assertOneViolation(t *testing.T, s *Store, rule engine.RuleType, class, timeslot string) {
	t.Helper()
	got := mustValidate(t, s)
	if len(got) != 1 {
		t.Fatalf("want 1 violation, got %d: %+v", len(got), got)
	}
	if got[0].Rule != rule {
		t.Errorf("rule = %q, want %q", got[0].Rule, rule)
	}
	wantCell := engine.Cell{Class: class, Timeslot: timeslot}
	if got[0].Cell != wantCell {
		t.Errorf("cell = %+v, want %+v", got[0].Cell, wantCell)
	}
}
