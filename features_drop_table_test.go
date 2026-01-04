package sqlformatter

import "testing"

type dropTableConfig struct {
	IfExists bool
}

func supportsDropTable(t *testing.T, format FormatFn, cfg dropTableConfig) {
	t.Helper()
	t.Run("formats DROP TABLE statement", func(t *testing.T) {
		result := format("DROP TABLE admin_role;")
		expected := dedent(`
			DROP TABLE admin_role;
		`)
		assertEqual(t, result, expected)
	})

	if cfg.IfExists {
		t.Run("formats DROP TABLE IF EXISTS statement", func(t *testing.T) {
			result := format("DROP TABLE IF EXISTS admin_role;")
			expected := dedent(`
				DROP TABLE IF EXISTS admin_role;
			`)
			assertEqual(t, result, expected)
		})
	}
}
