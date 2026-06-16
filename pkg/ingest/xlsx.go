package ingest

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ReadWorkbook opens path and returns the named sheets as trimmed string grids.
// This is the only file in pkg/ingest that imports excelize.
func ReadWorkbook(path, scheduleSheet, rulesSheet string) (schedule, rules Sheet, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	schedule, err = readSheet(f, scheduleSheet)
	if err != nil {
		return nil, nil, err
	}
	rules, err = readSheet(f, rulesSheet)
	if err != nil {
		return nil, nil, err
	}
	return schedule, rules, nil
}

func readSheet(f *excelize.File, name string) (Sheet, error) {
	rows, err := f.GetRows(name)
	if err != nil {
		return nil, fmt.Errorf("reading sheet %q: %w", name, err)
	}
	sheet := make(Sheet, len(rows))
	for i, row := range rows {
		cells := make([]string, len(row))
		for j, cell := range row {
			cells[j] = strings.TrimSpace(cell)
		}
		sheet[i] = cells
	}
	return sheet, nil
}
