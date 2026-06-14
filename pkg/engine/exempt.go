package engine

// isExempt reports whether a cell value should be skipped from all rule checks.
// A value is exempt when its first four characters are identical (e.g. "####",
// "@@@@ no class"), a convention used to mark blacked-out times that hold a
// placeholder rather than a real cohort. Values shorter than four characters
// are never exempt.
//
// Only the first four bytes are compared, matching the original add-in's
// substring(0,4) check; this is byte-wise, so multi-byte runes are not
// special-cased.
func isExempt(cohort string) bool {
	if len(cohort) < 4 {
		return false
	}
	first := cohort[0]
	return cohort[1] == first && cohort[2] == first && cohort[3] == first
}
