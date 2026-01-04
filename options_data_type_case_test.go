package sqlformatter

import "testing"

func supportsDataTypeCase(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("preserves data type keyword case by default", func(t *testing.T) {
		result := format("CREATE TABLE users ( user_id iNt PRIMARY KEY, total_earnings Decimal(5, 2) NOT NULL )")
		expected := dedent(`
			CREATE TABLE users (
			  user_id iNt PRIMARY KEY,
			  total_earnings Decimal(5, 2) NOT NULL
			)
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts data type keyword case to uppercase", func(t *testing.T) {
		result := format("CREATE TABLE users ( user_id iNt PRIMARY KEY, total_earnings Decimal(5, 2) NOT NULL )", FormatOptions{DataTypeCase: KeywordCaseUpper})
		expected := dedent(`
			CREATE TABLE users (
			  user_id INT PRIMARY KEY,
			  total_earnings DECIMAL(5, 2) NOT NULL
			)
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts data type keyword case to lowercase", func(t *testing.T) {
		result := format("CREATE TABLE users ( user_id iNt PRIMARY KEY, total_earnings Decimal(5, 2) NOT NULL )", FormatOptions{DataTypeCase: KeywordCaseLower})
		expected := dedent(`
			CREATE TABLE users (
			  user_id int PRIMARY KEY,
			  total_earnings decimal(5, 2) NOT NULL
			)
		`)
		assertEqual(t, result, expected)
	})
}
