package sqlformatter

import "testing"

type insertIntoConfig struct {
	WithoutInto bool
}

func supportsInsertInto(t *testing.T, format FormatFn, cfgs ...insertIntoConfig) {
	t.Helper()
	cfg := insertIntoConfig{}
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	t.Run("formats simple INSERT INTO", func(t *testing.T) {
		result := format("INSERT INTO Customers (ID, MoneyBalance, Address, City) VALUES (12,-123.4, 'Skagen 2111','Stv');")
		expected := dedent(`
			INSERT INTO
			  Customers (ID, MoneyBalance, Address, City)
			VALUES
			  (12, -123.4, 'Skagen 2111', 'Stv');
		`)
		assertEqual(t, result, expected)
	})

	if cfg.WithoutInto {
		t.Run("formats INSERT without INTO", func(t *testing.T) {
			result := format("INSERT Customers (ID, MoneyBalance, Address, City) VALUES (12,-123.4, 'Skagen 2111','Stv');")
			expected := dedent(`
				INSERT
				  Customers (ID, MoneyBalance, Address, City)
				VALUES
				  (12, -123.4, 'Skagen 2111', 'Stv');
			`)
			assertEqual(t, result, expected)
		})
	}
}
