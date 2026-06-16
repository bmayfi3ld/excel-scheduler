package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bmayfi3ld/excel-scheduler/pkg/engine"
)

// This file holds the shared operation surface. Each method returns structured
// data (engine types or the DTOs below), never formatted text, so the CLI and
// the Phase 2 MCP server can wrap the same function.

// TimeslotInfo describes a grid column for read commands.
type TimeslotInfo struct {
	Label  string `json:"label"`
	Day    string `json:"day,omitempty"`
	Period string `json:"period,omitempty"`
}

// ReportEntry is one placement in a per-cohort calendar.
type ReportEntry struct {
	Cohort   string `json:"cohort"`
	Class    string `json:"class"`
	Timeslot string `json:"timeslot"`
	Day      string `json:"day,omitempty"`
	Period   string `json:"period,omitempty"`
}

// Info is the self-describing summary of a schedule file.
type Info struct {
	Name            string `json:"name"`
	SchemaVersion   int    `json:"schemaVersion"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	Classes         int    `json:"classes"`
	Timeslots       int    `json:"timeslots"`
	Cohorts         int    `json:"cohorts"`
	Assignments     int    `json:"assignments"`
	OneClassAtATime bool   `json:"oneClassAtATime"`
	TravelRules     int    `json:"travelRules"`
	Blackouts       int    `json:"blackouts"`
}

// --- Structural builders (auto-append sort_order) ---

// AddClass appends a new grid row. Errors on a duplicate name.
func (s *Store) AddClass(name string) error {
	var next int
	if err := s.queryRow(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM class`).Scan(&next); err != nil {
		return err
	}
	if _, err := s.exec(`INSERT INTO class (name, sort_order) VALUES (?, ?)`, name, next); err != nil {
		return fmt.Errorf("adding class %q: %w", name, err)
	}
	return s.touch()
}

// AddCohort appends a name to the AllCohorts master list. Errors on a duplicate.
func (s *Store) AddCohort(name string) error {
	var next int
	if err := s.queryRow(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM cohort`).Scan(&next); err != nil {
		return err
	}
	if _, err := s.exec(`INSERT INTO cohort (name, sort_order) VALUES (?, ?)`, name, next); err != nil {
		return fmt.Errorf("adding cohort %q: %w", name, err)
	}
	return s.touch()
}

// AddTimeslot appends a new grid column. day and period are optional metadata.
func (s *Store) AddTimeslot(label, day, period string) error {
	var next int
	if err := s.queryRow(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM timeslot`).Scan(&next); err != nil {
		return err
	}
	if _, err := s.exec(
		`INSERT INTO timeslot (label, day, period, sort_order) VALUES (?, ?, ?, ?)`,
		label, nullable(day), nullable(period), next,
	); err != nil {
		return fmt.Errorf("adding timeslot %q: %w", label, err)
	}
	return s.touch()
}

// RemoveClass deletes a row; cascades to its assignments and travel groups.
func (s *Store) RemoveClass(name string) error {
	return s.deleteWhere(`DELETE FROM class WHERE name = ?`, name, "class", name)
}

// RemoveTimeslot deletes a column; cascades to its assignments and blackouts.
func (s *Store) RemoveTimeslot(label string) error {
	return s.deleteWhere(`DELETE FROM timeslot WHERE label = ?`, label, "timeslot", label)
}

// RemoveCohort drops a name from the AllCohorts master list. It deliberately
// does not touch assignments, since cohort_value is free text — afterward those
// cells become AllCohorts violations, which is the intended signal.
func (s *Store) RemoveCohort(name string) error {
	return s.deleteWhere(`DELETE FROM cohort WHERE name = ?`, name, "cohort", name)
}

// deleteWhere runs a single-arg delete and errors if nothing matched.
func (s *Store) deleteWhere(query, arg, kind, label string) error {
	res, err := s.exec(query, arg)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%s %q not found", kind, label)
	}
	return s.touch()
}

