package sqlformatter

import "testing"

type truncateTableConfig struct {
	WithTable    bool
	WithoutTable bool
}

func supportsTruncateTable(t *testing.T, format FormatFn, cfg truncateTableConfig) {
	t.Helper()
	withTable := cfg.WithTable
	if !cfg.WithoutTable && !cfg.WithTable {
		withTable = true
	}

	if withTable {
		t.Run("formats TRUNCATE TABLE statement", func(t *testing.T) {
			result := format("TRUNCATE TABLE Customers;")
			expected := dedent(`
				TRUNCATE TABLE Customers;
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.WithoutTable {
		t.Run("formats TRUNCATE statement without TABLE", func(t *testing.T) {
			result := format("TRUNCATE Customers;")
			expected := dedent(`
				TRUNCATE Customers;
			`)
			assertEqual(t, result, expected)
		})
	}
}
