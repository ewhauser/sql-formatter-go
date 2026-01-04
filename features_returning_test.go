package sqlformatter

import "testing"

func supportsReturning(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("places RETURNING to new line", func(t *testing.T) {
		result := format("INSERT INTO users (firstname, lastname) VALUES ('Joe', 'Cool') RETURNING id, firstname;")
		expected := dedent(`
			INSERT INTO
			  users (firstname, lastname)
			VALUES
			  ('Joe', 'Cool')
			RETURNING
			  id,
			  firstname;
		`)
		assertEqual(t, result, expected)
	})
}
