package sqlformatter

import "testing"

type arrayLiteralConfig struct {
	WithArrayPrefix   bool
	WithoutArrayPrefix bool
}

func supportsArrayLiterals(t *testing.T, format FormatFn, cfg arrayLiteralConfig) {
	t.Helper()
	if cfg.WithArrayPrefix {
		t.Run("supports ARRAY[] literals", func(t *testing.T) {
			result := format("SELECT ARRAY[1, 2, 3] FROM ARRAY['come-on', 'seriously', 'this', 'is', 'a', 'very', 'very', 'long', 'array'];")
			expected := dedent(`
				SELECT
				  ARRAY[1, 2, 3]
				FROM
				  ARRAY[
				    'come-on',
				    'seriously',
				    'this',
				    'is',
				    'a',
				    'very',
				    'very',
				    'long',
				    'array'
				  ];
			`)
			assertEqual(t, result, expected)
		})

		t.Run("dataTypeCase does not affect ARRAY literal case", func(t *testing.T) {
			result := format("SELECT ArrAy[1, 2]", FormatOptions{DataTypeCase: KeywordCaseUpper})
			expected := dedent(`
				SELECT
				  ArrAy[1, 2]
			`)
			assertEqual(t, result, expected)
		})

		t.Run("keywordCase affects ARRAY literal case", func(t *testing.T) {
			result := format("SELECT ArrAy[1, 2]", FormatOptions{KeywordCase: KeywordCaseUpper})
			expected := dedent(`
				SELECT
				  ARRAY[1, 2]
			`)
			assertEqual(t, result, expected)
		})

		t.Run("dataTypeCase affects ARRAY type case", func(t *testing.T) {
			result := format("CREATE TABLE foo ( items ArrAy )", FormatOptions{DataTypeCase: KeywordCaseUpper})
			expected := dedent(`
				CREATE TABLE foo (items ARRAY)
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.WithoutArrayPrefix {
		t.Run("supports array literals", func(t *testing.T) {
			result := format("SELECT [1, 2, 3] FROM ['come-on', 'seriously', 'this', 'is', 'a', 'very', 'very', 'long', 'array'];")
			expected := dedent(`
				SELECT
				  [1, 2, 3]
				FROM
				  [
				    'come-on',
				    'seriously',
				    'this',
				    'is',
				    'a',
				    'very',
				    'very',
				    'long',
				    'array'
				  ];
			`)
			assertEqual(t, result, expected)
		})
	}
}
