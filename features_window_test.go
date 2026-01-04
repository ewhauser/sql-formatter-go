package sqlformatter

import "testing"

func supportsWindow(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats WINDOW clause at top level", func(t *testing.T) {
		result := format("SELECT *, ROW_NUMBER() OVER wnd AS next_value FROM tbl WINDOW wnd AS (PARTITION BY id ORDER BY time);")
		expected := dedent(`
			SELECT
			  *,
			  ROW_NUMBER() OVER wnd AS next_value
			FROM
			  tbl
			WINDOW
			  wnd AS (
			    PARTITION BY
			      id
			    ORDER BY
			      time
			  );
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats multiple WINDOW specifications", func(t *testing.T) {
		result := format("SELECT * FROM table1 WINDOW w1 AS (PARTITION BY col1), w2 AS (PARTITION BY col1, col2);")
		expected := dedent(`
			SELECT
			  *
			FROM
			  table1
			WINDOW
			  w1 AS (
			    PARTITION BY
			      col1
			  ),
			  w2 AS (
			    PARTITION BY
			      col1,
			      col2
			  );
		`)
		assertEqual(t, result, expected)
	})
}
