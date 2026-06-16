package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmayfi3ld/excel-scheduler/pkg/ingest"
	"github.com/bmayfi3ld/excel-scheduler/pkg/store"
)

type importCmd struct {
	DB            string `required:"" help:"schedule database file to create"`
	From          string `required:"" help:"source .xlsx file"`
	Name          string `help:"schedule name (defaults to <file.db> basename without extension)"`
	ScheduleSheet string `default:"Schedule" help:"name of the Schedule sheet"`
	RulesSheet    string `default:"Rules" help:"name of the Rules sheet"`
}

func (c *importCmd) Run() error {
	if _, statErr := os.Stat(c.DB); statErr == nil { //nolint:gosec // path is a CLI argument
		return fmt.Errorf("file already exists: %s (remove it or choose another path)", c.DB)
	}

	displayName := c.Name
	if displayName == "" {
		base := filepath.Base(c.DB)
		displayName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	schedule, rules, err := ingest.ReadWorkbook(c.From, c.ScheduleSheet, c.RulesSheet)
	if err != nil {
		return fmt.Errorf("reading workbook: %w", err)
	}

	parsed, err := ingest.Parse(schedule, rules)
	if err != nil {
		return fmt.Errorf("parsing workbook: %w", err)
	}

	s, err := store.Create(c.DB, displayName)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	sum, err := ingest.Apply(s, parsed)
	if err != nil {
		_ = os.Remove(c.DB) //nolint:gosec // path is a CLI argument
		return fmt.Errorf("importing: %w", err)
	}

	fmt.Printf("Imported %q → %s\n\n", displayName, c.DB)
	fmt.Printf("  Classes:         %d\n", sum.Classes)
	fmt.Printf("  Timeslots:       %d\n", sum.Timeslots)
	fmt.Printf("  Cohorts:         %d\n", sum.Cohorts)
	fmt.Printf("  Assignments:     %d\n", sum.Assignments)
	fmt.Printf("  Travel rules:    %d\n", sum.TravelRules)
	fmt.Printf("  Blackouts:       %d\n", sum.Blackouts)
	fmt.Printf("  OneClassAtATime: %v\n", sum.OneClassAtATime)

	if len(sum.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(sum.Warnings))
		for _, w := range sum.Warnings {
			fmt.Printf("  ! %s\n", w.Text)
		}
	}

	violations, err := s.Validate()
	if err != nil {
		return fmt.Errorf("validating: %w", err)
	}
	fmt.Printf("\nValidation: %d violation(s)\n", len(violations))
	return nil
}
