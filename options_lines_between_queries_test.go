package sqlformatter

import "testing"

func supportsLinesBetweenQueries(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("defaults to single empty line between queries", func(t *testing.T) {
		result := format("SELECT * FROM foo; SELECT * FROM bar;")
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo;

			SELECT
			  *
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports more empty lines between queries", func(t *testing.T) {
		result := format("SELECT * FROM foo; SELECT * FROM bar;", FormatOptions{LinesBetweenQueries: 2, LinesBetweenQueriesSet: true})
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo;


			SELECT
			  *
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports no empty lines between queries", func(t *testing.T) {
		result := format("SELECT * FROM foo; SELECT * FROM bar;", FormatOptions{LinesBetweenQueries: 0, LinesBetweenQueriesSet: true})
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo;
			SELECT
			  *
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})
}
