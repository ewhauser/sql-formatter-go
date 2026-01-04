package sqlformatter

import "testing"

func supportsBetween(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats BETWEEN on single line", func(t *testing.T) {
		result := format("foo BETWEEN bar AND baz")
		assertEqual(t, result, "foo BETWEEN bar AND baz")
	})

	t.Run("supports qualified names as BETWEEN values", func(t *testing.T) {
		result := format("foo BETWEEN t.bar AND t.baz")
		assertEqual(t, result, "foo BETWEEN t.bar AND t.baz")
	})

	t.Run("formats BETWEEN with comments", func(t *testing.T) {
		result := format("WHERE foo BETWEEN /*C1*/ t.bar /*C2*/ AND /*C3*/ t.baz")
		expected := dedent(`
			WHERE
			  foo BETWEEN /*C1*/ t.bar /*C2*/ AND /*C3*/ t.baz
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports complex expressions inside BETWEEN", func(t *testing.T) {
		result := format("foo BETWEEN 1+2 AND 3+4")
		assertEqual(t, result, "foo BETWEEN 1 + 2 AND 3  + 4")
	})

	t.Run("supports CASE inside BETWEEN", func(t *testing.T) {
		result := format("foo BETWEEN CASE x WHEN 1 THEN 2 END AND 3")
		expected := dedent(`
			foo BETWEEN CASE x
			  WHEN 1 THEN 2
			END AND 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports AND after BETWEEN", func(t *testing.T) {
		result := format("SELECT foo BETWEEN 1 AND 2 AND x > 10")
		expected := dedent(`
			SELECT
			  foo BETWEEN 1 AND 2
			  AND x > 10
		`)
		assertEqual(t, result, expected)
	})
}
