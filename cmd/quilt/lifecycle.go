package main

import (
	"fmt"

	"github.com/bmayfi3ld/quilt/pkg/store"
)

type initCmd struct {
	DB   string `required:"" help:"schedule database file to create"`
	Name string `help:"schedule name (defaults to the file path)"`
}

func (c *initCmd) Run() error {
	displayName := c.Name
	if displayName == "" {
		displayName = c.DB
	}
	s, err := store.Create(c.DB, displayName)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	fmt.Printf("Created schedule %q at %s\n", displayName, c.DB)
	return nil
}

type infoCmd struct {
	withDB
	jsonFlag
}

func (c *infoCmd) Run(s *store.Store) error {
	info, err := s.Info()
	if err != nil {
		return err
	}
	if c.JSON {
		return printJSON(info)
	}
	fmt.Printf("Name:            %s\n", info.Name)
	fmt.Printf("Schema version:  %d\n", info.SchemaVersion)
	fmt.Printf("Created:         %s\n", info.CreatedAt)
	fmt.Printf("Updated:         %s\n", info.UpdatedAt)
	fmt.Printf("Classes:         %d\n", info.Classes)
	fmt.Printf("Timeslots:       %d\n", info.Timeslots)
	fmt.Printf("Cohorts:         %d\n", info.Cohorts)
	fmt.Printf("Assignments:     %d\n", info.Assignments)
	fmt.Printf("OneClassAtATime: %v\n", info.OneClassAtATime)
	fmt.Printf("Travel rules:    %d\n", info.TravelRules)
	fmt.Printf("Blackouts:       %d\n", info.Blackouts)
	return nil
}

type copyCmd struct {
	DB   string `required:"" help:"source schedule database file"`
	Out  string `required:"" help:"destination database file to create"`
	Name string `help:"name for the new schedule (keeps source name if omitted)"`
}

func (c *copyCmd) Run() error {
	if err := store.Copy(c.DB, c.Out, c.Name); err != nil {
		return err
	}
	fmt.Printf("Copied %s -> %s\n", c.DB, c.Out)
	return nil
}
