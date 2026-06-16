package ingest

import (
	"errors"
	"slices"
	"strings"
)

// Sheet is a 2D grid of trimmed cell strings, row-major.
type Sheet [][]string

// Timeslot is a parsed column header from the Schedule sheet.
type Timeslot struct {
	Label  string
	Day    string
	Period string
}

// Assignment is a filled cell in the Schedule grid.
type Assignment struct {
	Class    string
	Timeslot string
	Cohort   string
}

// Blackout is a cohort/timeslot restriction from the CohortBlacklist column.
type Blackout struct {
	Cohort   string
	Timeslot string
}

// Parsed holds everything extracted from the workbook, ready to apply to a store.
type Parsed struct {
	Timeslots       []Timeslot
	Classes         []string
	Cohorts         []string
	Assignments     []Assignment
	Travel          map[string][][]string // class → per-building cohort lists
	Blackouts       []Blackout
	OneClassAtATime bool
}

// Parse extracts schedule data from the two raw sheets.
// An empty or missing header structure is a hard error; referential mismatches
// (e.g. a CohortBlacklist timeslot that doesn't match a Schedule header) are
// preserved in Blackouts and become Warnings only when Apply is called.
func Parse(schedule, rules Sheet) (Parsed, error) {
	if len(schedule) < 2 {
		return Parsed{}, errors.New("schedule sheet must have at least one header row and one data row")
	}
	if len(schedule[0]) < 2 {
		return Parsed{}, errors.New("schedule sheet header row has no timeslot labels")
	}
	if len(rules) < 1 {
		return Parsed{}, errors.New("rules sheet is empty")
	}

	timeslots, classes, assignments := parseSchedule(schedule)

	cohorts := parseCohorts(column(rules, "AllCohorts"))
	travel := parseTravel(splitByBlanks(column(rules, "ClassRequiresTravel")))
	blackouts := parseBlacklist(splitByBlanks(column(rules, "CohortBlacklist")))
	oneClassAtATime := columnExists(rules[0], "OneClassAtATime")

	return Parsed{
		Timeslots:       timeslots,
		Classes:         classes,
		Cohorts:         cohorts,
		Assignments:     assignments,
		Travel:          travel,
		Blackouts:       blackouts,
		OneClassAtATime: oneClassAtATime,
	}, nil
}

// parseSchedule extracts timeslots, classes, and assignments from the Schedule sheet.
func parseSchedule(sheet Sheet) (timeslots []Timeslot, classes []string, assignments []Assignment) {
	headers := sheet[0]
	seenTS := map[string]bool{}
	for _, label := range headers[1:] {
		if label == "" || seenTS[label] {
			continue
		}
		seenTS[label] = true
		day, period := splitTimeslot(label)
		timeslots = append(timeslots, Timeslot{Label: label, Day: day, Period: period})
	}

	seenClass := map[string]bool{}
	for _, row := range sheet[1:] {
		if len(row) == 0 || row[0] == "" {
			continue
		}
		className := row[0]
		if !seenClass[className] {
			seenClass[className] = true
			classes = append(classes, className)
		}
		for j := 1; j < len(row); j++ {
			if row[j] == "" || j >= len(headers) || headers[j] == "" {
				continue
			}
			assignments = append(assignments, Assignment{
				Class:    className,
				Timeslot: headers[j],
				Cohort:   row[j],
			})
		}
	}
	return timeslots, classes, assignments
}

// parseCohorts returns deduplicated non-empty values from a column.
func parseCohorts(col []string) []string {
	seen := map[string]bool{}
	var cohorts []string
	for _, v := range col {
		if v != "" && !seen[v] {
			seen[v] = true
			cohorts = append(cohorts, v)
		}
	}
	return cohorts
}

// splitByBlanks splits a column by empty strings into sub-lists.
// A single empty string starts a new sub-list.
// Two consecutive empty strings produce a nil inner slice as a group separator.
func splitByBlanks(col []string) [][]string {
	var result [][]string
	var current []string
	for _, v := range col {
		if v == "" {
			result = append(result, current)
			current = nil
		} else {
			current = append(current, v)
		}
	}
	if len(current) > 0 {
		result = append(result, current)
	}
	return result
}

// column extracts the cells below the named header column.
func column(sheet Sheet, header string) []string {
	if len(sheet) == 0 {
		return nil
	}
	colIdx := -1
	for i, cell := range sheet[0] {
		if cell == header {
			colIdx = i
			break
		}
	}
	if colIdx < 0 {
		return nil
	}
	result := make([]string, 0, len(sheet)-1)
	for _, row := range sheet[1:] {
		if colIdx < len(row) {
			result = append(result, row[colIdx])
		} else {
			result = append(result, "")
		}
	}
	return result
}

// columnExists reports whether the header row contains the named column.
func columnExists(headerRow []string, name string) bool {
	return slices.Contains(headerRow, name)
}

// splitTimeslot splits "Monday, 8:40-9:20" → ("Monday", "8:40-9:20").
// If no ", " is present, returns (label, "").
func splitTimeslot(label string) (day, period string) {
	day, period, ok := strings.Cut(label, ", ")
	if !ok {
		return label, ""
	}
	return day, period
}

// parseTravel converts ClassRequiresTravel sub-lists into a class→buildings map.
// Each class group starts with [className, building1cohort, ...] followed by
// additional sub-lists for each building, terminated by a nil (double-blank) sub-list.
func parseTravel(subs [][]string) map[string][][]string {
	travel := make(map[string][][]string)
	var className string
	var buildings [][]string

	for _, sub := range subs {
		if len(sub) == 0 {
			if className != "" {
				travel[className] = buildings
				className = ""
				buildings = nil
			}
			continue
		}
		if className == "" {
			className = sub[0]
			if len(sub) > 1 {
				buildings = append(buildings, sub[1:])
			}
		} else {
			buildings = append(buildings, sub)
		}
	}
	if className != "" {
		travel[className] = buildings
	}
	return travel
}

// parseBlacklist converts CohortBlacklist sub-lists into Blackout records.
// The format mirrors ClassRequiresTravel: the cohort name is a single-element
// sub-list (blank-separated from its timeslots), followed by one or more
// timeslot sub-lists, terminated by a nil (double-blank) group separator.
func parseBlacklist(subs [][]string) []Blackout {
	var blackouts []Blackout
	var cohort string
	for _, sub := range subs {
		if len(sub) == 0 {
			cohort = ""
			continue
		}
		if cohort == "" {
			cohort = sub[0] // first sub-list after reset is the cohort header
			continue
		}
		for _, ts := range sub {
			if ts != "" {
				blackouts = append(blackouts, Blackout{Cohort: cohort, Timeslot: ts})
			}
		}
	}
	return blackouts
}
