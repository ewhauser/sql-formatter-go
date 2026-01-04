package sqlformatter

import "testing"

func supportsConstraints(t *testing.T, format FormatFn, actions []string) {
	t.Helper()
	for _, action := range actions {
		action := action
		t.Run("treats ON UPDATE/DELETE "+action+" as distinct keywords", func(t *testing.T) {
			result := format(dedent(`
        CREATE TABLE foo (
          update_time datetime ON UPDATE ` + action + `,
          delete_time datetime ON DELETE ` + action + `,
        );
      `))
			expected := dedent(`
				CREATE TABLE foo (
				  update_time datetime ON UPDATE ` + action + `,
				  delete_time datetime ON DELETE ` + action + `,
				);
			`)
			assertEqual(t, result, expected)
		})
	}
}
