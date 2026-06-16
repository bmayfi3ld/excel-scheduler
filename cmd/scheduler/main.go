// Command scheduler is a thin CLI over pkg/store: it parses flags, calls one
// store method, and renders the result. All logic lives in the store so the
// Phase 2 MCP server can wrap the same methods (CLI command ⇄ MCP tool).
//
// Form: scheduler <command> --db <file.db> [flags].
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/bmayfi3ld/excel-scheduler/pkg/store"
)

// errSilent signals "exit non-zero without printing an error line" — used by
// validate when violations exist (it has already printed them).
var errSilent = errors.New("silent failure")

type CLI struct {
	Init           initCmd           `cmd:"" help:"create an empty schedule"`
	Info           infoCmd           `cmd:"" help:"name, timestamps, counts, rule status"`
	Copy           copyCmd           `cmd:"" help:"branch a schedule into a new file"`
	Import         importCmd         `cmd:"" help:"import an .xlsx workbook into a new schedule"`
	AddClass       addClassCmd       `cmd:"" help:"add a class"`
	AddTimeslot    addTimeslotCmd    `cmd:"" help:"add a timeslot"`
	AddCohort      addCohortCmd      `cmd:"" help:"add a cohort"`
	RemoveClass    removeClassCmd    `cmd:"" help:"remove a class"`
	RemoveTimeslot removeTimeslotCmd `cmd:"" help:"remove a timeslot"`
	RemoveCohort   removeCohortCmd   `cmd:"" help:"remove a cohort"`
	EnableRule     enableRuleCmd     `cmd:"" help:"enable or disable a rule"`
	SetTravel      setTravelCmd      `cmd:"" help:"set travel groups for a class"`
	AddBlackout    addBlackoutCmd    `cmd:"" help:"add a cohort blackout"`
	Assign         assignCmd         `cmd:"" help:"assign a cohort to a class/timeslot cell"`
	Unassign       unassignCmd       `cmd:"" help:"clear a class/timeslot cell"`
	Grid           gridCmd           `cmd:"" help:"full master schedule grid"`
	ListClasses    listClassesCmd    `cmd:"" help:"list all classes"`
	ListTimeslots  listTimeslotsCmd  `cmd:"" help:"list all timeslots"`
	ListCohorts    listCohortsCmd    `cmd:"" help:"list all cohorts"`
	ShowRules      showRulesCmd      `cmd:"" help:"dump all rules"`
	Validate       validateCmd       `cmd:"" help:"check schedule for violations"`
	ListUnassigned listUnassignedCmd `cmd:"" help:"list unassigned cells"`
	Report         reportCmd         `cmd:"" help:"per-cohort calendar"`
}

// withDB is embedded by every command that opens an existing schedule. Its
// AfterApply hook opens the store and binds it for injection into Run.
type withDB struct {
	DB string `short:"d" required:"" type:"existingfile" help:"schedule database file"`
}

func (w *withDB) AfterApply(kctx *kong.Context, reg *closeReg) error {
	s, err := store.Open(w.DB)
	if err != nil {
		return err
	}
	reg.add(s)
	kctx.Bind(s)
	return nil
}

type closeReg struct{ stores []*store.Store }

func (c *closeReg) add(s *store.Store) { c.stores = append(c.stores, s) }
func (c *closeReg) closeAll() {
	for _, s := range c.stores {
		_ = s.Close()
	}
}

// jsonFlag is embedded by read commands that support --json output.
type jsonFlag struct {
	JSON bool `name:"json" help:"emit JSON"`
}

func run(args []string) error {
	cli := &CLI{}
	reg := &closeReg{}
	parser, err := kong.New(cli,
		kong.Name("scheduler"),
		kong.Description("a self-contained schedule per .db file"),
		kong.Bind(reg),
		kong.UsageOnError(),
	)
	if err != nil {
		return err
	}
	kctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	defer reg.closeAll()
	return kctx.Run()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if !errors.Is(err, errSilent) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
