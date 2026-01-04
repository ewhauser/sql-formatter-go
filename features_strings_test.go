package sqlformatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func supportsStrings(t *testing.T, format FormatFn, formatErr FormatErrFn, stringTypes []string) {
	t.Helper()

	if containsString(stringTypes, "\"\"-qq") || containsString(stringTypes, "\"\"-bs") {
		t.Run("supports double-quoted strings", func(t *testing.T) {
			assertEqual(t, format(`"foo JOIN bar"`), `"foo JOIN bar"`)
			expected := dedent(`
				SELECT
				  "where"
				FROM
				  "update"
			`)
			assertEqual(t, format(`SELECT "where" FROM "update"`), expected)
		})
	}

	if containsString(stringTypes, "\"\"-qq") {
		t.Run("supports escaping double-quote by doubling it", func(t *testing.T) {
			assertEqual(t, format(`"foo""bar"`), `"foo""bar"`)
		})

		if !containsString(stringTypes, "\"\"-bs") {
			t.Run("does not support escaping double-quote with backslash", func(t *testing.T) {
				err := formatErr(`"foo \" JOIN bar"`)
				require.Error(t, err)
				require.Contains(t, err.Error(), "Parse error")
			})
		}
	}

	if containsString(stringTypes, "\"\"-bs") {
		t.Run("supports escaping double-quote with backslash", func(t *testing.T) {
			assertEqual(t, format(`"foo \" JOIN bar"`), `"foo \" JOIN bar"`)
		})

		if !containsString(stringTypes, "\"\"-qq") {
			t.Run("does not support escaping double-quote by doubling it", func(t *testing.T) {
				assertEqual(t, format(`"foo "" JOIN bar"`), `"foo " " JOIN bar"`)
			})
		}
	}

	if containsString(stringTypes, "''-qq") || containsString(stringTypes, "''-bs") {
		t.Run("supports single-quoted strings", func(t *testing.T) {
			assertEqual(t, format(`'foo JOIN bar'`), `'foo JOIN bar'`)
			expected := dedent(`
				SELECT
				  'where'
				FROM
				  'update'
			`)
			assertEqual(t, format(`SELECT 'where' FROM 'update'`), expected)
		})
	}

	if containsString(stringTypes, "''-qq") {
		t.Run("supports escaping single-quote by doubling it", func(t *testing.T) {
			assertEqual(t, format(`'foo''bar'`), `'foo''bar'`)
		})

		if !containsString(stringTypes, "''-bs") {
			t.Run("does not support escaping single-quote with backslash", func(t *testing.T) {
				err := formatErr(`'foo \' JOIN bar'`)
				require.Error(t, err)
				require.Contains(t, err.Error(), "Parse error")
			})
		}
	}

	if containsString(stringTypes, "''-bs") {
		t.Run("supports escaping single-quote with backslash", func(t *testing.T) {
			assertEqual(t, format(`'foo \' JOIN bar'`), `'foo \' JOIN bar'`)
		})

		if !containsString(stringTypes, "''-qq") {
			t.Run("does not support escaping single-quote by doubling it", func(t *testing.T) {
				assertEqual(t, format(`'foo '' JOIN bar'`), `'foo ' ' JOIN bar'`)
			})
		}
	}

	if containsString(stringTypes, "U&''") {
		t.Run("supports unicode single-quoted strings", func(t *testing.T) {
			assertEqual(t, format(`U&'foo JOIN bar'`), `U&'foo JOIN bar'`)
			expected := dedent(`
				SELECT
				  U&'where'
				FROM
				  U&'update'
			`)
			assertEqual(t, format(`SELECT U&'where' FROM U&'update'`), expected)
		})

		t.Run("supports escaping in U&'' strings with repeated quote", func(t *testing.T) {
			assertEqual(t, format(`U&'foo '' JOIN bar'`), `U&'foo '' JOIN bar'`)
		})

		t.Run("detects consecutive U&'' strings as separate ones", func(t *testing.T) {
			assertEqual(t, format(`U&'foo'U&'bar'`), `U&'foo' U&'bar'`)
		})
	}

	if containsString(stringTypes, "N''") {
		t.Run("supports T-SQL unicode strings", func(t *testing.T) {
			assertEqual(t, format(`N'foo JOIN bar'`), `N'foo JOIN bar'`)
			expected := dedent(`
				SELECT
				  N'where'
				FROM
				  N'update'
			`)
			assertEqual(t, format(`SELECT N'where' FROM N'update'`), expected)
		})

		if containsString(stringTypes, "''-qq") {
			t.Run("supports escaping in N'' strings with repeated quote", func(t *testing.T) {
				assertEqual(t, format(`N'foo '' JOIN bar'`), `N'foo '' JOIN bar'`)
			})
		}
		if containsString(stringTypes, "''-bs") {
			t.Run("supports escaping in N'' strings with backslash", func(t *testing.T) {
				assertEqual(t, format(`N'foo \' JOIN bar'`), `N'foo \' JOIN bar'`)
			})
		}

		t.Run("detects consecutive N'' strings as separate ones", func(t *testing.T) {
			assertEqual(t, format(`N'foo'N'bar'`), `N'foo' N'bar'`)
		})
	}

	if containsString(stringTypes, "X''") {
		t.Run("supports hex byte sequences", func(t *testing.T) {
			assertEqual(t, format(`x'0E'`), `x'0E'`)
			assertEqual(t, format(`X'1F0A89C3'`), `X'1F0A89C3'`)
			expected := dedent(`
				SELECT
				  x'2B'
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT x'2B' FROM foo`), expected)
		})

		t.Run("detects consecutive X'' strings as separate ones", func(t *testing.T) {
			assertEqual(t, format(`X'AE01'X'01F6'`), `X'AE01' X'01F6'`)
		})
	}

	if containsString(stringTypes, `X""`) {
		t.Run(`supports hex byte sequences with double-quotes`, func(t *testing.T) {
			assertEqual(t, format(`x"0E"`), `x"0E"`)
			assertEqual(t, format(`X"1F0A89C3"`), `X"1F0A89C3"`)
			expected := dedent(`
				SELECT
				  x"2B"
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT x"2B" FROM foo`), expected)
		})

		t.Run(`detects consecutive X"" strings as separate ones`, func(t *testing.T) {
			assertEqual(t, format(`X"AE01"X"01F6"`), `X"AE01" X"01F6"`)
		})
	}

	if containsString(stringTypes, "B''") {
		t.Run("supports bit sequences", func(t *testing.T) {
			assertEqual(t, format(`b'01'`), `b'01'`)
			assertEqual(t, format(`B'10110'`), `B'10110'`)
			expected := dedent(`
				SELECT
				  b'0101'
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT b'0101' FROM foo`), expected)
		})

		t.Run("detects consecutive B'' strings as separate ones", func(t *testing.T) {
			assertEqual(t, format(`B'1001'B'0110'`), `B'1001' B'0110'`)
		})
	}

	if containsString(stringTypes, `B""`) {
		t.Run("supports bit sequences with double-quotes", func(t *testing.T) {
			assertEqual(t, format(`b"01"`), `b"01"`)
			assertEqual(t, format(`B"10110"`), `B"10110"`)
			expected := dedent(`
				SELECT
				  b"0101"
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT b"0101" FROM foo`), expected)
		})

		t.Run(`detects consecutive B"" strings as separate ones`, func(t *testing.T) {
			assertEqual(t, format(`B"1001"B"0110"`), `B"1001" B"0110"`)
		})
	}

	if containsString(stringTypes, "R''") {
		t.Run("supports no escaping in raw strings", func(t *testing.T) {
			expected := dedent(`
				SELECT
				  r'some \\',
				  R'text'
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT r'some \\',R'text' FROM foo`), expected)
		})

		t.Run("detects consecutive r'' strings as separate ones", func(t *testing.T) {
			assertEqual(t, format(`r'a ha'r'hm mm'`), `r'a ha' r'hm mm'`)
		})
	}

	if containsString(stringTypes, `R""`) {
		t.Run("supports no escaping in raw strings with double-quotes", func(t *testing.T) {
			expected := dedent(`
				SELECT
				  r"some \\",
				  R"text"
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT r"some \\", R"text" FROM foo`), expected)
		})

		t.Run(`detects consecutive r"" strings as separate ones`, func(t *testing.T) {
			assertEqual(t, format(`r"a ha"r"hm mm"`), `r"a ha" r"hm mm"`)
		})
	}

	if containsString(stringTypes, "E''") {
		t.Run("supports E'' strings with C-style escapes", func(t *testing.T) {
			assertEqual(t, format(`E'blah blah'`), `E'blah blah'`)
			assertEqual(t, format(`E'some \' FROM escapes'`), `E'some \' FROM escapes'`)
			expected := dedent(`
				SELECT
				  E'blah'
				FROM
				  foo
			`)
			assertEqual(t, format(`SELECT E'blah' FROM foo`), expected)
			assertEqual(t, format(`E'blah''blah'`), `E'blah''blah'`)
		})

		t.Run(`detects consecutive E'' strings as separate ones`, func(t *testing.T) {
			assertEqual(t, format(`e'a ha'e'hm mm'`), `e'a ha' e'hm mm'`)
		})
	}

	if containsString(stringTypes, "$$") {
		t.Run("supports dollar-quoted strings", func(t *testing.T) {
			assertEqual(t, format(`$$foo JOIN bar$$`), `$$foo JOIN bar$$`)
			assertEqual(t, format(`$$foo $ JOIN bar$$`), `$$foo $ JOIN bar$$`)
			assertEqual(t, format("$$foo \n bar$$"), "$$foo \n bar$$")
			expected := dedent(`
				SELECT
				  $$where$$
				FROM
				  $$update$$
			`)
			assertEqual(t, format(`SELECT $$where$$ FROM $$update$$`), expected)
		})

		t.Run("supports tagged dollar-quoted strings", func(t *testing.T) {
			assertEqual(t, format(`$xxx$foo $$ LEFT JOIN $yyy$ bar$xxx$`), `$xxx$foo $$ LEFT JOIN $yyy$ bar$xxx$`)
		})
	}
}
