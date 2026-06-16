package store

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Copy makes an independent duplicate of a schedule file — the first-class
// "branch this schedule" operation behind the multi-file model. It refuses to
// overwrite an existing destination, verifies the source is a real scheduler
// database, byte-copies it, then refreshes the copy's created_at (and name, if
// given) so the new file reads as a fresh schedule rather than the original.
//
// Because schedule files are always at rest between commands, a plain `cp`
// works too; Copy adds the verify + metadata refresh.
func Copy(src, dst, name string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("file already exists: %s", dst)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Verify src is a valid scheduler database before copying.
	s, err := Open(src)
	if err != nil {
		return err
	}
	if err := s.Close(); err != nil {
		return err
	}

	if err := copyFile(src, dst); err != nil {
		return err
	}

	d, err := Open(dst)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()

	now := nowStamp()
	if name != "" {
		_, err = d.exec(`UPDATE meta SET name = ?, created_at = ?, updated_at = ? WHERE id = 1`, name, now, now)
	} else {
		_, err = d.exec(`UPDATE meta SET created_at = ?, updated_at = ? WHERE id = 1`, now, now)
	}
	return err
}

// copyFile byte-copies src to a freshly created dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // src is the user-named schedule file to copy
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst) //nolint:gosec // dst is the user-named destination file
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
