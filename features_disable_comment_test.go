package sqlformatter

import "testing"

func supportsDisableComment(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("does not format text between disable/enable comments", func(t *testing.T) {
		result := format(dedent(`
      SELECT foo FROM bar;
      /* sql-formatter-disable */
      SELECT foo FROM bar;
      /* sql-formatter-enable */
      SELECT foo FROM bar;
    `))
		expected := dedent(`
			SELECT
			  foo
			FROM
			  bar;

			/* sql-formatter-disable */
			SELECT foo FROM bar;
			/* sql-formatter-enable */
			SELECT
			  foo
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("preserves indentation between disable/enable comments", func(t *testing.T) {
		result := format(dedent(`
      /* sql-formatter-disable */
      SELECT
        foo
          FROM
            bar;
      /* sql-formatter-enable */
    `))
		expected := dedent(`
			/* sql-formatter-disable */
			SELECT
			  foo
			    FROM
			      bar;
			/* sql-formatter-enable */
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not format text after disable until end", func(t *testing.T) {
		result := format(dedent(`
      SELECT foo FROM bar;
      /* sql-formatter-disable */
      SELECT foo FROM bar;

      SELECT foo FROM bar;
    `))
		expected := dedent(`
			SELECT
			  foo
			FROM
			  bar;

			/* sql-formatter-disable */
			SELECT foo FROM bar;

			SELECT foo FROM bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not parse code between disable/enable comments", func(t *testing.T) {
		result := format(dedent(`
      SELECT /*sql-formatter-disable*/ ?!{}[] /*sql-formatter-enable*/ FROM bar;
    `))
		expected := dedent(`
			SELECT
			  /*sql-formatter-disable*/ ?!{}[] /*sql-formatter-enable*/
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})
}
