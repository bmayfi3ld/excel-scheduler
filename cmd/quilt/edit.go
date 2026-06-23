package main

import (
	"fmt"
	"strings"

	"github.com/bmayfi3ld/quilt/pkg/store"
)

// --- Structure ---

type addClassCmd struct {
	withDB
	Name string `required:"" help:"class name"`
}

func (c *addClassCmd) Run(s *store.Store) error {
	if err := s.AddClass(c.Name); err != nil {
		return err
	}
	fmt.Printf("Added class %q\n", c.Name)
	return nil
}

type addTimeslotCmd struct {
	withDB
	Label  string `required:"" help:"timeslot label"`
	Day    string `help:"day (optional)"`
	Period string `help:"period (optional)"`
}

func (c *addTimeslotCmd) Run(s *store.Store) error {
	if err := s.AddTimeslot(c.Label, c.Day, c.Period); err != nil {
		return err
	}
	fmt.Printf("Added timeslot %q\n", c.Label)
	return nil
}

type addCohortCmd struct {
	withDB
	Name string `required:"" help:"cohort name"`
}

func (c *addCohortCmd) Run(s *store.Store) error {
	if err := s.AddCohort(c.Name); err != nil {
		return err
	}
	fmt.Printf("Added cohort %q\n", c.Name)
	return nil
}

type removeClassCmd struct {
	withDB
	Name string `required:"" help:"class name"`
}

func (c *removeClassCmd) Run(s *store.Store) error {
	if err := s.RemoveClass(c.Name); err != nil {
		return err
	}
	fmt.Printf("Removed class %q\n", c.Name)
	return nil
}

type removeTimeslotCmd struct {
	withDB
	Label string `required:"" help:"timeslot label"`
}

func (c *removeTimeslotCmd) Run(s *store.Store) error {
	if err := s.RemoveTimeslot(c.Label); err != nil {
		return err
	}
	fmt.Printf("Removed timeslot %q\n", c.Label)
	return nil
}

type removeCohortCmd struct {
	withDB
	Name string `required:"" help:"cohort name"`
}

func (c *removeCohortCmd) Run(s *store.Store) error {
	if err := s.RemoveCohort(c.Name); err != nil {
		return err
	}
	fmt.Printf("Removed cohort %q\n", c.Name)
	return nil
}

// --- Rules ---

type enableRuleCmd struct {
	withDB
	Rule string `arg:"" enum:"one-class-at-a-time" help:"rule name"`
	Off  bool   `help:"disable the rule instead of enabling it"`
}

func (c *enableRuleCmd) Run(s *store.Store) error {
	switch c.Rule {
	case "one-class-at-a-time":
		if err := s.SetOneClassAtATime(!c.Off); err != nil {
			return err
		}
	}
	state := "enabled"
	if c.Off {
		state = "disabled"
	}
	fmt.Printf("Rule %q %s\n", c.Rule, state)
	return nil
}

type setTravelCmd struct {
	withDB
	Class    string   `required:"" help:"class the travel rule applies to"`
	Building []string `required:"" help:"comma-separated cohorts colocated in one building (repeatable)"`
}

func (c *setTravelCmd) Run(s *store.Store) error {
	groups := make([][]string, 0, len(c.Building))
	for _, b := range c.Building {
		var cohorts []string
		for _, part := range strings.Split(b, ",") {
			if part = strings.TrimSpace(part); part != "" {
				cohorts = append(cohorts, part)
			}
		}
		groups = append(groups, cohorts)
	}
	if err := s.SetTravelGroups(c.Class, groups); err != nil {
		return err
	}
	fmt.Printf("Set travel rule for %q across %d buildings\n", c.Class, len(groups))
	return nil
}

type addBlackoutCmd struct {
	withDB
	Cohort   string `required:"" help:"cohort name"`
	Timeslot string `required:"" help:"timeslot label"`
}

func (c *addBlackoutCmd) Run(s *store.Store) error {
	if err := s.AddBlackout(c.Cohort, c.Timeslot); err != nil {
		return err
	}
	fmt.Printf("Blacked out %q during %q\n", c.Cohort, c.Timeslot)
	return nil
}

// --- Grid edits ---

type assignCmd struct {
	withDB
	Class    string `required:"" help:"class name"`
	Timeslot string `required:"" help:"timeslot label"`
	Cohort   string `required:"" help:"cohort value (any text, incl. exempt placeholders)"`
}

func (c *assignCmd) Run(s *store.Store) error {
	if err := s.Assign(c.Class, c.Timeslot, c.Cohort); err != nil {
		return err
	}
	fmt.Printf("Assigned %q to %q during %q\n", c.Cohort, c.Class, c.Timeslot)
	return nil
}

type unassignCmd struct {
	withDB
	Class    string `required:"" help:"class name"`
	Timeslot string `required:"" help:"timeslot label"`
}

func (c *unassignCmd) Run(s *store.Store) error {
	if err := s.Unassign(c.Class, c.Timeslot); err != nil {
		return err
	}
	fmt.Printf("Cleared %q during %q\n", c.Class, c.Timeslot)
	return nil
}
