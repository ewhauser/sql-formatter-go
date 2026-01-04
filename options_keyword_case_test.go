package sqlformatter

import "testing"

func supportsKeywordCase(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("preserves keyword case by default", func(t *testing.T) {
		result := format("select distinct * frOM foo left JOIN bar WHERe cola > 1 and colb = 3")
		expected := dedent(`
			select distinct
			  *
			frOM
			  foo
			  left JOIN bar
			WHERe
			  cola > 1
			  and colb = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts keywords to uppercase", func(t *testing.T) {
		result := format("select distinct * frOM foo left JOIN mycol WHERe cola > 1 and colb = 3", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT DISTINCT
			  *
			FROM
			  foo
			  LEFT JOIN mycol
			WHERE
			  cola > 1
			  AND colb = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts keywords to lowercase", func(t *testing.T) {
		result := format("select distinct * frOM foo left JOIN bar WHERe cola > 1 and colb = 3", FormatOptions{KeywordCase: KeywordCaseLower})
		expected := dedent(`
			select distinct
			  *
			from
			  foo
			  left join bar
			where
			  cola > 1
			  and colb = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not uppercase keywords inside strings", func(t *testing.T) {
		result := format("select 'distinct' as foo", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  'distinct' AS foo
		`)
		assertEqual(t, result, expected)
	})

	t.Run("treats dot-separated keywords as identifiers", func(t *testing.T) {
		result := format("select table.and from set.select", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  table.and
			FROM
			  set.select
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats multi-word reserved clauses into single line", func(t *testing.T) {
		result := format(`select * from mytable
      inner
      join
      mytable2 on mytable1.col1 = mytable2.col1
      where mytable2.col1 = 5
      group
      bY mytable1.col2
      order
      by
      mytable2.col3;`, FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  *
			FROM
			  mytable
			  INNER JOIN mytable2 ON mytable1.col1 = mytable2.col1
			WHERE
			  mytable2.col1 = 5
			GROUP BY
			  mytable1.col2
			ORDER BY
			  mytable2.col3;
		`)
		assertEqual(t, result, expected)
	})
}