// --- Rule configuration ---

// SetOneClassAtATime toggles the OneClassAtATime rule.
func (s *Store) SetOneClassAtATime(on bool) error {
	v := 0
	if on {
		v = 1
	}
	if _, err := s.exec(`UPDATE meta SET one_class_at_a_time = ? WHERE id = 1`, v); err != nil {
		return err
	}
	return s.touch()
}

// SetTravelGroups replaces the building groupings for a class. buildings[i] is
// the set of cohorts colocated in building i. The class must already exist.
func (s *Store) SetTravelGroups(class string, buildings [][]string) error {
	classID, err := s.classID(class)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tx, err := s.begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM travel_group WHERE class_id = ?`, classID); err != nil {
		return err
	}
	for i, building := range buildings {
		for _, cohort := range building {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO travel_group (class_id, building_index, cohort_name) VALUES (?, ?, ?)`,
				classID, i, cohort,
			); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return s.touch()
}

// AddBlackout forbids a cohort from being scheduled during a timeslot. The
// timeslot must exist; the cohort name is free text matched by the engine.
func (s *Store) AddBlackout(cohort, timeslot string) error {
	timeslotID, err := s.timeslotID(timeslot)
	if err != nil {
		return err
	}
	if _, err := s.exec(
		`INSERT INTO blackout (cohort_name, timeslot_id) VALUES (?, ?)`,
		cohort, timeslotID,
	); err != nil {
		return fmt.Errorf("adding blackout for %q at %q: %w", cohort, timeslot, err)
	}
	return s.touch()
}

// --- Grid edits ---

// Assign fills a cell. The class and timeslot must exist; cohort accepts any
// text, including exempt placeholders like "#### closed". Re-assigning a filled
// cell overwrites it.
func (s *Store) Assign(class, timeslot, cohort string) error {
	classID, err := s.classID(class)
	if err != nil {
		return err
	}
	timeslotID, err := s.timeslotID(timeslot)
	if err != nil {
		return err
	}
	if _, err := s.exec(
		`INSERT INTO assignment (class_id, timeslot_id, cohort_value) VALUES (?, ?, ?)
		 ON CONFLICT(class_id, timeslot_id) DO UPDATE SET cohort_value = excluded.cohort_value`,
		classID, timeslotID, cohort,
	); err != nil {
		return err
	}
	return s.touch()
}

// Unassign clears a cell. The class and timeslot must exist; clearing an
// already-empty cell is not an error.
func (s *Store) Unassign(class, timeslot string) error {
	classID, err := s.classID(class)
	if err != nil {
		return err
	}
	timeslotID, err := s.timeslotID(timeslot)
	if err != nil {
		return err
	}
	if _, err := s.exec(
		`DELETE FROM assignment WHERE class_id = ? AND timeslot_id = ?`,
		classID, timeslotID,
	); err != nil {
		return err
	}
	return s.touch()
}

// --- Queries ---

// Validate loads the grid and runs the engine, returning every violation.
func (s *Store) Validate() ([]engine.Violation, error) {
	g, r, err := s.Load()
	if err != nil {
		return nil, err
	}
	return engine.Validate(g, r), nil
}

// ListUnassigned returns every empty cell (class × timeslot with no assignment).
func (s *Store) ListUnassigned() ([]engine.Cell, error) {
	g, _, err := s.Load()
	if err != nil {
		return nil, err
	}
	var out []engine.Cell
	for row, class := range g.Classes {
		for col, timeslot := range g.Timeslots {
			if g.Cells[row][col] == "" {
				out = append(out, engine.Cell{Class: class, Timeslot: timeslot})
			}
		}
	}
	return out, nil
}

