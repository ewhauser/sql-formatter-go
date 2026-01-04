package sqlformatter

import "testing"

type numbersConfig struct {
	Underscore bool
}

func supportsNumbers(t *testing.T, format FormatFn, cfg numbersConfig) {
	t.Helper()
	t.Run("supports decimal numbers", func(t *testing.T) {
		result := format("SELECT 42, -35.04, 105., 2.53E+3, 1.085E-5;")
		expected := dedent(`
			SELECT
			  42,
			  -35.04,
			  105.,
			  2.53E+3,
			  1.085E-5;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports hex and binary numbers", func(t *testing.T) {
		result := format("SELECT 0xAE, 0x10F, 0b1010001;")
		expected := dedent(`
			SELECT
			  0xAE,
			  0x10F,
			  0b1010001;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("handles floats as single tokens", func(t *testing.T) {
		result := format("SELECT 1e-9 AS a, 1.5e+10 AS b, 3.5E12 AS c, 3.5e12 AS d;")
		expected := dedent(`
			SELECT
			  1e-9 AS a,
			  1.5e+10 AS b,
			  3.5E12 AS c,
			  3.5e12 AS d;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("handles floats with trailing point", func(t *testing.T) {
		result := format("SELECT 1000. AS a;")
		expected := dedent(`
			SELECT
			  1000. AS a;
		`)
		assertEqual(t, result, expected)

		result = format("SELECT a, b / 1000. AS a_s, 100. * b / SUM(a_s);")
		expected = dedent(`
			SELECT
			  a,
			  b / 1000. AS a_s,
			  100. * b / SUM(a_s);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports decimal values without leading digits", func(t *testing.T) {
		result := format("SELECT .456 AS foo;")
		expected := dedent(`
			SELECT
			  .456 AS foo;
		`)
		assertEqual(t, result, expected)
	})

	if cfg.Underscore {
		t.Run("supports underscore separators in numeric literals", func(t *testing.T) {
			result := format("SELECT 1_000_000, 3.14_159, 0x1A_2B_3C, 0b1010_0001, 1.5e+1_0;")
			expected := dedent(`
				SELECT
				  1_000_000,
				  3.14_159,
				  0x1A_2B_3C,
				  0b1010_0001,
				  1.5e+1_0;
			`)
			assertEqual(t, result, expected)
		})
	}
}
