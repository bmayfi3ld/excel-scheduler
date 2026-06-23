package main

import (
	"errors"
	"path/filepath"
	"testing"
)

// mustRun runs command args and fails the test if it errors.
func mustRun(t *testing.T, args ...string) {
	t.Helper()
	if err := run(args); err != nil {
		t.Fatalf("run(%v): %v", args, err)
	}
}

// TestCLIEndToEnd drives the dispatcher through a full session: create a
// schedule, build it, assign a cross-building cell, and assert validate signals
// a non-zero exit (errSilent) while a clean grid does not.
func TestCLIEndToEnd(t *testing.T) {
	db := filepath.Join(t.TempDir(), "cli.db")

	mustRun(t, "init", "--db", db, "--name", "CLI test")
	mustRun(t, "add-class", "--db", db, "--name", "Latin Cart")
	mustRun(t, "add-class", "--db", db, "--name", "Math")
	mustRun(t, "add-timeslot", "--db", db, "--label", "Mon 9am")
	mustRun(t, "add-timeslot", "--db", db, "--label", "Mon 10am")
	mustRun(t, "add-cohort", "--db", db, "--name", "1st")
	mustRun(t, "add-cohort", "--db", db, "--name", "4th")
	mustRun(t, "set-travel", "--db", db, "--class", "Latin Cart", "--building", "1st,2nd,3rd", "--building", "4th,5th,6th")

	// A clean grid validates without signaling failure.
	mustRun(t, "assign", "--db", db, "--class", "Latin Cart", "--timeslot", "Mon 9am", "--cohort", "1st")
	if err := run([]string{"validate", "--db", db}); err != nil {
		t.Fatalf("validate on clean grid: %v", err)
	}

	// Introduce a cross-building violation; validate must return errSilent.
	mustRun(t, "assign", "--db", db, "--class", "Latin Cart", "--timeslot", "Mon 10am", "--cohort", "4th")
	err := run([]string{"validate", "--db", db})
	if !errors.Is(err, errSilent) {
		t.Fatalf("validate with violations: got %v, want errSilent", err)
	}

	// --json read path should not error.
	mustRun(t, "grid", "--db", db, "--json")
	mustRun(t, "show-rules", "--db", db, "--json")
	mustRun(t, "report", "--db", db, "--json")
	mustRun(t, "list-unassigned", "--db", db)
}

func TestCLIErrors(t *testing.T) {
	db := filepath.Join(t.TempDir(), "err.db")
	mustRun(t, "init", "--db", db)

	tests := []struct {
		name string
		args []string
	}{
		{"unknown command", []string{"frobnicate"}},
		{"missing file arg", []string{"info"}},
		{"add-class without name", []string{"add-class", "--db", db}},
		{"assign unknown class", []string{"assign", "--db", db, "--class", "X", "--timeslot", "Y", "--cohort", "Z"}},
		{"enable unknown rule", []string{"enable-rule", "--db", db, "made-up-rule"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args); err == nil {
				t.Errorf("run(%v) = nil, want error", tt.args)
			}
		})
	}
}

// TestCLICopy verifies copy produces an independent file via the CLI surface.
func TestCLICopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.db")
	dst := filepath.Join(dir, "dst.db")

	mustRun(t, "init", "--db", src, "--name", "orig")
	mustRun(t, "add-class", "--db", src, "--name", "Math")
	mustRun(t, "copy", "--db", src, "--out", dst, "--name", "branch")

	// Edit the copy; the original must be unaffected (checked in store tests,
	// here we just confirm the commands succeed and copy refuses an existing dst).
	mustRun(t, "add-class", "--db", dst, "--name", "Art")
	if err := run([]string{"copy", "--db", src, "--out", dst}); err == nil {
		t.Error("copy over existing dst should error")
	}
}