// Report returns a per-cohort calendar. An empty cohort reports all cohorts.
func (s *Store) Report(cohort string) ([]ReportEntry, error) {
	query := `SELECT a.cohort_value, c.name, t.label, t.day, t.period
		 FROM assignment a
		 JOIN class c ON c.id = a.class_id
		 JOIN timeslot t ON t.id = a.timeslot_id`
	args := []any{}
	if cohort != "" {
		query += ` WHERE a.cohort_value = ?`
		args = append(args, cohort)
	}
	query += ` ORDER BY a.cohort_value, t.sort_order, t.id, c.sort_order, c.id`

	rows, err := s.query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []ReportEntry
	for rows.Next() {
		var e ReportEntry
		var day, period sql.NullString
		if err := rows.Scan(&e.Cohort, &e.Class, &e.Timeslot, &day, &period); err != nil {
			return nil, err
		}
		e.Day = day.String
		e.Period = period.String
		out = append(out, e)
	}
	return out, rows.Err()
}

// Info reports the schedule's metadata and counts — the self-describing view.
func (s *Store) Info() (Info, error) {
	var info Info
	err := s.queryRow(
		`SELECT name, schema_version, one_class_at_a_time, created_at, updated_at FROM meta WHERE id = 1`,
	).Scan(&info.Name, &info.SchemaVersion, &info.OneClassAtATime, &info.CreatedAt, &info.UpdatedAt)
	if err != nil {
		return info, err
	}

	counts := []struct {
		query string
		dst   *int
	}{
		{`SELECT COUNT(*) FROM class`, &info.Classes},
		{`SELECT COUNT(*) FROM timeslot`, &info.Timeslots},
		{`SELECT COUNT(*) FROM cohort`, &info.Cohorts},
		{`SELECT COUNT(*) FROM assignment`, &info.Assignments},
		{`SELECT COUNT(DISTINCT class_id) FROM travel_group`, &info.TravelRules},
		{`SELECT COUNT(*) FROM blackout`, &info.Blackouts},
	}
	for _, c := range counts {
		if err := s.queryRow(c.query).Scan(c.dst); err != nil { //nolint:gosec // queries are internal constants
			return info, err
		}
	}
	return info, nil
}

// Grid returns the full master schedule (classes × timeslots with cells).
func (s *Store) Grid() (engine.Grid, error) {
	g, _, err := s.Load()
	return g, err
}

// RulesConfig returns the configured Rules for display.
func (s *Store) RulesConfig() (engine.Rules, error) {
	_, r, err := s.Load()
	return r, err
}

// Classes returns the grid row labels in display order.
func (s *Store) Classes() ([]string, error) {
	names, _, err := s.loadAxis(`SELECT id, name FROM class ORDER BY sort_order, id`)
	return names, err
}

// Cohorts returns the AllCohorts master list in display order.
func (s *Store) Cohorts() ([]string, error) {
	return s.cohortNames()
}

// Timeslots returns the grid columns (label + optional day/period) in order.
func (s *Store) Timeslots() ([]TimeslotInfo, error) {
	rows, err := s.query(`SELECT label, day, period FROM timeslot ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []TimeslotInfo
	for rows.Next() {
		var t TimeslotInfo
		var day, period sql.NullString
		if err := rows.Scan(&t.Label, &day, &period); err != nil {
			return nil, err
		}
		t.Day = day.String
		t.Period = period.String
		out = append(out, t)
	}
	return out, rows.Err()
}

// --- helpers ---

// classID resolves a class name to its id, with a friendly not-found error.
func (s *Store) classID(name string) (int64, error) {
	return s.lookupID(`SELECT id FROM class WHERE name = ?`, name, "class")
}

// timeslotID resolves a timeslot label to its id.
func (s *Store) timeslotID(label string) (int64, error) {
	return s.lookupID(`SELECT id FROM timeslot WHERE label = ?`, label, "timeslot")
}

func (s *Store) lookupID(query, arg, kind string) (int64, error) {
	var id int64
	err := s.queryRow(query, arg).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%s %q not found", kind, arg)
	}
	return id, err
}

// nullable maps an empty string to a SQL NULL so optional columns stay clean.
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
