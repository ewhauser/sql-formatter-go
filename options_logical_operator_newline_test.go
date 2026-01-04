package sqlformatter

import "testing"

func supportsLogicalOperatorNewline(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("adds newline before logical operator by default", func(t *testing.T) {
		result := format("SELECT a WHERE true AND false;")
		expected := dedent(`
			SELECT
			  a
			WHERE
			  true
			  AND false;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports newline after logical operator", func(t *testing.T) {
		result := format("SELECT a WHERE true AND false;", FormatOptions{LogicalOperatorNewline: LogicalOperatorNewlineAfter})
		expected := dedent(`
			SELECT
			  a
			WHERE
			  true AND
			  false;
		`)
		assertEqual(t, result, expected)
	})
}
