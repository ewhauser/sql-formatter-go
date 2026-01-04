package sqlformatter

import "testing"

type alterTableConfig struct {
	AddColumn   bool
	DropColumn  bool
	Modify      bool
	RenameTo    bool
	RenameColumn bool
}

func supportsAlterTable(t *testing.T, format FormatFn, cfg alterTableConfig) {
	t.Helper()
	if cfg.AddColumn {
		t.Run("formats ALTER TABLE ADD COLUMN", func(t *testing.T) {
			result := format("ALTER TABLE supplier ADD COLUMN unit_price DECIMAL NOT NULL;")
			expected := dedent(`
				ALTER TABLE supplier
				ADD COLUMN unit_price DECIMAL NOT NULL;
			`)
			assertEqual(t, result, expected)
		})
	}
	if cfg.DropColumn {
		t.Run("formats ALTER TABLE DROP COLUMN", func(t *testing.T) {
			result := format("ALTER TABLE supplier DROP COLUMN unit_price;")
			expected := dedent(`
				ALTER TABLE supplier
				DROP COLUMN unit_price;
			`)
			assertEqual(t, result, expected)
		})
	}
	if cfg.Modify {
		t.Run("formats ALTER TABLE MODIFY", func(t *testing.T) {
			result := format("ALTER TABLE supplier MODIFY supplier_id DECIMAL NULL;")
			expected := dedent(`
				ALTER TABLE supplier
				MODIFY supplier_id DECIMAL NULL;
			`)
			assertEqual(t, result, expected)
		})
	}
	if cfg.RenameTo {
		t.Run("formats ALTER TABLE RENAME TO", func(t *testing.T) {
			result := format("ALTER TABLE supplier RENAME TO the_one_who_supplies;")
			expected := dedent(`
				ALTER TABLE supplier
				RENAME TO the_one_who_supplies;
			`)
			assertEqual(t, result, expected)
		})
	}
	if cfg.RenameColumn {
		t.Run("formats ALTER TABLE RENAME COLUMN", func(t *testing.T) {
			result := format("ALTER TABLE supplier RENAME COLUMN supplier_id TO id;")
			expected := dedent(`
				ALTER TABLE supplier
				RENAME COLUMN supplier_id TO id;
			`)
			assertEqual(t, result, expected)
		})
	}
}
