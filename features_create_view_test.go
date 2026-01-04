package sqlformatter

import "testing"

type createViewConfig struct {
	OrReplace   bool
	Materialized bool
	IfNotExists bool
}

func supportsCreateView(t *testing.T, format FormatFn, cfg createViewConfig) {
	t.Helper()
	t.Run("formats CREATE VIEW", func(t *testing.T) {
		result := format("CREATE VIEW my_view AS SELECT id, fname, lname FROM tbl;")
		expected := dedent(`
			CREATE VIEW my_view AS
			SELECT
			  id,
			  fname,
			  lname
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CREATE VIEW with columns", func(t *testing.T) {
		result := format("CREATE VIEW my_view (id, fname, lname) AS SELECT * FROM tbl;")
		expected := dedent(`
			CREATE VIEW my_view (id, fname, lname) AS
			SELECT
			  *
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	if cfg.OrReplace {
		t.Run("formats CREATE OR REPLACE VIEW", func(t *testing.T) {
			result := format("CREATE OR REPLACE VIEW v1 AS SELECT 42;")
			expected := dedent(`
				CREATE OR REPLACE VIEW v1 AS
				SELECT
				  42;
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.Materialized {
		t.Run("formats CREATE MATERIALIZED VIEW", func(t *testing.T) {
			result := format("CREATE MATERIALIZED VIEW mat_view AS SELECT 42;")
			expected := dedent(`
				CREATE MATERIALIZED VIEW mat_view AS
				SELECT
				  42;
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.IfNotExists {
		t.Run("formats CREATE VIEW IF NOT EXISTS", func(t *testing.T) {
			result := format("CREATE VIEW IF NOT EXISTS my_view AS SELECT 42;")
			expected := dedent(`
				CREATE VIEW IF NOT EXISTS my_view AS
				SELECT
				  42;
			`)
			assertEqual(t, result, expected)
		})
	}
}
