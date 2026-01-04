package sqlformatter

import "testing"

func supportsFunctionCase(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("preserves function name case by default", func(t *testing.T) {
		result := format("SELECT MiN(price) AS min_price, Cast(item_code AS INT) FROM products")
		expected := dedent(`
			SELECT
			  MiN(price) AS min_price,
			  Cast(item_code AS INT)
			FROM
			  products
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts function names to uppercase", func(t *testing.T) {
		result := format("SELECT MiN(price) AS min_price, Cast(item_code AS INT) FROM products", FormatOptions{FunctionCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  MIN(price) AS min_price,
			  CAST(item_code AS INT)
			FROM
			  products
		`)
		assertEqual(t, result, expected)
	})

	t.Run("converts function names to lowercase", func(t *testing.T) {
		result := format("SELECT MiN(price) AS min_price, Cast(item_code AS INT) FROM products", FormatOptions{FunctionCase: KeywordCaseLower})
		expected := dedent(`
			SELECT
			  min(price) AS min_price,
			  cast(item_code AS INT)
			FROM
			  products
		`)
		assertEqual(t, result, expected)
	})
}
