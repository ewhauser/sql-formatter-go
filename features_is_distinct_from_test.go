package sqlformatter

import "testing"

func supportsIsDistinctFrom(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats IS DISTINCT FROM operator", func(t *testing.T) {
		result := format("SELECT x IS DISTINCT FROM y, x IS NOT DISTINCT FROM y")
		expected := dedent(`
			SELECT
			  x IS DISTINCT FROM y,
			  x IS NOT DISTINCT FROM y
		`)
		assertEqual(t, result, expected)
	})
}
