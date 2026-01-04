package sqlformatter

import "testing"

func supportsOnConflict(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("supports INSERT ON CONFLICT", func(t *testing.T) {
		result := format("INSERT INTO tbl VALUES (1,'Blah') ON CONFLICT DO NOTHING;")
		expected := dedent(`
			INSERT INTO
			  tbl
			VALUES
			  (1, 'Blah')
			ON CONFLICT DO NOTHING;
		`)
		assertEqual(t, result, expected)
	})
}
