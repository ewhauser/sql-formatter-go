package sqlformatter

import "testing"

var standardSetOperations = []string{
	"UNION",
	"UNION ALL",
	"UNION DISTINCT",
	"EXCEPT",
	"EXCEPT ALL",
	"EXCEPT DISTINCT",
	"INTERSECT",
	"INTERSECT ALL",
	"INTERSECT DISTINCT",
}

func supportsSetOperations(t *testing.T, format FormatFn, operations ...[]string) {
	t.Helper()
	ops := standardSetOperations
	if len(operations) > 0 {
		ops = operations[0]
	}
	for _, op := range ops {
		op := op
		t.Run("formats "+op, func(t *testing.T) {
			result := format("SELECT * FROM foo " + op + " SELECT * FROM bar;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  foo
				` + op + `
				SELECT
				  *
				FROM
				  bar;
			`)
			assertEqual(t, result, expected)
		})

		t.Run("formats "+op+" inside subquery", func(t *testing.T) {
			result := format("SELECT * FROM (SELECT * FROM foo " + op + " SELECT * FROM bar) AS tbl;")
			expected := dedent(`
				SELECT
				  *
				FROM
				  (
				    SELECT
				      *
				    FROM
				      foo
				    ` + op + `
				    SELECT
				      *
				    FROM
				      bar
				  ) AS tbl;
			`)
			assertEqual(t, result, expected)
		})
	}
}
