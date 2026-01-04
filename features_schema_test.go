package sqlformatter

import "testing"

func supportsSchema(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats simple SET SCHEMA statements", func(t *testing.T) {
		result := format("SET SCHEMA schema1;")
		expected := dedent(`
			SET SCHEMA schema1;
		`)
		assertEqual(t, result, expected)
	})
}
