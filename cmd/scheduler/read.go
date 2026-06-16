package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/bmayfi3ld/excel-scheduler/pkg/store"
)

type gridCmd struct {
	withDB
	jsonFlag
}

func (c *gridCmd) Run(s *store.Store) error {
	g, err := s.Grid()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(g)
	}
	if len(g.Classes) == 0 || len(g.Timeslots) == 0 {
		fmt.Println("(empty grid — add classes and timeslots first)")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprint(tw, "class")
	for _, ts := range g.Timeslots {
		fmt.Fprintf(tw, "\t%s", ts)
	}
	fmt.Fprintln(tw)
	for row, class := range g.Classes {
		fmt.Fprint(tw, class)
		for col := range g.Timeslots {
			cell := g.Cells[row][col]
			if cell == "" {
				cell = "-"
			}
			fmt.Fprintf(tw, "\t%s", cell)
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}

type listClassesCmd struct {
	withDB
	jsonFlag
}

func (c *listClassesCmd) Run(s *store.Store) error {
	classes, err := s.Classes()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(classes)
	}
	return printList(classes, "(no classes)")
}

type listTimeslotsCmd struct {
	withDB
	jsonFlag
}

func (c *listTimeslotsCmd) Run(s *store.Store) error {
	timeslots, err := s.Timeslots()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(timeslots)
	}
	if len(timeslots) == 0 {
		fmt.Println("(no timeslots)")
		return nil
	}
	for _, t := range timeslots {
		line := t.Label
		if t.Day != "" || t.Period != "" {
			line += fmt.Sprintf("  (day=%s period=%s)", t.Day, t.Period)
		}
		fmt.Println(line)
	}
	return nil
}

type listCohortsCmd struct {
	withDB
	jsonFlag
}

func (c *listCohortsCmd) Run(s *store.Store) error {
	cohorts, err := s.Cohorts()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(cohorts)
	}
	return printList(cohorts, "(no cohorts)")
}

type showRulesCmd struct {
	withDB
	jsonFlag
}

func (c *showRulesCmd) Run(s *store.Store) error {
	rules, err := s.RulesConfig()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(rules)
	}
	fmt.Printf("OneClassAtATime: %v\n", rules.OneClassAtATime)
	fmt.Println("\nClassRequiresTravel:")
	if len(rules.ClassRequiresTravel) == 0 {
		fmt.Println("  (none)")
	}
	for _, tr := range rules.ClassRequiresTravel {
		fmt.Printf("  %s:\n", tr.Class)
		for i, building := range tr.Buildings {
			fmt.Printf("    building %d: %v\n", i+1, building)
		}
	}
	fmt.Println("\nCohortBlacklist:")
	if len(rules.CohortBlacklist) == 0 {
		fmt.Println("  (none)")
	}
	for _, bl := range rules.CohortBlacklist {
		fmt.Printf("  %s: %v\n", bl.Cohort, bl.Timeslots)
	}
	return nil
}

// validateCmd exits non-zero when violations exist, making it scriptable.
type validateCmd struct {
	withDB
	jsonFlag
}

func (c *validateCmd) Run(s *store.Store) error {
	violations, err := s.Validate()
	if err != nil {
		return err
	}
	if c.JSON {
		if err := printJSON(violations); err != nil {
			return err
		}
	} else if len(violations) == 0 {
		fmt.Println("No violations.")
	} else {
		fmt.Printf("%d violation(s):\n", len(violations))
		for _, v := range violations {
			fmt.Printf("  [%s] %s @ %s: %s\n", v.Rule, v.Cell.Class, v.Cell.Timeslot, v.Message)
		}
	}
	if len(violations) > 0 {
		return errSilent
	}
	return nil
}

type listUnassignedCmd struct {
	withDB
	jsonFlag
}

func (c *listUnassignedCmd) Run(s *store.Store) error {
	cells, err := s.ListUnassigned()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(cells)
	}
	if len(cells) == 0 {
		fmt.Println("(every cell is assigned)")
		return nil
	}
	for _, cell := range cells {
		fmt.Printf("%s @ %s\n", cell.Class, cell.Timeslot)
	}
	return nil
}

type reportCmd struct {
	withDB
	Cohort string `help:"limit to one cohort (default: all)"`
	jsonFlag
}

func (c *reportCmd) Run(s *store.Store) error {
	entries, err := s.Report(c.Cohort)
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(entries)
	}
	if len(entries) == 0 {
		fmt.Println("(no assignments)")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "cohort\ttimeslot\tclass")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Cohort, e.Timeslot, e.Class)
	}
	return tw.Flush()
}

func printList(items []string, empty string) error {
	if len(items) == 0 {
		fmt.Println(empty)
		return nil
	}
	for _, item := range items {
		fmt.Println(item)
	}
	return nil
}
