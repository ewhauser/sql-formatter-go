package sqlformatter

import "testing"

type operatorConfig struct {
	LogicalOperators []string
	Any              bool
}

func supportsOperators(t *testing.T, format FormatFn, operators []string, cfg operatorConfig) {
	t.Helper()
	standardOperators := []string{"+", "-", "*", "/", ">", "<", "=", "<>", "<=", ">=", "!="}
	ops := append(append([]string{}, standardOperators...), operators...)

	for _, op := range ops {
		op := op
		t.Run("supports "+op+" operator", func(t *testing.T) {
			result := format("foo" + op + " bar " + op + "zap")
			expected := "foo " + op + " bar " + op + " zap"
			assertEqual(t, result, expected)
		})
	}

	for _, op := range ops {
		op := op
		t.Run("supports "+op+" operator in dense mode", func(t *testing.T) {
			result := format("foo "+op+" bar", FormatOptions{DenseOperators: true})
			expected := "foo" + op + "bar"
			assertEqual(t, result, expected)
		})
	}

	logicalOps := cfg.LogicalOperators
	if len(logicalOps) == 0 {
		logicalOps = []string{"AND", "OR"}
	}
	for _, op := range logicalOps {
		op := op
		t.Run("supports "+op+" operator", func(t *testing.T) {
			result := format("SELECT true " + op + " false AS foo;")
			expected := dedent(`
				SELECT
				  true
				  ` + op + ` false AS foo;
			`)
			assertEqual(t, result, expected)
		})
	}

	t.Run("supports set operators", func(t *testing.T) {
		assertEqual(t, format("foo ALL bar"), "foo ALL bar")
		assertEqual(t, format("EXISTS bar"), "EXISTS bar")
		assertEqual(t, format("foo IN (1, 2, 3)"), "foo IN (1, 2, 3)")
		assertEqual(t, format("foo LIKE 'hello%'"), "foo LIKE 'hello%'")
		assertEqual(t, format("foo IS NULL"), "foo IS NULL")
		assertEqual(t, format("UNIQUE foo"), "UNIQUE foo")
	})

	if cfg.Any {
		t.Run("supports ANY set-operator", func(t *testing.T) {
			assertEqual(t, format("foo = ANY (1, 2, 3)"), "foo = ANY (1, 2, 3)")
		})
	}
}
