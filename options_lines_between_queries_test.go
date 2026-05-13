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

	t.Run("does not add empty line before line-comment statement markers", func(t *testing.T) {
		result := format("-- migrate:up\nSELECT 1;\n-- migrate:down\nSELECT 2;")
		expected := dedent(`
			-- migrate:up
			SELECT
			  1;
			-- migrate:down
			SELECT
			  2;
		`)
		assertEqual(t, result, expected)
	})
}
