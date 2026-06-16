package ingest

import (
	"reflect"
	"testing"
)

func TestSplitByBlanks(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  [][]string
	}{
		{"empty", []string{}, nil},
		{"no blanks", []string{"a", "b"}, [][]string{{"a", "b"}}},
		{"single blank", []string{"a", "", "b"}, [][]string{{"a"}, {"b"}}},
		{"double blank produces nil inner", []string{"a", "", "", "b"}, [][]string{{"a"}, nil, {"b"}}},
		{"travel shape two buildings", []string{"Latin A", "3M", "3B", "", "4B", "4M", "", ""},
			[][]string{{"Latin A", "3M", "3B"}, {"4B", "4M"}, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitByBlanks(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitByBlanks(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitTimeslot(t *testing.T) {
	tests := []struct {
		label, wantDay, wantPeriod string
	}{
		{"Monday, 8:40-9:20", "Monday", "8:40-9:20"},
		{"Friday, 2:40-3:20", "Friday", "2:40-3:20"},
		{"Friday 2:40-3:20", "Friday 2:40-3:20", ""},
		{"Wednesday, 10:00-10:40", "Wednesday", "10:00-10:40"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			gotDay, gotPeriod := splitTimeslot(tt.label)
			if gotDay != tt.wantDay || gotPeriod != tt.wantPeriod {
				t.Errorf("splitTimeslot(%q) = (%q, %q), want (%q, %q)",
					tt.label, gotDay, gotPeriod, tt.wantDay, tt.wantPeriod)
			}
		})
	}
}

// TestParse exercises the full Parse path against an inline fixture that covers:
// the two-building ClassRequiresTravel case and the comma-mismatch blackout
// ("Friday 2:40-3:20" vs schedule header "Friday, 2:40-3:20").
//
// Column structure (matches the real workbook convention):
//   - ClassRequiresTravel: class name on own sub-list, then building sub-lists, double-blank between classes
//   - CohortBlacklist: cohort name on own sub-list, then timeslot sub-list, double-blank between cohorts
func TestParse(t *testing.T) {
	schedule := Sheet{
		{"", "Monday, 8:40-9:20", "Friday, 2:40-3:20"},
		{"Latin A", "3M", ""},
		{"Art", "", "PKTB"},
	}
	// Rules columns (A=AllCohorts, B=ClassRequiresTravel, C=CohortBlacklist, D=OneClassAtATime)
	// Each column is independent; rows are aligned only for sheet representation.
	rules := Sheet{
		{"AllCohorts", "ClassRequiresTravel", "CohortBlacklist", "OneClassAtATime"},
		{"3M", "Latin A", "3M", ""},            // A: cohort; B: class header; C: cohort header
		{"3B", "", "", ""},                     // A: cohort; B: blank (sep after class name); C: blank (sep after cohort name)
		{"PKTB", "3M", "Friday 2:40-3:20", ""}, // A: cohort; B: bldg1 cohort; C: timeslot (no comma = mismatch)
		{"", "3B", "", ""},                     // B: bldg1 cohort; C: blank (end timeslots)
		{"", "", "", ""},                       // B: blank (end bldg1); C: blank (double-blank = group sep)
		{"", "4B", "", ""},                     // B: bldg2 cohort
		{"", "4M", "", ""},                     // B: bldg2 cohort
		{"", "", "", ""},                       // B: blank (end bldg2)
		{"", "", "", ""},                       // B: blank (double-blank = class sep)
	}

	p, err := Parse(schedule, rules)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Timeslots) != 2 {
		t.Errorf("timeslots: got %d, want 2", len(p.Timeslots))
	}
	if p.Timeslots[0].Label != "Monday, 8:40-9:20" || p.Timeslots[0].Day != "Monday" || p.Timeslots[0].Period != "8:40-9:20" {
		t.Errorf("timeslot[0]: got %+v", p.Timeslots[0])
	}

	if len(p.Classes) != 2 {
		t.Errorf("classes: got %d, want 2", len(p.Classes))
	}

	if len(p.Cohorts) != 3 {
		t.Errorf("cohorts: got %d, want 3", len(p.Cohorts))
	}

	if len(p.Assignments) != 2 {
		t.Errorf("assignments: got %d, want 2", len(p.Assignments))
	}

	bldgs, ok := p.Travel["Latin A"]
	if !ok {
		t.Fatal("expected travel entry for Latin A")
	}
	wantBldgs := [][]string{{"3M", "3B"}, {"4B", "4M"}}
	if !reflect.DeepEqual(bldgs, wantBldgs) {
		t.Errorf("travel[Latin A]: got %v, want %v", bldgs, wantBldgs)
	}

	if len(p.Blackouts) != 1 {
		t.Fatalf("blackouts: got %d, want 1", len(p.Blackouts))
	}
	if p.Blackouts[0].Cohort != "3M" || p.Blackouts[0].Timeslot != "Friday 2:40-3:20" {
		t.Errorf("blackout[0]: got %+v", p.Blackouts[0])
	}

	if !p.OneClassAtATime {
		t.Error("expected OneClassAtATime = true")
	}
}
