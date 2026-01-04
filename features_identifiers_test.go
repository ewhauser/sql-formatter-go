package sqlformatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func supportsIdentifiers(t *testing.T, format FormatFn, formatErr FormatErrFn, identifierTypes []string) {
	t.Helper()

	if containsString(identifierTypes, "\"\"-qq") {
		t.Run("supports double-quoted identifiers", func(t *testing.T) {
			assertEqual(t, format(`"foo JOIN bar"`), `"foo JOIN bar"`)
			expected := dedent(`
				SELECT
				  "where"
				FROM
				  "update"
			`)
			assertEqual(t, format(`SELECT "where" FROM "update"`), expected)
		})

		t.Run("no space around dot between double-quoted identifiers", func(t *testing.T) {
			result := format(`SELECT "my table"."col name";`)
			expected := dedent(`
				SELECT
				  "my table"."col name";
			`)
			assertEqual(t, result, expected)
		})

		t.Run("supports escaping double-quote by doubling it", func(t *testing.T) {
			assertEqual(t, format(`"foo""bar"`), `"foo""bar"`)
		})

		t.Run("does not support escaping double-quote with backslash", func(t *testing.T) {
			err := formatErr(`"foo \" JOIN bar"`)
			require.Error(t, err)
			require.Contains(t, err.Error(), "Parse error")
		})
	}

	if containsString(identifierTypes, "``") {
		t.Run("supports backtick-quoted identifiers", func(t *testing.T) {
			assertEqual(t, format("`foo JOIN bar`"), "`foo JOIN bar`")
			expected := dedent("\n\t\t\tSELECT\n\t\t\t  `where`\n\t\t\tFROM\n\t\t\t  `update`\n\t\t\t")
			assertEqual(t, format("SELECT `where` FROM `update`"), expected)
		})

		t.Run("supports escaping backtick by doubling it", func(t *testing.T) {
			assertEqual(t, format("`foo `` JOIN bar`"), "`foo `` JOIN bar`")
		})

		t.Run("no space around dot between backtick-quoted identifiers", func(t *testing.T) {
			result := format("SELECT `my table`.`col name`;")
			expected := dedent("\n\t\t\tSELECT\n\t\t\t  `my table`.`col name`;\n\t\t\t")
			assertEqual(t, result, expected)
		})
	}

	if containsString(identifierTypes, "U&\"\"") {
		t.Run("supports unicode double-quoted identifiers", func(t *testing.T) {
			assertEqual(t, format(`U&"foo JOIN bar"`), `U&"foo JOIN bar"`)
			expected := dedent(`
				SELECT
				  U&"where"
				FROM
				  U&"update"
			`)
			assertEqual(t, format(`SELECT U&"where" FROM U&"update"`), expected)
		})

		t.Run("no space around dot between unicode double-quoted identifiers", func(t *testing.T) {
			result := format(`SELECT U&"my table".U&"col name";`)
			expected := dedent(`
				SELECT
				  U&"my table".U&"col name";
			`)
			assertEqual(t, result, expected)
		})

		t.Run("supports escaping in U&\"\" strings by repeated quote", func(t *testing.T) {
			assertEqual(t, format(`U&"foo "" JOIN bar"`), `U&"foo "" JOIN bar"`)
		})

		t.Run("detects consecutive U&\"\" identifiers as separate ones", func(t *testing.T) {
			assertEqual(t, format(`U&"foo"U&"bar"`), `U&"foo" U&"bar"`)
		})

		t.Run("does not support escaping in U&\"\" strings with backslash", func(t *testing.T) {
			err := formatErr(`U&"foo \" JOIN bar"`)
			require.Error(t, err)
			require.Contains(t, err.Error(), "Parse error")
		})
	}

	if containsString(identifierTypes, "[]") {
		t.Run("supports bracket-quoted identifiers", func(t *testing.T) {
			assertEqual(t, format("[foo JOIN bar]"), "[foo JOIN bar]")
			expected := dedent(`
				SELECT
				  [where]
				FROM
				  [update]
			`)
			assertEqual(t, format("SELECT [where] FROM [update]"), expected)
		})

		t.Run("supports escaping close-bracket by doubling it", func(t *testing.T) {
			assertEqual(t, format("[foo ]] JOIN bar]"), "[foo ]] JOIN bar]")
		})

		t.Run("no space around dot between bracket-quoted identifiers", func(t *testing.T) {
			result := format("SELECT [my table].[col name];")
			expected := dedent(`
				SELECT
				  [my table].[col name];
			`)
			assertEqual(t, result, expected)
		})
	}
}
