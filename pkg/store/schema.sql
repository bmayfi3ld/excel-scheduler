-- Schema for a single, self-contained schedule (one schedule per .db file).
-- Foreign keys are enforced via PRAGMA foreign_keys(1) set in the connection
-- DSN (see store.go), not here, since PRAGMAs are per-connection.

-- Singleton schedule metadata. id is pinned to 1 so the row is a singleton and
-- the OneClassAtATime toggle lives alongside the name/timestamps.
CREATE TABLE meta (
    id                  INTEGER PRIMARY KEY CHECK (id = 1),
    name                TEXT    NOT NULL,
    schema_version      INTEGER NOT NULL,
    one_class_at_a_time INTEGER NOT NULL DEFAULT 0,
    created_at          TEXT    NOT NULL,
    updated_at          TEXT    NOT NULL
);

-- Grid rows. sort_order drives display order.
CREATE TABLE class (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL
);

-- Grid columns. sort_order drives display order AND the ClassRequiresTravel
-- adjacency (a class can't cross buildings between two consecutive columns).
CREATE TABLE timeslot (
    id         INTEGER PRIMARY KEY,
    label      TEXT    NOT NULL UNIQUE,
    day        TEXT,
    period     TEXT,
    sort_order INTEGER NOT NULL
);

-- The AllCohorts master list of valid cohort names.
CREATE TABLE cohort (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL
);

-- A filled cell. cohort_value is FREE TEXT (not a FK to cohort): the engine's
-- AllCohorts rule must be able to flag values absent from the master list, and
-- exempt placeholders like "#### closed" are valid cell values but not cohorts.
CREATE TABLE assignment (
    class_id     INTEGER NOT NULL REFERENCES class(id)    ON DELETE CASCADE,
    timeslot_id  INTEGER NOT NULL REFERENCES timeslot(id) ON DELETE CASCADE,
    cohort_value TEXT    NOT NULL,
    PRIMARY KEY (class_id, timeslot_id)
);

-- ClassRequiresTravel building groupings, per class. cohort_name is free text,
-- matched by name like the engine does.
CREATE TABLE travel_group (
    class_id       INTEGER NOT NULL REFERENCES class(id) ON DELETE CASCADE,
    building_index INTEGER NOT NULL,
    cohort_name    TEXT    NOT NULL
);

-- CohortBlacklist: a cohort may not be scheduled in the listed timeslot.
-- cohort_name is free text, matched by name like the engine does.
CREATE TABLE blackout (
    cohort_name TEXT    NOT NULL,
    timeslot_id INTEGER NOT NULL REFERENCES timeslot(id) ON DELETE CASCADE,
    PRIMARY KEY (cohort_name, timeslot_id)
);
