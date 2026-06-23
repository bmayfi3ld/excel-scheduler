package store

import (
	"sort"

	"github.com/bmayfi3ld/quilt/pkg/engine"
)

// Load reconstructs the engine inputs from the relational tables: an ordered
// Grid (classes × timeslots with filled cells) and the full Rules set. This is
// the relational → engine transform deferred from Phase 0; every validation /
// report method funnels through it so the engine sees exactly one shape.
func (s *Store) Load() (engine.Grid, engine.Rules, error) {
	var g engine.Grid
	var r engine.Rules

	classes, classOf, err := s.loadAxis(`SELECT id, name FROM class ORDER BY sort_order, id`)
	if err != nil {
		return g, r, err
	}
	timeslots, timeslotOf, err := s.loadAxis(`SELECT id, label FROM timeslot ORDER BY sort_order, id`)
	if err != nil {
		return g, r, err
	}
	g.Classes = classes
	g.Timeslots = timeslots

	if g.Cells, err = s.loadCells(classOf, timeslotOf, len(classes), len(timeslots)); err != nil {
		return g, r, err
	}

	if r.AllCohorts, err = s.cohortNames(); err != nil {
		return g, r, err
	}
	if r.ClassRequiresTravel, err = s.loadTravel(); err != nil {
		return g, r, err
	}
	if r.CohortBlacklist, err = s.loadBlackout(); err != nil {
		return g, r, err
	}

	oneClass, err := s.oneClassAtATime()
	if err != nil {
		return g, r, err
	}
	r.OneClassAtATime = oneClass

	return g, r, nil
}

// loadAxis reads an ordered list of (id, label) rows and returns the labels in
// order plus a map from row id to its position, used to place assignment cells.
// Both the class and timeslot axes share this shape.
func (s *Store) loadAxis(query string) ([]string, map[int64]int, error) {
	rows, err := s.query(query)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	var labels []string
	indexOf := map[int64]int{}
	for rows.Next() {
		var id int64
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			return nil, nil, err
		}
		indexOf[id] = len(labels)
		labels = append(labels, label)
	}
	return labels, indexOf, rows.Err()
}

// loadCells builds the row-major cell matrix from the assignment table. Cells
// with no assignment stay "" (engine's unassigned marker). Assignments whose
// class or timeslot is somehow absent from the axes are skipped defensively.
func (s *Store) loadCells(classOf, timeslotOf map[int64]int, nRows, nCols int) ([][]string, error) {
	cells := make([][]string, nRows)
	for i := range cells {
		cells[i] = make([]string, nCols)
	}

	rows, err := s.query(`SELECT class_id, timeslot_id, cohort_value FROM assignment`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var classID, timeslotID int64
		var value string
		if err := rows.Scan(&classID, &timeslotID, &value); err != nil {
			return nil, err
		}
		row, okRow := classOf[classID]
		col, okCol := timeslotOf[timeslotID]
		if okRow && okCol {
			cells[row][col] = value
		}
	}
	return cells, rows.Err()
}

// loadTravel groups travel_group rows into one TravelRule per class, with the
// buildings ordered by building_index. Cohort names within a building keep
// their insertion (rowid) order.
func (s *Store) loadTravel() ([]engine.TravelRule, error) {
	rows, err := s.query(
		`SELECT c.name, tg.building_index, tg.cohort_name
		 FROM travel_group tg JOIN class c ON c.id = tg.class_id
		 ORDER BY c.sort_order, c.id, tg.building_index, tg.rowid`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var order []string // class names in first-seen order
	byClass := map[string]map[int][]string{}
	for rows.Next() {
		var class, cohort string
		var building int
		if err := rows.Scan(&class, &building, &cohort); err != nil {
			return nil, err
		}
		if _, ok := byClass[class]; !ok {
			byClass[class] = map[int][]string{}
			order = append(order, class)
		}
		byClass[class][building] = append(byClass[class][building], cohort)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var rules []engine.TravelRule
	for _, class := range order {
		groups := byClass[class]
		indexes := make([]int, 0, len(groups))
		for i := range groups {
			indexes = append(indexes, i)
		}
		sort.Ints(indexes)

		buildings := make([][]string, 0, len(indexes))
		for _, i := range indexes {
			buildings = append(buildings, groups[i])
		}
		rules = append(rules, engine.TravelRule{Class: class, Buildings: buildings})
	}
	return rules, nil
}

// loadBlackout groups blackout rows into one BlacklistRule per cohort, with the
// blocked timeslots ordered by their column order.
func (s *Store) loadBlackout() ([]engine.BlacklistRule, error) {
	rows, err := s.query(
		`SELECT b.cohort_name, t.label
		 FROM blackout b JOIN timeslot t ON t.id = b.timeslot_id
		 ORDER BY b.cohort_name, t.sort_order, t.id`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var order []string
	byCohort := map[string][]string{}
	for rows.Next() {
		var cohort, label string
		if err := rows.Scan(&cohort, &label); err != nil {
			return nil, err
		}
		if _, ok := byCohort[cohort]; !ok {
			order = append(order, cohort)
		}
		byCohort[cohort] = append(byCohort[cohort], label)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var rules []engine.BlacklistRule
	for _, cohort := range order {
		rules = append(rules, engine.BlacklistRule{Cohort: cohort, Timeslots: byCohort[cohort]})
	}
	return rules, nil
}

// cohortNames returns the AllCohorts master list in display order.
func (s *Store) cohortNames() ([]string, error) {
	rows, err := s.query(`SELECT name FROM cohort ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// oneClassAtATime reads the OneClassAtATime toggle from meta.
func (s *Store) oneClassAtATime() (bool, error) {
	var v int
	if err := s.queryRow(`SELECT one_class_at_a_time FROM meta WHERE id = 1`).Scan(&v); err != nil {
		return false, err
	}
	return v != 0, nil
}
