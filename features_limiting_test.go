package sqlformatter

import "testing"

type limitingConfig struct {
	Limit     bool
	Offset    bool
	FetchFirst bool
	FetchNext bool
}

func supportsLimiting(t *testing.T, format FormatFn, types limitingConfig) {
	t.Helper()
	if types.Limit {
		t.Run("formats LIMIT with two comma-separated values", func(t *testing.T) {
			result := format("SELECT * FROM tbl LIMIT 5, 10;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				LIMIT
				  5, 10;
			`)
			assertEqual(t, result, expected)
		})

		t.Run("formats LIMIT with complex expressions", func(t *testing.T) {
			result := format("SELECT * FROM tbl LIMIT abs(-5) - 1, (2 + 3) * 5;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				LIMIT
				  abs(-5) - 1, (2 + 3) * 5;
			`)
			assertEqual(t, result, expected)
		})

		t.Run("formats LIMIT with comments", func(t *testing.T) {
			result := format("SELECT * FROM tbl LIMIT --comment\n 5,--comment\n6;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				LIMIT --comment
				  5, --comment
				  6;
			`)
			assertEqual(t, result, expected)
		})

		t.Run("formats LIMIT in tabular style", func(t *testing.T) {
			result := format("SELECT * FROM tbl LIMIT 5, 6;", FormatOptions{IndentStyle: IndentStyleTabularLeft})
			expected := dedent(`
				SELECT    *
				FROM      tbl
				LIMIT     5, 6;
			`)
			assertEqual(t, result, expected)
		})
	}

	if types.Limit && types.Offset {
		t.Run("formats LIMIT of single value and OFFSET", func(t *testing.T) {
			result := format("SELECT * FROM tbl LIMIT 5 OFFSET 8;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				LIMIT
				  5
				OFFSET
				  8;
			`)
			assertEqual(t, result, expected)
		})
	}

	if types.FetchFirst {
		t.Run("formats FETCH FIRST", func(t *testing.T) {
			result := format("SELECT * FROM tbl FETCH FIRST 10 ROWS ONLY;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				FETCH FIRST
				  10 ROWS ONLY;
			`)
			assertEqual(t, result, expected)
		})
	}

	if types.FetchNext {
		t.Run("formats FETCH NEXT", func(t *testing.T) {
			result := format("SELECT * FROM tbl FETCH NEXT 1 ROW ONLY;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				FETCH NEXT
				  1 ROW ONLY;
			`)
			assertEqual(t, result, expected)
		})
	}

	if types.FetchFirst && types.Offset {
		t.Run("formats OFFSET FETCH FIRST", func(t *testing.T) {
			result := format("SELECT * FROM tbl OFFSET 250 ROWS FETCH FIRST 5 ROWS ONLY;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				OFFSET
				  250 ROWS
				FETCH FIRST
				  5 ROWS ONLY;
			`)
			assertEqual(t, result, expected)
		})
	}

	if types.FetchNext && types.Offset {
		t.Run("formats OFFSET FETCH NEXT", func(t *testing.T) {
			result := format("SELECT * FROM tbl OFFSET 250 ROWS FETCH NEXT 5 ROWS ONLY;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  tbl
				OFFSET
				  250 ROWS
				FETCH NEXT
				  5 ROWS ONLY;
			`)
			assertEqual(t, result, expected)
		})
	}
}
