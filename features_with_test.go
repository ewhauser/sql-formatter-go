package sqlformatter

import "testing"

func supportsWith(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats WITH clause with multiple CTE", func(t *testing.T) {
		result := format(`
      WITH
      cte_1 AS (
        SELECT a FROM b WHERE c = 1
      ),
      cte_2 AS (
        SELECT c FROM d WHERE e = 2
      ),
      final AS (
        SELECT * FROM cte_1 LEFT JOIN cte_2 ON b = d
      )
      SELECT * FROM final;
    `)
		expected := dedent(`
			WITH
			  cte_1 AS (
			    SELECT
			      a
			    FROM
			      b
			    WHERE
			      c = 1
			  ),
			  cte_2 AS (
			    SELECT
			      c
			    FROM
			      d
			    WHERE
			      e = 2
			  ),
			  final AS (
			    SELECT
			      *
			    FROM
			      cte_1
			      LEFT JOIN cte_2 ON b = d
			  )
			SELECT
			  *
			FROM
			  final;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats WITH clause with parameterized CTE", func(t *testing.T) {
		result := format(`
      WITH cte_1(id, parent_id) AS (
        SELECT id, parent_id
        FROM tab1
        WHERE parent_id IS NULL
      )
      SELECT id, parent_id FROM cte_1;
    `)
		expected := dedent(`
			WITH
			  cte_1 (id, parent_id) AS (
			    SELECT
			      id,
			      parent_id
			    FROM
			      tab1
			    WHERE
			      parent_id IS NULL
			  )
			SELECT
			  id,
			  parent_id
			FROM
			  cte_1;
		`)
		assertEqual(t, result, expected)
	})
}
