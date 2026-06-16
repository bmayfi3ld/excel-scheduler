// Package store owns a schedule's SQLite file: creating/opening it, the
// embedded schema, CRUD, the Load() transform into engine inputs, and the
// shared operation methods that both the CLI and (Phase 2) the MCP server call.
//
// The guiding model is that each .db file is one self-contained schedule, used
// exactly like an .xlsx file — one per year, copy to draft a new class. There
// is no registry and no global state: every operation names the file it acts
// on, so the filesystem is the document model.
//
// Operation methods return structured data (DTOs / engine types), never
// pre-formatted text, so the same method backs a CLI command and an MCP tool.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo)
)

//go:embed schema.sql
var schemaFS embed.FS

// SchemaVersion is the schema the current binary understands. Open refuses
// files written by a newer version; there is no migration framework yet.
const SchemaVersion = 1

// Store wraps the database handle for one schedule file.
type Store struct {
	db *sql.DB
}

// dsn builds the connection string, enabling foreign-key enforcement (which is
// off by default in SQLite and is per-connection, so it lives here).
func dsn(path string) string {
	return fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", path)
}

// Create makes a new schedule file at path and seeds its metadata. It is an
// error if the file already exists, so a schedule is never silently clobbered.
func Create(path, name string) (*Store, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("file already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, err
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.ExecContext(context.Background(), string(schema)); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	now := nowStamp()
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO meta (id, name, schema_version, one_class_at_a_time, created_at, updated_at)
		 VALUES (1, ?, ?, 0, ?, ?)`,
		name, SchemaVersion, now, now,
	)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seeding metadata: %w", err)
	}

	return &Store{db: db}, nil
}

// Open opens an existing schedule file, verifying it is a scheduler database
// and not from a newer schema version than this binary supports.
func Open(path string) (*Store, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("cannot open schedule %q: %w", path, err)
	}

	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, err
	}

	var version int
	if err := db.QueryRowContext(context.Background(), `SELECT schema_version FROM meta WHERE id = 1`).Scan(&version); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%q is not a scheduler database: %w", path, err)
	}
	if version > SchemaVersion {
		_ = db.Close()
		return nil, fmt.Errorf("schedule schema version %d is newer than supported version %d", version, SchemaVersion)
	}

	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// The exec/query/queryRow/begin helpers funnel every DB call through the
// context-aware variants (with a background context, since the store's public
// API is context-free per the plan), keeping call sites terse.
func (s *Store) exec(query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(context.Background(), query, args...)
}

func (s *Store) query(query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(context.Background(), query, args...)
}

func (s *Store) queryRow(query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(context.Background(), query, args...)
}

func (s *Store) begin() (*sql.Tx, error) {
	return s.db.BeginTx(context.Background(), nil)
}

// touch bumps meta.updated_at to now; called after every mutating operation.
func (s *Store) touch() error {
	_, err := s.exec(`UPDATE meta SET updated_at = ? WHERE id = 1`, nowStamp())
	return err
}

// nowStamp returns the current UTC time in RFC3339, the format stored in the
// created_at/updated_at columns.
func nowStamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
