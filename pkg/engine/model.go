// Package engine implements the schedule validation rules engine.
//
// A schedule is modeled as a Grid — a 2D matrix of cohort assignments where
// classes are rows and timeslots are columns. Validate walks the grid and
// applies a set of Rules, returning a Violation for every cell that breaks a
// rule. The engine only computes violations; it does not mutate the grid or
// render output (highlighting cells, attaching comments, etc.), which is left
// to callers.
//
// The domain types here are the stable contract that later phases (CLI, MCP
// tools, generated views) feed into and render from, so the engine is the
// single source of truth for whether a schedule is valid.
package engine

// Grid is the schedule under validation: classes are rows and timeslots are
// columns, both in a fixed order. Each cell holds the name of the cohort
// assigned to that (class, timeslot) pair, or "" when the slot is unassigned.
//
// Cells is row-major and must be rectangular: len(Cells) == len(Classes) and
// every len(Cells[r]) == len(Timeslots). Cells[r][c] is the cohort scheduled
// for Classes[r] during Timeslots[c].
type Grid struct {
	Classes   []string   // row labels, in display order
	Timeslots []string   // column labels, in chronological order — adjacency drives ClassRequiresTravel
	Cells     [][]string // Cells[r][c] = cohort at (Classes[r], Timeslots[c]); "" == empty
}

// Cell identifies a single position in the grid by its labels rather than its
// indices, so a Violation stays meaningful without a reference to the Grid.
type Cell struct {
	Class    string // the class (row) label
	Timeslot string // the timeslot (column) label
}

// Rules is the full set of constraints applied to a Grid. ClassRequiresTravel
// and CohortBlacklist are off when their slices are empty (no entries to
// match), and OneClassAtATime is off unless explicitly set true. AllCohorts is
// the exception: an empty list flags every assigned cohort rather than
// disabling the check, so it must be populated whenever the grid has
// assignments.
type Rules struct {
	AllCohorts          []string        // master list of valid cohort names; any cohort not in it is flagged
	ClassRequiresTravel []TravelRule    // per-class building groupings restricting back-to-back assignments
	CohortBlacklist     []BlacklistRule // per-cohort timeslots in which that cohort may not be scheduled
	OneClassAtATime     bool            // when true, flag a cohort that appears in two classes in the same timeslot
}

// TravelRule constrains a mobile Class (e.g. "Latin Cart") that must travel
// between buildings. Buildings[i] is the set of cohorts whose homerooms are
// colocated in building i. The class cannot move between two different
// buildings in a single timeslot transition, so an assignment is flagged when
// the cohort in the prior column lives in a different building than the
// cohort in the current column. A rule with fewer than two buildings imposes
// no constraint.
type TravelRule struct {
	Class     string     // the class (row) this rule applies to
	Buildings [][]string // Buildings[i] = cohorts colocated in building i
}

// BlacklistRule forbids a Cohort from being scheduled for any class during the
// listed Timeslots (e.g. lunch, recess, assemblies).
type BlacklistRule struct {
	Cohort    string   // the cohort this rule applies to
	Timeslots []string // timeslot labels in which Cohort may not be scheduled
}

// RuleType names the rule that produced a Violation, matching the rule names
// used in the documentation and the add-in.
type RuleType string

const (
	RuleAllCohorts          RuleType = "AllCohorts"
	RuleClassRequiresTravel RuleType = "ClassRequiresTravel"
	RuleCohortBlacklist     RuleType = "CohortBlacklist"
	RuleOneClassAtATime     RuleType = "OneClassAtATime"
)

// Violation reports a single broken rule at a single cell. A cell may produce
// more than one Violation when it breaks multiple rules.
type Violation struct {
	Cell    Cell     // where the violation occurred
	Cohort  string   // the offending cohort assigned at Cell
	Rule    RuleType // which rule was broken
	Message string   // human-readable explanation suitable for display to the scheduler
}
