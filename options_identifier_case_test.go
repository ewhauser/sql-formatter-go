package sqlformatter

import "testing"

func supportsIdentifierCase(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("preserves identifier case by default", func(t *testing.T) {
		result := format(dedent("select Abc, 'mytext' as MyText from tBl1 left join Tbl2 where colA > 1 and colB = 3"))
		expected := dedent(`
			select
			  Abc,
			  'mytext' as MyText
			from
			  tBl1
			  left join Tbl2
			where
			  colA > 1
			  and colB = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts identifiers to uppercase", func(t *testing.T) {
		result := format(dedent("select Abc, 'mytext' as MyText from tBl1 left join Tbl2 where colA > 1 and colB = 3"), FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			select
			  ABC,
			  'mytext' as MYTEXT
			from
			  TBL1
			  left join TBL2
			where
			  COLA > 1
			  and COLB = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts identifiers to lowercase", func(t *testing.T) {
		result := format(dedent("select Abc, 'mytext' as MyText from tBl1 left join Tbl2 where colA > 1 and colB = 3"), FormatOptions{IdentifierCase: KeywordCaseLower})
		expected := dedent(`
			select
			  abc,
			  'mytext' as mytext
			from
			  tbl1
			  left join tbl2
			where
			  cola > 1
			  and colb = 3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not uppercase quoted identifiers", func(t *testing.T) {
		result := format("select \"abc\" as foo", FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			select
			  "abc" as FOO
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts multi-part identifiers to uppercase", func(t *testing.T) {
		result := format("select Abc from Part1.Part2.Part3", FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			select
			  ABC
			from
			  PART1.PART2.PART3
		`)
		assertEqual(t, result, expected)
	})

	t.Run("function names not affected by identifierCase", func(t *testing.T) {
		result := format("select count(*) from tbl", FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			select
			  count(*)
			from
			  TBL
		`)
		assertEqual(t, result, expected)
	})
}
