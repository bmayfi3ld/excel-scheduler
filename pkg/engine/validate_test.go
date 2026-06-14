package engine

import (
	"testing"
)

// travelRules mirrors the rules.md ClassRequiresTravel example:
// Latin Cart and Lunch Cart, three buildings each.
var travelRules = []TravelRule{
	{
		Class: "Latin Cart",
		Buildings: [][]string{
			{"1st", "2nd", "3rd"},
			{"4th", "5th", "6th"},
		},
	},
	{
		Class: "Lunch Cart",
		Buildings: [][]string{
			{"1st", "2nd", "3rd"},
			{"4th", "5th", "6th"},
			{"7th", "8th", "9th"},
		},
	},
}

var allCohorts = []string{"1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th"}

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		grid           Grid
		rules          Rules
		wantCount      int
		wantViolations []struct {
			rule RuleType
			cell Cell
		}
	}{
		{
			// rules.md:122-134 — Latin Cart 1st→4th crosses buildings: 1 violation.
			// Lunch Cart 1st→2nd same building: 0 violations.
			name: "Travel/cross-building gives violation",
			grid: Grid{
				Classes:   []string{"Latin Cart", "Lunch Cart"},
				Timeslots: []string{"Monday, 9am", "Monday, 10am", "Monday, 11am", "Monday, 12pm"},
				Cells: [][]string{
					{"1st", "4th", "", ""},
					{"", "", "1st", "2nd"},
				},
			},
			rules: Rules{
				AllCohorts:          allCohorts,
				ClassRequiresTravel: travelRules,
			},
			wantCount: 1,
			wantViolations: []struct {
				rule RuleType
				cell Cell
			}{
				{RuleClassRequiresTravel, Cell{Class: "Latin Cart", Timeslot: "Monday, 10am"}},
			},
		},
		{
			// Same building transition: no travel violation.
			name: "Travel/same-building no violation",
			grid: Grid{
				Classes:   []string{"Lunch Cart"},
				Timeslots: []string{"Monday, 11am", "Monday, 12pm"},
				Cells:     [][]string{{"1st", "2nd"}},
			},
			rules: Rules{
				AllCohorts:          allCohorts,
				ClassRequiresTravel: travelRules,
			},
			wantCount: 0,
		},
		{
			// rules.md:155-181 — 1st blacklisted at Monday, 11am → 1 violation.
			name: "Blacklist/blocked timeslot gives violation",
			grid: Grid{
				Classes:   []string{"Math"},
				Timeslots: []string{"Monday, 11am"},
				Cells:     [][]string{{"1st"}},
			},
			rules: Rules{
				AllCohorts: allCohorts,
				CohortBlacklist: []BlacklistRule{
					{Cohort: "1st", Timeslots: []string{"Monday, 11am", "Monday, 12pm"}},
					{Cohort: "2nd", Timeslots: []string{"Monday, 12pm", "Monday, 1pm"}},
				},
			},
			wantCount: 1,
			wantViolations: []struct {
				rule RuleType
				cell Cell
			}{
				{RuleCohortBlacklist, Cell{Class: "Math", Timeslot: "Monday, 11am"}},
			},
		},
		{
			// Cohort at a non-blacklisted slot: no violation.
			name: "Blacklist/allowed timeslot no violation",
			grid: Grid{
				Classes:   []string{"Math"},
				Timeslots: []string{"Monday, 9am"},
				Cells:     [][]string{{"1st"}},
			},
			rules: Rules{
				AllCohorts: allCohorts,
				CohortBlacklist: []BlacklistRule{
					{Cohort: "1st", Timeslots: []string{"Monday, 11am", "Monday, 12pm"}},
				},
			},
			wantCount: 0,
		},
		{
			// rules.md:199-206 — Math=1st and Art=1st at Monday, 9am → 2 violations.
			name: "OneClassAtATime/conflict gives two violations",
			grid: Grid{
				Classes:   []string{"Math", "Art"},
				Timeslots: []string{"Monday, 9am"},
				Cells:     [][]string{{"1st"}, {"1st"}},
			},
			rules: Rules{
				AllCohorts:      allCohorts,
				OneClassAtATime: true,
			},
			wantCount: 2,
			wantViolations: []struct {
				rule RuleType
				cell Cell
			}{
				{RuleOneClassAtATime, Cell{Class: "Math", Timeslot: "Monday, 9am"}},
				{RuleOneClassAtATime, Cell{Class: "Art", Timeslot: "Monday, 9am"}},
			},
		},
		{
			// Same grid with rule disabled → 0 violations.
			name: "OneClassAtATime/disabled no violation",
			grid: Grid{
				Classes:   []string{"Math", "Art"},
				Timeslots: []string{"Monday, 9am"},
				Cells:     [][]string{{"1st"}, {"1st"}},
			},
			rules: Rules{
				AllCohorts:      allCohorts,
				OneClassAtATime: false,
			},
			wantCount: 0,
		},
		{
			// rules.md:42-44 — cohort "53rd" not in list → 1 violation.
			name: "AllCohorts/unknown cohort gives violation",
			grid: Grid{
				Classes:   []string{"Math"},
				Timeslots: []string{"Monday, 9am"},
				Cells:     [][]string{{"53rd"}},
			},
			rules: Rules{
				AllCohorts: allCohorts,
			},
			wantCount: 1,
			wantViolations: []struct {
				rule RuleType
				cell Cell
			}{
				{RuleAllCohorts, Cell{Class: "Math", Timeslot: "Monday, 9am"}},
			},
		},
		{
			// Exempt cell ("#### blocked") must never be flagged even if it would break AllCohorts.
			name: "Exemption/repeating-char cell skipped",
			grid: Grid{
				Classes:   []string{"Math"},
				Timeslots: []string{"Monday, 9am"},
				Cells:     [][]string{{"#### blocked"}},
			},
			rules: Rules{
				AllCohorts:      allCohorts,
				OneClassAtATime: true,
			},
			wantCount: 0,
		},
		{
			// A fully valid grid produces zero violations.
			name: "Clean/valid grid no violations",
			grid: Grid{
				Classes:   []string{"Math", "Art"},
				Timeslots: []string{"Monday, 9am", "Monday, 10am"},
				Cells: [][]string{
					{"1st", "2nd"},
					{"3rd", "4th"},
				},
			},
			rules: Rules{
				AllCohorts:      allCohorts,
				OneClassAtATime: true,
				CohortBlacklist: []BlacklistRule{
					{Cohort: "1st", Timeslots: []string{"Monday, 11am"}},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.grid, tt.rules)
			if len(got) != tt.wantCount {
				t.Errorf("violation count = %d, want %d; got %+v", len(got), tt.wantCount, got)
				return
			}
			for i, want := range tt.wantViolations {
				if got[i].Rule != want.rule {
					t.Errorf("[%d] Rule = %q, want %q", i, got[i].Rule, want.rule)
				}
				if got[i].Cell != want.cell {
					t.Errorf("[%d] Cell = %+v, want %+v", i, got[i].Cell, want.cell)
				}
			}
		})
	}
}
