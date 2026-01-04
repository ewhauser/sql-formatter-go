package sqlformatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func supportsExpressionWidth(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("throws error when expressionWidth negative", func(t *testing.T) {
		err := formatPostgresErr(t, "SELECT *", FormatOptions{ExpressionWidth: -2, ExpressionWidthSet: true})
		require.Error(t, err)
		require.Equal(t, "expressionWidth config must be positive number. Received -2 instead.", err.Error())
	})

	t.Run("throws error when expressionWidth is zero", func(t *testing.T) {
		err := formatPostgresErr(t, "SELECT *", FormatOptions{ExpressionWidth: 0, ExpressionWidthSet: true})
		require.Error(t, err)
		require.Equal(t, "expressionWidth config must be positive number. Received 0 instead.", err.Error())
	})

	t.Run("breaks parenthesized expressions when exceed width", func(t *testing.T) {
		result := format("SELECT product.price + (product.original_price * product.sales_tax) AS total FROM product;", FormatOptions{ExpressionWidth: 40, ExpressionWidthSet: true})
		expected := dedent(`
			SELECT
			  product.price + (
			    product.original_price * product.sales_tax
			  ) AS total
			FROM
			  product;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("keeps parenthesized expressions on single line", func(t *testing.T) {
		result := format("SELECT product.price + (product.original_price * product.sales_tax) AS total FROM product;", FormatOptions{ExpressionWidth: 50, ExpressionWidthSet: true})
		expected := dedent(`
			SELECT
			  product.price + (product.original_price * product.sales_tax) AS total
			FROM
			  product;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("calculates parenthesized expression length with spaces", func(t *testing.T) {
		result := format("SELECT (price * tax) AS total FROM table_name WHERE (amount > 25);", FormatOptions{ExpressionWidth: 10, ExpressionWidthSet: true, DenseOperators: true})
		expected := dedent(`
    SELECT
      (price*tax) AS total
    FROM
      table_name
    WHERE
      (amount>25);
    `)
		assertEqual(t, result, expected)
	})

	t.Run("formats inline when params shorter than width", func(t *testing.T) {
		result := format("SELECT (?, ?, ?) AS total;", FormatOptions{ExpressionWidth: 11, ExpressionWidthSet: true, ParamTypes: &ParamTypes{Positional: true}, Params: []string{"10", "20", "30"}})
		expected := dedent(`
			SELECT
			  (10, 20, 30) AS total;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats not inline when params longer than width", func(t *testing.T) {
		result := format("SELECT (?, ?, ?) AS total;", FormatOptions{ExpressionWidth: 11, ExpressionWidthSet: true, ParamTypes: &ParamTypes{Positional: true}, Params: []string{"100", "200", "300"}})
		expected := dedent(`
			SELECT
			  (
			    100,
			    200,
			    300
			  ) AS total;
		`)
		assertEqual(t, result, expected)
	})
}
