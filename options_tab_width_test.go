package sqlformatter

import "testing"

func supportsTabWidth(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("indents with 2 spaces by default", func(t *testing.T) {
		result := format("SELECT count(*),Column1 FROM Table1;")
		expected := dedent(`
			SELECT
			  count(*),
			  Column1
			FROM
			  Table1;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports indenting with 4 spaces", func(t *testing.T) {
		result := format("SELECT count(*),Column1 FROM Table1;", FormatOptions{TabWidth: 4})
		expected := dedent(`
			SELECT
			    count(*),
			    Column1
			FROM
			    Table1;
		`)
		assertEqual(t, result, expected)
	})
}
