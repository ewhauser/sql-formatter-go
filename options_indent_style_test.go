package sqlformatter

import "testing"

func supportsIndentStyle(t *testing.T, format FormatFn) {
	t.Helper()
	baseQuery := `
    SELECT COUNT(a.column1), MAX(b.column2 + b.column3), b.column4 AS four
    FROM ( SELECT column1, column5 FROM table1 ) a
    JOIN table2 b ON a.column5 = b.column5
    WHERE column6 AND column7
    GROUP BY column4;
  `

	t.Run("supports standard mode", func(t *testing.T) {
		result := format(baseQuery, FormatOptions{IndentStyle: IndentStyleStandard})
		expected := dedent(`
			SELECT
			  COUNT(a.column1),
			  MAX(b.column2 + b.column3),
			  b.column4 AS four
			FROM
			  (
			    SELECT
			      column1,
			      column5
			    FROM
			      table1
			  ) a
			  JOIN table2 b ON a.column5 = b.column5
			WHERE
			  column6
			  AND column7
			GROUP BY
			  column4;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("indentStyle tabularLeft aligns clause keywords", func(t *testing.T) {
		result := format(baseQuery, FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    COUNT(a.column1),
			          MAX(b.column2 + b.column3),
			          b.column4 AS four
			FROM      (
			          SELECT    column1,
			                    column5
			          FROM      table1
			          ) a
			JOIN      table2 b ON a.column5 = b.column5
			WHERE     column6
			AND       column7
			GROUP BY  column4;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("tabularLeft handles long keywords", func(t *testing.T) {
		result := format(dedent(`
            SELECT *
            FROM a
            UNION ALL
            SELECT *
            FROM b
            LEFT OUTER JOIN c;
          `), FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    *
			FROM      a
			UNION ALL
			SELECT    *
			FROM      b
			LEFT      OUTER JOIN c;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("tabularLeft indents set operations inside subqueries", func(t *testing.T) {
		result := format(`SELECT * FROM (
            SELECT * FROM a
            UNION ALL
            SELECT * FROM b) AS tbl;`, FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    *
			FROM      (
			          SELECT    *
			          FROM      a
			          UNION ALL
			          SELECT    *
			          FROM      b
			          ) AS tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("tabularLeft handles multiple levels of nested queries", func(t *testing.T) {
		result := format("SELECT age FROM (SELECT fname, lname, age FROM (SELECT fname, lname FROM persons) JOIN (SELECT age FROM ages)) as mytable;", FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    age
			FROM      (
			          SELECT    fname,
			                    lname,
			                    age
			          FROM      (
			                    SELECT    fname,
			                              lname
			                    FROM      persons
			                    )
			          JOIN      (
			                    SELECT    age
			                    FROM      ages
			                    )
			          ) as mytable;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("tabularLeft does not indent semicolon when newlineBeforeSemicolon true", func(t *testing.T) {
		result := format("SELECT firstname, lastname, age FROM customers;", FormatOptions{IndentStyle: IndentStyleTabularLeft, NewlineBeforeSemicolon: true})
		expected := dedent(`
			SELECT    firstname,
			          lastname,
			          age
			FROM      customers
			;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("tabularLeft formats BETWEEN..AND", func(t *testing.T) {
		result := format("SELECT * FROM tbl WHERE id BETWEEN 1 AND 5000;", FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    *
			FROM      tbl
			WHERE     id BETWEEN 1 AND 5000;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("indentStyle tabularRight aligns clause keywords", func(t *testing.T) {
		result := format(baseQuery, FormatOptions{IndentStyle: IndentStyleTabularRight})
		expected := "   SELECT COUNT(a.column1),\n          MAX(b.column2 + b.column3),\n          b.column4 AS four\n     FROM (\n             SELECT column1,\n                    column5\n               FROM table1\n          ) a\n     JOIN table2 b ON a.column5 = b.column5\n    WHERE column6\n      AND column7\n GROUP BY column4;"
		assertEqual(t, result, expected)
	})

	t.Run("tabularRight handles long keywords", func(t *testing.T) {
		result := format(dedent(`
            SELECT *
            FROM a
            UNION ALL
            SELECT *
            FROM b
            LEFT OUTER JOIN c;
          `), FormatOptions{IndentStyle: IndentStyleTabularRight})
		expected := "   SELECT *\n     FROM a\nUNION ALL\n   SELECT *\n     FROM b\n     LEFT OUTER JOIN c;"
		assertEqual(t, result, expected)
	})

	t.Run("tabularRight formats BETWEEN..AND", func(t *testing.T) {
		result := format("SELECT * FROM tbl WHERE id BETWEEN 1 AND 5000;", FormatOptions{IndentStyle: IndentStyleTabularRight})
		expected := "   SELECT *\n     FROM tbl\n    WHERE id BETWEEN 1 AND 5000;"
		assertEqual(t, result, expected)
	})
}
