package sqlformatter

import "testing"

func supportsArrayAndMapAccessors(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("supports square brackets for array indexing", func(t *testing.T) {
		result := format("SELECT arr[1], order_lines[5].productId;")
		expected := dedent(`
			SELECT
			  arr[1],
			  order_lines[5].productId;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports square brackets for map lookup", func(t *testing.T) {
		result := format("SELECT alpha['a'], beta['gamma'].zeta, yota['foo.bar-baz'];")
		expected := dedent(`
			SELECT
			  alpha['a'],
			  beta['gamma'].zeta,
			  yota['foo.bar-baz'];
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports square brackets for map lookup - uppercase", func(t *testing.T) {
		result := format("SELECT Alpha['a'], Beta['gamma'].zeTa, yotA['foo.bar-baz'];", FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  ALPHA['a'],
			  BETA['gamma'].ZETA,
			  YOTA['foo.bar-baz'];
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports namespaced array identifiers", func(t *testing.T) {
		result := format("SELECT foo.coalesce['blah'];")
		expected := dedent(`
			SELECT
			  foo.coalesce['blah'];
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats array accessor with comment in-between", func(t *testing.T) {
		result := format("SELECT arr /* comment */ [1];")
		expected := dedent(`
			SELECT
			  arr/* comment */ [1];
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats namespaced array accessor with comment in-between", func(t *testing.T) {
		result := format("SELECT foo./* comment */arr[1];")
		expected := dedent(`
			SELECT
			  foo./* comment */ arr[1];
		`)
		assertEqual(t, result, expected)
	})

	t.Run("changes case of array accessors when identifierCase used", func(t *testing.T) {
		result := format("SELECT arr[1];", FormatOptions{IdentifierCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  ARR[1];
		`)
		assertEqual(t, result, expected)

		result = format("SELECT NS.Arr[1];", FormatOptions{IdentifierCase: KeywordCaseLower})
		expected = dedent(`
			SELECT
			  ns.arr[1];
		`)
		assertEqual(t, result, expected)
	})
}
