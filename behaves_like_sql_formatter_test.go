package sqlformatter

import "testing"

func behavesLikeSqlFormatter(t *testing.T, format FormatFn) {
	t.Helper()
	supportsDisableComment(t, format)
	supportsCase(t, format)
	supportsWith(t, format)

	supportsTabWidth(t, format)
	supportsUseTabs(t, format)
	supportsKeywordCase(t, format)
	supportsIdentifierCase(t, format)
	supportsFunctionCase(t, format)
	supportsIndentStyle(t, format)
	supportsLinesBetweenQueries(t, format)
	supportsExpressionWidth(t, format)
	supportsNewlineBeforeSemicolon(t, format)
	supportsLogicalOperatorNewline(t, format)
	supportsParamTypes(t, format)
	supportsWindowFunctions(t, format)

	t.Run("formats SELECT with asterisks", func(t *testing.T) {
		result := format("SELECT tbl.*, count(*), col1 * col2 FROM tbl;")
		expected := dedent(`
			SELECT
			  tbl.*,
			  count(*),
			  col1 * col2
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats complex SELECT", func(t *testing.T) {
		result := format("SELECT DISTINCT name, ROUND(age/7) field1, 18 + 20 AS field2, 'some string' FROM foo;")
		expected := dedent(`
			SELECT DISTINCT
			  name,
			  ROUND(age / 7) field1,
			  18 + 20 AS field2,
			  'some string'
			FROM
			  foo;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats SELECT with complex WHERE", func(t *testing.T) {
		result := format(`
      SELECT * FROM foo WHERE Column1 = 'testing'
      AND ( (Column2 = Column3 OR Column4 >= ABS(5)) );
    `)
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			WHERE
			  Column1 = 'testing'
			  AND (
			    (
			      Column2 = Column3
			      OR Column4 >= ABS(5)
			    )
			  );
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats SELECT with top level reserved words", func(t *testing.T) {
		result := format(`
      SELECT * FROM foo WHERE name = 'John' GROUP BY some_column
      HAVING column > 10 ORDER BY other_column;
    `)
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			WHERE
			  name = 'John'
			GROUP BY
			  some_column
			HAVING
			  column > 10
			ORDER BY
			  other_column;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("allows keywords as column names in tbl.col syntax", func(t *testing.T) {
		result := format("SELECT mytable.update, mytable.select FROM mytable WHERE mytable.from > 10;")
		expected := dedent(`
			SELECT
			  mytable.update,
			  mytable.select
			FROM
			  mytable
			WHERE
			  mytable.from > 10;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats ORDER BY", func(t *testing.T) {
		result := format(`
      SELECT * FROM foo ORDER BY col1 ASC, col2 DESC;
    `)
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			ORDER BY
			  col1 ASC,
			  col2 DESC;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats SELECT query with SELECT query inside it", func(t *testing.T) {
		result := format("SELECT *, SUM(*) AS total FROM (SELECT * FROM Posts WHERE age > 10) WHERE a > b")
		expected := dedent(`
			SELECT
			  *,
			  SUM(*) AS total
			FROM
			  (
			    SELECT
			      *
			    FROM
			      Posts
			    WHERE
			      age > 10
			  )
			WHERE
			  a > b
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats open paren after comma", func(t *testing.T) {
		result := format("INSERT INTO TestIds (id) VALUES (4),(5), (6),(7),(9),(10),(11);")
		expected := dedent(`
			INSERT INTO
			  TestIds (id)
			VALUES
			  (4),
			  (5),
			  (6),
			  (7),
			  (9),
			  (10),
			  (11);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("keeps short parenthesized list with nested parenthesis on single line", func(t *testing.T) {
		result := format("SELECT (a + b * (c - SIN(1)));")
		expected := dedent(`
			SELECT
			  (a + b * (c - SIN(1)));
		`)
		assertEqual(t, result, expected)
	})

	t.Run("breaks long parenthesized lists to multiple lines", func(t *testing.T) {
		result := format(`
      INSERT INTO some_table (id_product, id_shop, id_currency, id_country, id_registration) (
      SELECT COALESCE(dq.id_discounter_shopping = 2, dq.value, dq.value / 100),
      COALESCE (dq.id_discounter_shopping = 2, 'amount', 'percentage') FROM foo);
    `)
		expected := dedent(`
			INSERT INTO
			  some_table (
			    id_product,
			    id_shop,
			    id_currency,
			    id_country,
			    id_registration
			  ) (
			    SELECT
			      COALESCE(
			        dq.id_discounter_shopping = 2,
			        dq.value,
			        dq.value / 100
			      ),
			      COALESCE(
			        dq.id_discounter_shopping = 2,
			        'amount',
			        'percentage'
			      )
			    FROM
			      foo
			  );
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats top-level multi-word reserved words with inconsistent spacing", func(t *testing.T) {
		result := format("SELECT * FROM foo LEFT \t   \n JOIN mycol ORDER \n BY blah")
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			  LEFT JOIN mycol
			ORDER BY
			  blah
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats long double parenthized queries to multiple lines", func(t *testing.T) {
		result := format("((foo = '0123456789-0123456789-0123456789-0123456789'))")
		expected := dedent(`
			(
			  (
			    foo = '0123456789-0123456789-0123456789-0123456789'
			  )
			)
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats short double parenthized queries to one line", func(t *testing.T) {
		result := format("((foo = 'bar'))")
		assertEqual(t, result, "((foo = 'bar'))")
	})

	t.Run("supports unicode letters in identifiers", func(t *testing.T) {
		result := format("SELECT 结合使用, тест FROM töörõõm;")
		expected := dedent(`
			SELECT
			  结合使用,
			  тест
			FROM
			  töörõõm;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports unicode numbers in identifiers", func(t *testing.T) {
		result := format("SELECT my၁၂၃ FROM tbl༡༢༣;")
		expected := dedent(`
			SELECT
			  my၁၂၃
			FROM
			  tbl༡༢༣;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports unicode diacritical marks in identifiers", func(t *testing.T) {
		result := format("SELECT o\u0303 FROM tbl;")
		expected := dedent("\n\t\t\tSELECT\n\t\t\t  o\u0303\n\t\t\tFROM\n\t\t\t  tbl;\n\t\t\t")
		assertEqual(t, result, expected)
	})
}
