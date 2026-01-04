package sqlformatter

import "testing"

func supportsWindowFunctions(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("supports ROWS BETWEEN in window functions", func(t *testing.T) {
		result := format(`
        SELECT
          RANK() OVER (
            PARTITION BY explosion
            ORDER BY day ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
          ) AS amount
        FROM
          tbl
      `)
		expected := dedent(`
			SELECT
			  RANK() OVER (
			    PARTITION BY
			      explosion
			    ORDER BY
			      day ROWS BETWEEN 6 PRECEDING
			      AND CURRENT ROW
			  ) AS amount
			FROM
			  tbl
		`)
		assertEqual(t, result, expected)
	})
}
