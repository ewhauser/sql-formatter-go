package sqlformatter

import "testing"

type updateConfig struct {
	WhereCurrentOf bool
}

func supportsUpdate(t *testing.T, format FormatFn, cfg updateConfig) {
	t.Helper()
	t.Run("formats simple UPDATE statement", func(t *testing.T) {
		result := format("UPDATE Customers SET ContactName='Alfred Schmidt', City='Hamburg' WHERE CustomerName='Alfreds Futterkiste';")
		expected := dedent(`
			UPDATE Customers
			SET
			  ContactName = 'Alfred Schmidt',
			  City = 'Hamburg'
			WHERE
			  CustomerName = 'Alfreds Futterkiste';
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats UPDATE statement with AS part", func(t *testing.T) {
		result := format("UPDATE customers SET total_orders = order_summary.total  FROM ( SELECT * FROM bank) AS order_summary")
		expected := dedent(`
			UPDATE customers
			SET
			  total_orders = order_summary.total
			FROM
			  (
			    SELECT
			      *
			    FROM
			      bank
			  ) AS order_summary
		`)
		assertEqual(t, result, expected)
	})

	if cfg.WhereCurrentOf {
		t.Run("formats UPDATE statement with cursor position", func(t *testing.T) {
			result := format("UPDATE Customers SET Name='John' WHERE CURRENT OF my_cursor;")
			expected := dedent(`
				UPDATE Customers
				SET
				  Name = 'John'
				WHERE CURRENT OF my_cursor;
			`)
			assertEqual(t, result, expected)
		})
	}
}
