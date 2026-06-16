package ingest

import (
	"fmt"
	"strings"

	"github.com/bmayfi3ld/excel-scheduler/pkg/store"
)

// Warning records a non-fatal import issue (duplicate name, referential miss).
type Warning struct {
	Text string
}

// Summary is the result of Apply: counts of what was added and any warnings.
type Summary struct {
	Classes         int
	Timeslots       int
	Cohorts         int
	Assignments     int
	TravelRules     int
	Blackouts       int
	OneClassAtATime bool
	Warnings        []Warning
}

// Apply writes parsed schedule data to s in dependency order.
// Duplicate names and referential misses (unknown class/timeslot) become
// Warnings; genuine I/O or DB errors abort and are returned as an error.
func Apply(s *store.Store, p Parsed) (Summary, error) {
	var sum Summary
	if err := applyEntities(s, p, &sum); err != nil {
		return sum, err
	}
	if err := applyRules(s, p, &sum); err != nil {
		return sum, err
	}
	return sum, nil
}

// applyEntities adds timeslots, classes, cohorts, and assignments.
func applyEntities(s *store.Store, p Parsed, sum *Summary) error {
	for _, t := range p.Timeslots {
		if ok, err := tryApply(func() error { return s.AddTimeslot(t.Label, t.Day, t.Period) }, sum, "timeslot "+t.Label); err != nil {
			return err
		} else if ok {
			sum.Timeslots++
		}
	}
	for _, c := range p.Classes {
		if ok, err := tryApply(func() error { return s.AddClass(c) }, sum, "class "+c); err != nil {
			return err
		} else if ok {
			sum.Classes++
		}
	}
	for _, c := range p.Cohorts {
		if ok, err := tryApply(func() error { return s.AddCohort(c) }, sum, "cohort "+c); err != nil {
			return err
		} else if ok {
			sum.Cohorts++
		}
	}
	for _, a := range p.Assignments {
		label := fmt.Sprintf("assignment %s@%s→%s", a.Class, a.Timeslot, a.Cohort)
		if ok, err := tryApply(func() error { return s.Assign(a.Class, a.Timeslot, a.Cohort) }, sum, label); err != nil {
			return err
		} else if ok {
			sum.Assignments++
		}
	}
	return nil
}

// applyRules adds travel groups, blackouts, and the OneClassAtATime flag.
func applyRules(s *store.Store, p Parsed, sum *Summary) error {
	for class, buildings := range p.Travel {
		if ok, err := tryApply(func() error { return s.SetTravelGroups(class, buildings) }, sum, "travel for "+class); err != nil {
			return err
		} else if ok {
			sum.TravelRules++
		}
	}
	for _, b := range p.Blackouts {
		label := fmt.Sprintf("blackout %s@%s", b.Cohort, b.Timeslot)
		if ok, err := tryApply(func() error { return s.AddBlackout(b.Cohort, b.Timeslot) }, sum, label); err != nil {
			return err
		} else if ok {
			sum.Blackouts++
		}
	}
	if p.OneClassAtATime {
		if err := s.SetOneClassAtATime(true); err != nil {
			return fmt.Errorf("enabling OneClassAtATime: %w", err)
		}
		sum.OneClassAtATime = true
	}
	return nil
}

// tryApply calls fn and records a warning when the error is skippable.
// Returns (true, nil) on success, (false, nil) on skipped, (false, err) on failure.
func tryApply(fn func() error, sum *Summary, label string) (bool, error) {
	if err := fn(); err != nil {
		if isSkippable(err) {
			sum.Warnings = append(sum.Warnings, Warning{Text: label + ": " + err.Error()})
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// isSkippable returns true for UNIQUE constraint violations and "not found"
// errors — the two categories of referential mismatch that become Warnings.
func isSkippable(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "UNIQUE constraint") || strings.Contains(s, "not found")
}
