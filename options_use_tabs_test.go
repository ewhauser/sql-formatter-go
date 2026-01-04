package sqlformatter

import "testing"

func supportsUseTabs(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("supports indenting with tabs", func(t *testing.T) {
		result := format("SELECT count(*),Column1 FROM Table1;", FormatOptions{UseTabs: true})
		expected := "SELECT\n\tcount(*),\n\tColumn1\nFROM\n\tTable1;"
		assertEqual(t, result, expected)
	})

	t.Run("ignores tabWidth when useTabs enabled", func(t *testing.T) {
		result := format("SELECT count(*),Column1 FROM Table1;", FormatOptions{UseTabs: true, TabWidth: 10})
		expected := "SELECT\n\tcount(*),\n\tColumn1\nFROM\n\tTable1;"
		assertEqual(t, result, expected)
	})
}
