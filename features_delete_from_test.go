package sqlformatter

import "testing"

type deleteFromConfig struct {
	WithoutFrom bool
}

func supportsDeleteFrom(t *testing.T, format FormatFn, cfgs ...deleteFromConfig) {
	t.Helper()
	cfg := deleteFromConfig{}
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	t.Run("formats DELETE FROM statement", func(t *testing.T) {
		result := format("DELETE FROM Customers WHERE CustomerName='Alfred' AND Phone=5002132;")
		expected := dedent(`
			DELETE FROM Customers
			WHERE
			  CustomerName = 'Alfred'
			  AND Phone = 5002132;
		`)
		assertEqual(t, result, expected)
	})

	if cfg.WithoutFrom {
		t.Run("formats DELETE statement without FROM", func(t *testing.T) {
			result := format("DELETE Customers WHERE CustomerName='Alfred';")
			expected := dedent(`
				DELETE Customers
				WHERE
				  CustomerName = 'Alfred';
			`)
			assertEqual(t, result, expected)
		})
	}
}
