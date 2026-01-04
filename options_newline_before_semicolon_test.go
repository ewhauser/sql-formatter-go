package sqlformatter

import "testing"

func supportsNewlineBeforeSemicolon(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats lonely semicolon", func(t *testing.T) {
		result := format(";")
		assertEqual(t, result, ";")
	})

	t.Run("does not add newline before lonely semicolon", func(t *testing.T) {
		result := format(";", FormatOptions{NewlineBeforeSemicolon: true})
		assertEqual(t, result, ";")
	})

	t.Run("defaults to semicolon on end of last line", func(t *testing.T) {
		result := format("SELECT a FROM b;")
		expected := dedent(`
			SELECT
			  a
			FROM
			  b;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("places semicolon on same line as single-line clause", func(t *testing.T) {
		result := format("SELECT a FROM;")
		expected := dedent(`
			SELECT
			  a
			FROM;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports semicolon on separate line", func(t *testing.T) {
		result := format("SELECT a FROM b;", FormatOptions{NewlineBeforeSemicolon: true})
		expected := dedent(`
			SELECT
			  a
			FROM
			  b
			;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats multiple lonely semicolons", func(t *testing.T) {
		result := format(";;;")
		expected := dedent(`
			;

			;

			;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not introduce extra lines between semicolons when newlineBeforeSemicolon true", func(t *testing.T) {
		result := format(";;;", FormatOptions{NewlineBeforeSemicolon: true})
		expected := dedent(`
			;

			;

			;
		`)
		assertEqual(t, result, expected)
	})
}
