package engine

import "fmt"

// Validate applies every configured rule in r to each assigned cell of g and
// returns all violations found. Empty cells and rule-exempt cells (see
// isExempt) are skipped entirely. A single cell may yield multiple violations.
//
// The grid is walked row-major (classes outer, timeslots inner), so the
// returned slice has a deterministic order, which callers and tests can rely
// on. The returned slice is nil when there are no violations.
func Validate(g Grid, r Rules) []Violation {
	var violations []Violation

	for row, class := range g.Classes {
		for col, timeslot := range g.Timeslots {
			cohort := g.Cells[row][col]
			if cohort == "" || isExempt(cohort) {
				continue
			}

			cell := Cell{Class: class, Timeslot: timeslot}

			violations = append(violations, checkAllCohorts(cell, cohort, r)...)
			violations = append(violations, checkTravel(g, row, col, cell, cohort, r)...)
			violations = append(violations, checkBlacklist(cell, cohort, timeslot, r)...)
			if r.OneClassAtATime {
				violations = append(violations, checkOneAtATime(g, row, col, class, timeslot, cohort)...)
			}
		}
	}

	return violations
}

// checkAllCohorts flags cohort when it is not present in the master list
// r.AllCohorts, catching typos and cohorts that do not belong in the schedule.
// Note that an empty AllCohorts flags every assigned cohort rather than
// disabling the check, so callers must supply the master list whenever any
// cells are assigned.
func checkAllCohorts(cell Cell, cohort string, r Rules) []Violation {
	if contains(r.AllCohorts, cohort) {
		return nil
	}
	return []Violation{{
		Cell:    cell,
		Cohort:  cohort,
		Rule:    RuleAllCohorts,
		Message: fmt.Sprintf("The cohort '%s' isn't in the total list of classes, check the AllCohorts rule.", cohort),
	}}
}

// checkTravel enforces the ClassRequiresTravel rule for the cell at (row, col).
// A mobile class cannot move between two different buildings in a single
// timeslot transition, so the cell is flagged when the prior column's cohort
// (same class row) lives in a different building than the current cohort.
//
// The first column (col == 0) has no prior cell and is never flagged. The rule
// only applies when a TravelRule exists for this class and defines at least two
// buildings. If either the current or the prior cohort is not placed in any
// building, no judgment can be made and the cell is not flagged.
func checkTravel(g Grid, row, col int, cell Cell, cohort string, r Rules) []Violation {
	if col == 0 {
		return nil
	}

	rule := findTravelRule(r.ClassRequiresTravel, g.Classes[row])
	if rule == nil || len(rule.Buildings) < 2 {
		return nil
	}

	currentBuilding := buildingOf(rule.Buildings, cohort)
	if currentBuilding == -1 {
		return nil
	}

	prior := g.Cells[row][col-1]
	priorBuilding := buildingOf(rule.Buildings, prior)
	if priorBuilding == -1 || priorBuilding == currentBuilding {
		return nil
	}

	return []Violation{{
		Cell:    cell,
		Cohort:  cohort,
		Rule:    RuleClassRequiresTravel,
		Message: fmt.Sprintf("The class '%s' can't go to one cohort '%s' if the previous one was '%s', it is too far away (or requires setup) — see the ClassRequiresTravel rule.", g.Classes[row], cohort, prior),
	}}
}

// checkBlacklist enforces the CohortBlacklist rule: it flags the cell when
// cohort has a BlacklistRule listing the current timeslot among the slots in
// which that cohort may not be scheduled. At most one violation is returned
// even if multiple matching rules exist.
func checkBlacklist(cell Cell, cohort, timeslot string, r Rules) []Violation {
	for _, bl := range r.CohortBlacklist {
		if bl.Cohort != cohort {
			continue
		}
		if contains(bl.Timeslots, timeslot) {
			return []Violation{{
				Cell:    cell,
				Cohort:  cohort,
				Rule:    RuleCohortBlacklist,
				Message: fmt.Sprintf("The cohort '%s' is not allowed to have any class during '%s' as defined in the CohortBlacklist rule.", cohort, timeslot),
			}}
		}
	}
	return nil
}

// checkOneAtATime enforces the OneClassAtATime rule for the cell at (row, col):
// it flags the cell when the same cohort is also assigned to a different class
// in the same timeslot (column), since a cohort cannot attend two classes at
// once. Each side of a conflict is flagged independently — Validate calls this
// for both cells — so a pair of conflicting assignments produces two
// violations. Scanning stops at the first conflicting class found, as one is
// enough to flag this cell.
func checkOneAtATime(g Grid, row, col int, class, timeslot, cohort string) []Violation {
	for otherRow, otherClass := range g.Classes {
		if otherRow == row {
			continue
		}
		if g.Cells[otherRow][col] == cohort {
			return []Violation{{
				Cell:    Cell{Class: class, Timeslot: timeslot},
				Cohort:  cohort,
				Rule:    RuleOneClassAtATime,
				Message: fmt.Sprintf("The cohort '%s' is scheduled for both '%s' and '%s' during '%s'. Cohorts can only attend one class at a time according to the OneClassAtATime rule.", cohort, class, otherClass, timeslot),
			}}
		}
	}
	return nil
}

// findTravelRule returns the TravelRule for the given class, or nil if none of
// the rules apply to it.
func findTravelRule(rules []TravelRule, class string) *TravelRule {
	for i := range rules {
		if rules[i].Class == class {
			return &rules[i]
		}
	}
	return nil
}

// buildingOf returns the index of the first building containing cohort, or -1
// if cohort is not placed in any building.
func buildingOf(buildings [][]string, cohort string) int {
	for i, b := range buildings {
		if contains(b, cohort) {
			return i
		}
	}
	return -1
}

// contains reports whether s appears in list.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
